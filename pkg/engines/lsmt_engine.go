package engines

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type (
	lsmtSegmentFactory func(opts lsmtSegmentConfig) lsmtSegment

	lsmtOptions struct {
		Path           string
		SegmentFactory lsmtSegmentFactory
	}
	lsmtEngine struct {
		compactInterval time.Duration
		path            string // base path

		segmentFactory lsmtSegmentFactory
		segments       []lsmtSegment
		segmentsL      sync.RWMutex

		lastSegment lsmtSegment
	}
)

func (e *lsmtEngine) Search(key string) (string, error) {
	if r, err := e.lastSegment.Search(key); err == nil {
		return r, nil
	}

	limit := len(e.segments) - 1

	for limit >= 0 {
		if r, err := e.segments[limit].Search(key); err == nil {
			return r, nil
		}
		limit--
	}

	return "", errors.New("not found") // TODO: bloom filter
}

func (e *lsmtEngine) Insert(key, value string) error {
	return e.lastSegment.Insert(key, value)
}

func (e *lsmtEngine) compact() {

}

func (e *lsmtEngine) add(s lsmtSegment) {
	e.segmentsL.Lock()

	e.segments = append(e.segments, s)

	if e.lastSegment == nil {
		e.lastSegment = s
	} else if s.Compare(e.lastSegment) == 1 {
		e.lastSegment = s
	}
	e.segmentsL.Unlock()
	log.Println("loaded segment", s)
	log.Println("most recent segment", e.lastSegment)

}

func (e *lsmtEngine) scan() error {
	lastDir := filepath.Base(e.path) // the last directory in the hierarchy. E.g /var/lib/data -> data. It is needed
	// because WalkDir also send it to WalkFunc
	err := filepath.WalkDir(e.path, func(path string, d fs.DirEntry, err error) error {
		log.Println(path, err)
		if err != nil {
			log.Println("occurred while scanning up lsmt dir", err)
			return err
		}

		if d != nil && d.IsDir() {
			log.Println(path, d.Name(), d.Type(), d.IsDir(), err)
			if strings.EqualFold(lastDir, d.Name()) {
				return nil
			}
			return filepath.SkipDir
		}
		name := filepath.Base(path)
		if !strings.Contains(name, "segment-") {
			return nil
		}

		e.add(e.segmentFactory(lsmtSegmentOpts{
			SegmentPath: path,
			SizeLimit:   4 * 1024 * 1024,
		}))

		return nil
	})

	if err != nil {
		log.Println("cloud not scan", err)
		return err
	}

	if e.lastSegment == nil {
		log.Println("there was no first segment, creating one")
		e.add(e.segmentFactory(lsmtSegmentOpts{
			SegmentPath: newSegmentName(e.path),
			SizeLimit:   4 * 1024 * 1024,
		}))
	}

	return nil
}

func (e *lsmtEngine) close() {
	wg := sync.WaitGroup{}
	wg.Add(len(e.segments))
	for _, seg := range e.segments {
		seg := seg
		go func() {
			log.Println("closing segment", seg)
			if err := seg.Close(); err != nil {
				log.Println("error closing segment", seg)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func (e *lsmtEngine) open() {
	wg := sync.WaitGroup{}
	wg.Add(len(e.segments))
	for _, seg := range e.segments {
		seg := seg
		go func() {
			log.Println("opening segment", seg)
			if err := seg.Open(); err != nil {
				log.Println("error opening segment", seg, err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func (e *lsmtEngine) run(ctx context.Context) {
	if err := e.scan(); err != nil {
		log.Fatalln(err)
		return
	}
	defer e.close()
	e.open()
	ticker := time.NewTicker(e.compactInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("context canceled due to", ctx.Err())
			return
		case <-ticker.C:
			e.compact()
		}
	}
}

func NewLSMTEngine(opts lsmtOptions) *lsmtEngine {
	if strings.EqualFold(opts.Path, "") {
		opts.Path = "/var/lib/visigoth/data/"
	}
	return &lsmtEngine{
		compactInterval: 10 * time.Second,
		path:            opts.Path,
		segments:        nil,
		segmentFactory:  opts.SegmentFactory,
		segmentsL:       sync.RWMutex{},
	}
}

func NewLSMTEngineBuilder(opts lsmtOptions) func(ctx context.Context) *lsmtEngine {
	return func(ctx context.Context) *lsmtEngine {
		engi := NewLSMTEngine(opts)
		go engi.run(ctx)
		return engi
	}
}
