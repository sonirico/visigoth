package visigoth

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/sonirico/vago/slices"
	"github.com/sonirico/vago/streams"
)

type Repo interface {
	List() []string
	ListAliases() AliasesResult
	Has(name string) bool
	HasAlias(name string) bool
	Alias(alias string, in string) bool
	UnAlias(alias, index string) bool
	Put(in string, req DocRequest)
	Search(index string, terms string, engine Engine) (streams.ReadStream[SearchResult], error)
	Rename(old string, new string) bool
	Drop(in string) bool
}

type AliasesResultRow struct {
	Alias   string
	Indices []string
}

type AliasesResult struct {
	Aliases []AliasesResultRow
}

// IndexRepo handles a collection of indexes
type IndexRepo struct {
	indices   map[string]Index
	indicesMu *sync.RWMutex
	aliases   map[string][]string
	aliasesMu *sync.RWMutex

	indexBuilder Builder
}

func (h *IndexRepo) List() []string {
	h.indicesMu.RLock()
	defer h.indicesMu.RUnlock()
	indices := make([]string, len(h.indices), len(h.indices))
	i := 0
	for iname := range h.indices {
		indices[i] = iname
		i++
	}
	return indices
}

func (h *IndexRepo) Has(name string) bool {
	h.indicesMu.RLock()
	_, ok := h.indices[name]
	h.indicesMu.RUnlock()
	return ok
}

func (h *IndexRepo) HasAlias(name string) bool {
	h.aliasesMu.RLock()
	_, ok := h.aliases[name]
	h.aliasesMu.RUnlock()
	return ok
}

func (h *IndexRepo) Alias(alias string, index string) bool {
	// 1. Check the index exists
	h.indicesMu.RLock()
	if _, ok := h.indices[index]; !ok {
		h.indicesMu.RUnlock()
		return false
	}
	// 2. Index exists, apply alias
	h.aliasesMu.Lock()
	indices, ok := h.aliases[alias]
	added := false
	if ok {
		// alias already exists, check if already has the index
		exists := false
		for _, ri := range indices {
			if ri == index {
				exists = true
				break
			}
		}
		if !exists {
			h.aliases[alias] = append(indices, index)
			added = true
		}
	} else {
		// alias did not exist, create it
		h.aliases[alias] = []string{index}
		added = true
	}
	h.aliasesMu.Unlock()
	h.indicesMu.RUnlock()
	return added
}

func (h *IndexRepo) UnAlias(alias, index string) bool {
	if len(index) == 0 { // remove the entire alias if no index is specified
		h.aliasesMu.Lock()
		_, ok := h.aliases[alias]
		delete(h.aliases, alias)
		h.aliasesMu.Unlock()
		return ok
	}
	// only remove an index-alias association
	h.indicesMu.RLock()
	if _, ok := h.indices[index]; !ok {
		return false
	}
	h.indicesMu.RUnlock()
	h.aliasesMu.Lock()
	indices, ok := h.aliases[alias]
	if !ok {
		return false
	}
	// alias already exists, check if already has the index
	// TODO: improve
	var newIndices []string
	for _, aliasedIndexName := range indices {
		if aliasedIndexName != index {
			newIndices = append(newIndices, aliasedIndexName)
		}
	}
	if len(newIndices) != len(indices) {
		h.aliases[alias] = newIndices
	}
	h.aliasesMu.Unlock()
	return true
}

// Rename handles index renaming
func (h *IndexRepo) Rename(old string, new string) bool {
	// 1. Check the index exists
	h.indicesMu.Lock()
	index, ok := h.indices[old]
	if !ok {
		h.indicesMu.Unlock()
		return false
	}
	// 2. Perform the swap
	h.indices[new] = index
	delete(h.indices, old)
	// 3. Update indices to
	for _, indices := range h.aliases {
		for i, indexName := range indices {
			if indexName == old {
				indices[i] = new
			}
		}
	}
	h.indicesMu.Unlock()
	return true
}

