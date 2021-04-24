package engines

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type (
	basicLSMTSegment struct {
		path string
		wfd  *os.File
		rfd  *os.File

		limit int

		pairs []lsmtPair
	}

	lsmtPair struct {
		offsetStart int
		offsetEnd   int

		key   string
		value string
	}
)

func (s *lsmtPair) String() string {
	return fmt.Sprintf("pair{key=%s,value=%s}",
		s.key, s.value)
}

func (s *lsmtPair) Encode(w io.Writer) (int, error) {
	kl := make([]byte, 2)
	vl := make([]byte, 4)
	binary.LittleEndian.PutUint16(kl, uint16(len(s.key)))
	binary.LittleEndian.PutUint32(vl, uint32(len(s.value)))
	n1, e := w.Write(kl)
	n2, e := w.Write([]byte(s.key))
	n3, e := w.Write(vl)
	n4, e := w.Write([]byte(s.value))
	return n1 + n2 + n3 + n4, e
}

func (s *lsmtPair) Decode(r io.Reader) (n int, err error) {
	klb := make([]byte, 2)
	_, err = io.ReadAtLeast(r, klb, 2)
	if err != nil {
		return 2, err
	}
	kl := binary.LittleEndian.Uint16(klb)
	kb := make([]byte, kl)
	_, err = io.ReadAtLeast(r, kb, int(kl))

	if err != nil {
		return 0, err
	}
	s.key = string(kb)

	vlb := make([]byte, 4)
	_, err = io.ReadAtLeast(r, vlb, 4)

	if err != nil {
		return 0, err
	}
	vl := binary.LittleEndian.Uint32(vlb)
	vb := make([]byte, vl)
	_, err = io.ReadAtLeast(r, vb, int(vl))

	if err != nil {
		return 0, err
	}
	s.value = string(vb)
	return 6 + len(s.key) + len(s.value), nil
}

func (s *basicLSMTSegment) Name() string {
	return s.path
}

func (s *basicLSMTSegment) Compare(other lsmtSegment) int {
	return strings.Compare(s.path, other.Name())
}

func (s *basicLSMTSegment) Open() error {
	fd, err := os.OpenFile(s.path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Println("could not open file descriptor for write")
		return err
	}
	s.wfd = fd
	s.rfd, err = os.OpenFile(s.path, os.O_RDONLY, 0666)
	if err != nil {
		log.Println("could not open file descriptor for read")
		return err
	}
	return nil
}

func (s *basicLSMTSegment) Close() error {
	var e error
	if s.wfd != nil {
		e = s.wfd.Close()
		s.wfd = nil
	}
	return e
}

func (s *basicLSMTSegment) String() string {
	return s.path
}

func (s *basicLSMTSegment) Insert(key, value string) error {
	pair := lsmtPair{
		key:   key,
		value: value,
	}
	s.pairs = append(s.pairs, pair)
	n, err := pair.Encode(s.wfd)
	if err != nil {
		log.Println("error writing segement pair with key: ", pair.key)
		return err
	}
	log.Println(n, "bytes written")
	return nil
}

func (s *basicLSMTSegment) Search(key string) (string, error) {
	var r string
	for _, pair := range s.pairs {
		if pair.key == key {
			return pair.value, nil
		}
	}

	for {
		pair := lsmtPair{}
		if _, err := pair.Decode(s.rfd); err != nil {
			break
		}
		log.Println("pair decoded on search", pair)
		s.pairs = append(s.pairs, pair)
		if pair.key == key {
			return pair.value, nil
		}

	}

	return r, errors.New("key not found (search)")
}

func newBasicLSMTSegment(opts lsmtSegmentConfig) lsmtSegment {
	// path = filepath.Join(path, "segment-" + strconv.Itoa(int(time.Now().UnixNano())))
	return &basicLSMTSegment{path: opts.Path(), limit: opts.Size()}

}

func newSegmentName(path string) string {
	return filepath.Join(path, "segment-"+strconv.Itoa(int(time.Now().UnixNano())))
}
