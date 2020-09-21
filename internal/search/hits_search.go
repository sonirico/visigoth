package search

import (
	"fmt"
	"github.com/sonirico/visigoth/internal"
	"github.com/sonirico/visigoth/internal/container"
	"sort"
)

type HitsSearchRow interface {
	internal.Row
	Hits() int
}

type info struct {
	annotated bool
	doc       internal.Doc
	hits      int
}

func (i info) Hits() int {
	return i.hits
}

func (i info) Doc() internal.Doc {
	return i.doc
}

func (i *info) Ser(serializer internal.Serializer) []byte {
	return serializer.Serialize(i)
}

type hitsSearchResult struct {
	items []HitsSearchRow
}

func newHitsSearchResult() *hitsSearchResult {
	return &hitsSearchResult{items: []HitsSearchRow{}}
}

func (s *hitsSearchResult) Add(row HitsSearchRow) {
	s.items = append(s.items, row)
}

func (s *hitsSearchResult) Get(index int) internal.Row {
	if len(s.items) <= index {
		return nil
	}
	return s.items[index]
}

func (s *hitsSearchResult) Len() int {
	return len(s.items)
}

func (s *hitsSearchResult) Less(i, j int) bool {
	return s.items[i].Hits() > s.items[j].Hits()
}

func (s *hitsSearchResult) Swap(i, j int) {
	tmp := s.items[j]
	s.items[j] = s.items[i]
	s.items[i] = tmp
}

func (s *hitsSearchResult) Docs() []internal.Doc {
	docs := make([]internal.Doc, len(s.items))
	i := 0
	for _, info := range s.items {
		docs[i] = info.Doc()
		fmt.Println(fmt.Sprintf("%d hits ->>> %s", info.Hits(), info.Doc().Id()))
		i++
	}
	return docs
}

func HitsSearchEngine(tokens [][]byte, indexable Indexer) internal.Iterator {
	threshold := len(tokens)
	docSet := make(map[internal.HashKey]*info)
	result := newHitsSearchResult()
	for _, token := range tokens {
		indexed := indexable.Indexed(string(token))
		if indexed == nil {
			continue
		}
		for _, index := range indexed {
			doc := indexable.Document(index)
			var searchInfo *info
			if inf, ok := docSet[doc.Hash()]; ok {
				inf.hits++
				searchInfo = inf
			} else {
				searchInfo = &info{
					hits: 1,
					doc:  doc,
				}
				docSet[doc.Hash()] = searchInfo
			}
			if searchInfo.hits >= threshold && !searchInfo.annotated {
				searchInfo.annotated = true
				result.Add(searchInfo)
			}
		}
	}
	sort.Sort(result)
	return container.NewResultIterator(result)
}
