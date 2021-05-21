package search

import (
	"sort"

	"github.com/sonirico/visigoth/internal/container"
	"github.com/sonirico/visigoth/pkg/entities"
)

type HitsSearchRow interface {
	entities.Row
	Hits() int
}

type hitsSearchResultRow struct {
	annotated bool
	doc       entities.Doc
	hits      int
}

func (i hitsSearchResultRow) Hits() int {
	return i.hits
}

func (i hitsSearchResultRow) Doc() entities.Doc {
	return i.doc
}

func (i hitsSearchResultRow) Ser(serializer entities.Serializer) []byte {
	return serializer.Serialize(i)
}

type hitsSearchResult struct {
	items []hitsSearchResultRow
}

func newHitsSearchResult() hitsSearchResult {
	return hitsSearchResult{items: []hitsSearchResultRow{}}
}

func (s *hitsSearchResult) Add(row hitsSearchResultRow) {
	s.items = append(s.items, row)
}

func (s hitsSearchResult) Get(index int) entities.Row {
	if len(s.items) <= index {
		return nil
	}
	return s.items[index]
}

func (s hitsSearchResult) Len() int {
	return len(s.items)
}

func (s hitsSearchResult) Less(i, j int) bool {
	return s.items[i].hits > s.items[j].hits
}

func (s *hitsSearchResult) Swap(i, j int) {
	tmp := s.items[j]
	s.items[j] = s.items[i]
	s.items[i] = tmp
}

func HitsSearchEngine(tokens []string, indexable Indexer) entities.Iterator {
	threshold := len(tokens)
	docSet := make(map[entities.HashKey]hitsSearchResultRow)
	result := newHitsSearchResult()
	for _, token := range tokens {
		indexed := indexable.Indexed(token)
		if indexed == nil {
			continue
		}
		for _, index := range indexed {
			doc := indexable.Document(index)
			var searchInfo hitsSearchResultRow
			if inf, ok := docSet[doc.Hash()]; ok {
				inf.hits++
				searchInfo = inf
			} else {
				searchInfo = hitsSearchResultRow{
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
	sort.Sort(&result)
	return container.NewResultIterator(result)
}