func (h *IndexRepo) Search(
	indexName string,
	terms string,
	engine Engine,
) (streams.ReadStream[SearchResult], error) {
	indices, ok := h.getIndices(indexName)
	if !ok {
		return nil, fmt.Errorf("index with name '%s' does not exist", indexName)
	}

	h.indicesMu.RLock()
	if len(indices) == 1 {
		sr := indices[0].Search(terms, engine)
		h.indicesMu.RUnlock()
		return streams.MemReader(sr, nil), nil
	}

	var (
		wg     sync.WaitGroup
		result slices.Slice[SearchResult]
	)
	wg.Add(len(indices))

	rl := new(sync.Mutex)
	for _, index := range indices {
		go func(idx Index) {
			rl.Lock()
			sr := idx.Search(terms, engine)
			result.AppendVector(sr)
			rl.Unlock()
			wg.Done()
		}(index)
	}
	wg.Wait()
	h.indicesMu.RUnlock()
	return streams.MemReader(result, nil), nil
}

func (h *IndexRepo) Put(indexName string, doc DocRequest) {
	indices, ok := h.getIndices(indexName)
	h.indicesMu.Lock()
	if !ok {
		// TODO sanitize name
		// TODO parametrize index engine
		in := h.indexBuilder(indexName)
		h.indices[indexName] = in
		in.Put(doc)
		h.indicesMu.Unlock()
		return
	}
	var wg sync.WaitGroup
	wg.Add(len(indices))
	for _, in := range indices {
		go func(index Index) {
			index.Put(doc)
			wg.Done()
		}(in)
	}
	wg.Wait()
	h.indicesMu.Unlock()
}

func (h *IndexRepo) Drop(indexName string) bool {
	h.indicesMu.RLock()
	_, ok := h.indices[indexName]
	if !ok {
		h.indicesMu.RUnlock()
		return false
	}
	h.indicesMu.RUnlock()
	h.indicesMu.Lock()
	h.aliasesMu.Lock()
	// Drop references to index
	for alias, indices := range h.aliases {
		var newIndices []string
		for _, aliasedIndexName := range indices {
			if aliasedIndexName != indexName {
				newIndices = append(newIndices, indexName)
			}
		}
		if len(newIndices) < 1 {
			delete(h.aliases, alias)
		} else if len(newIndices) != len(indices) {
			h.aliases[alias] = newIndices
		}
	}
	// Actually drop the index
	delete(h.indices, indexName)
	h.indicesMu.Unlock()
	h.aliasesMu.Unlock()
	return true
}

func (h *IndexRepo) ListAliases() AliasesResult {
	h.aliasesMu.RLock()
	aliases := make([]AliasesResultRow, len(h.aliases), len(h.aliases))
	i := 0
	for k, v := range h.aliases {
		aliases[i] = AliasesResultRow{Alias: k, Indices: v}
		i++
	}
	h.aliasesMu.RUnlock()
	return AliasesResult{Aliases: aliases}
}

func (h *IndexRepo) getIndices(name string) ([]Index, bool) {
	// 1. Search on indices directly
	h.indicesMu.RLock()
	in, ok := h.indices[name]
	if ok {
		h.indicesMu.RUnlock()
		return []Index{in}, true
	}
	// 2. If not found, search on aliases
	h.aliasesMu.RLock()
	aliasedIndices, ok := h.aliases[name]
	if !ok {
		h.aliasesMu.RUnlock()
		h.indicesMu.RUnlock()
		return nil, false
	}
	res := make([]Index, len(aliasedIndices))
	for i, index := range aliasedIndices {
		res[i], _ = h.indices[index]
	}
	h.indicesMu.RUnlock()
	h.aliasesMu.RUnlock()
	return res, true
}

func (h *IndexRepo) String() string {
	h.indicesMu.RLock()
	var buf bytes.Buffer
	for _, index := range h.indices {
		buf.WriteString(fmt.Sprintln(index))
	}
	h.indicesMu.RUnlock()
	return buf.String()
}

func NewIndexRepo(builder Builder) *IndexRepo {
	return &IndexRepo{
		indices:      make(map[string]Index),
		indicesMu:    new(sync.RWMutex),
		aliases:      make(map[string][]string),
		aliasesMu:    new(sync.RWMutex),
		indexBuilder: builder,
	}
}
