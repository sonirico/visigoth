package repos

import (
	"fmt"
	"github.com/sonirico/visigoth/internal"
	"github.com/sonirico/visigoth/internal/search"
	"sync"
)

import vindex "github.com/sonirico/visigoth/internal/index"
import vtoken "github.com/sonirico/visigoth/internal/tokenizer"

type IndexRepo interface {
	List() []string
	Has(name string) bool
	Alias(alias string, in string) bool
	UnAlias(alias string) bool
	Put(in string, req internal.DocRequest)
	Search(index string, terms string, engine search.Engine) (internal.Iterator, error)
	Rename(old string, new string) bool
	Drop(in string) bool
}

// indexRepo handles a collection of indexes
type indexRepo struct {
	// TODO: two mutexes
	L       sync.RWMutex
	indices map[string]vindex.Index
	aliases map[string]string
	writers chan struct{}
}

func NewIndexRepo() IndexRepo {
	return &indexRepo{
		L:       sync.RWMutex{},
		aliases: make(map[string]string),
		indices: make(map[string]vindex.Index),
	}
}

func (h *indexRepo) List() []string {
	h.L.RLock()
	defer h.L.RUnlock()
	indices := make([]string, len(h.indices), len(h.indices))
	i := 0
	for iname := range h.indices {
		indices[i] = iname
		i++
	}
	return indices
}

func (h *indexRepo) Has(name string) bool {
	h.L.RLock()
	defer h.L.RUnlock()
	_, ok := h.indices[name]
	return ok
}

func (h *indexRepo) Alias(alias string, in string) bool {
	h.L.RLock()
	if _, ok := h.aliases[alias]; ok {
		// Alias already defined, override
		h.aliases[alias] = in
		h.L.RUnlock()
		return false
	}
	if _, ok := h.indices[in]; !ok {
		// No index to point to
		h.L.RUnlock()
		return false
	}
	h.L.RUnlock()
	h.L.Lock()
	h.aliases[alias] = in
	h.L.Unlock()
	return true
}

func (h *indexRepo) UnAlias(alias string) bool {
	h.L.RLock()
	if _, ok := h.aliases[alias]; !ok {
		// Alias already defined
		h.L.RUnlock()
		return false
	}
	h.L.RUnlock()
	h.L.Lock()
	delete(h.aliases, alias)
	h.L.Unlock()

	return true
}

func (h *indexRepo) Rename(old string, new string) bool {
	h.L.Lock()
	defer h.L.Unlock()

	index, ok := h.indices[old]
	if !ok {
		return false
	}

	h.indices[new] = index
	delete(h.indices, old)

	for alias, iname := range h.aliases {
		if iname == old {
			h.aliases[alias] = new
		}
	}

	return true
}

func (h *indexRepo) Search(iname string, terms string, engi search.Engine) (internal.Iterator, error) {
	h.L.RLock()
	in := h.getIndex(iname)
	if in == nil {
		h.L.RUnlock()
		return nil, fmt.Errorf("index with name '%s' does not exist", iname)
	}
	sr := in.Search(terms, engi)
	h.L.RUnlock()
	return sr, nil
}

func (h *indexRepo) Put(iname string, doc internal.DocRequest) {
	in := h.getIndex(iname)
	if in == nil {
		h.L.Lock()
		// TODO sanitize name
		// TODO parametrize index engine
		in = newMemoIndex(iname)
		h.indices[iname] = in
		h.L.Unlock()
	}
	in.Put(doc)
}

func (h *indexRepo) Drop(iname string) bool {
	h.L.RLock()
	if index := h.getIndex(iname); index == nil {
		h.L.RUnlock()
		return false
	}
	h.L.RUnlock()
	h.L.Lock()
	for alias, aliasedIndexName := range h.aliases {
		if iname == aliasedIndexName {
			delete(h.aliases, alias)
		}
	}
	delete(h.indices, iname)
	h.L.Unlock()
	return true
}

func (h *indexRepo) getIndex(name string) vindex.Index {
	in, ok := h.indices[name]
	if ok {
		return in
	}
	iname, ok := h.aliases[name]
	if !ok {
		return nil
	}
	in, ok = h.indices[iname]
	if !ok {
		return nil
	}
	return in
}

func newMemoIndex(iname string) *vindex.MemoryIndex {
	return vindex.NewMemoryIndex(iname, vtoken.SpanishTokenizer) // TODO: engines and/or build args
}
