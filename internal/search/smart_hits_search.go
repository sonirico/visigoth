package search

import (
	"github.com/sonirico/visigoth/internal/container"
	"github.com/sonirico/visigoth/pkg/entities"
	"sort"
	"sync"
)

const (
	triggerSmartSearchThreshold = 20
)

type docSet struct {
	set sync.Map
	L   sync.RWMutex
}

func newDocSet() *docSet {
	return &docSet{
		set: sync.Map{},
		L:   sync.RWMutex{},
	}
}

func (d *docSet) check(key entities.HashKey) (*info, bool) {
	inter, ok := d.set.Load(key)
	if ok {
		info := inter.(*info)
		info.hits++
		return info, true
	}
	return nil, false
}

func (d *docSet) store(key entities.HashKey, info *info) {
	d.set.Store(key, info)
}

// SmartHitsSearchEngine will either use concurrent search or sequential search based on
// token amount
func SmartHitsSearchEngine(tokens [][]byte, indexable Indexer) entities.Iterator {
	// TODO: test benchmark
	if len(tokens) < triggerSmartSearchThreshold {
		return HitsSearchEngine(tokens, indexable)
	}
	threshold := len(tokens)
	docSet := newDocSet()
	result := newHitsSearchResult()
	wg := sync.WaitGroup{}
	for _, token := range tokens {
		indexed := indexable.Indexed(string(token))
		if indexed == nil {
			continue
		}
		for _, index := range indexed {
			wg.Add(1)
			go func(index int) {
				doc := indexable.Document(index)
				var searchInfo *info
				if inf, ok := docSet.check(doc.Hash()); ok {
					searchInfo = inf
				} else {
					searchInfo = &info{
						hits: 1,
						doc:  doc,
					}
					docSet.store(doc.Hash(), searchInfo)
				}
				if searchInfo.hits >= threshold && !searchInfo.annotated {
					searchInfo.annotated = true
					result.Add(searchInfo)
				}
				wg.Done()
			}(index)
		}
	}
	wg.Wait()
	sort.Sort(result)
	return container.NewResultIterator(result)
}
