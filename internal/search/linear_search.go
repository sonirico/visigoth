package search

import (
	"github.com/sonirico/visigoth/internal/container"
	"github.com/sonirico/visigoth/pkg/entities"
)

// intersection results the elements in common between two slices. Considers that they have been previously sorted ASC
func intersection(a []int, b []int) []int {
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	r := make([]int, 0, maxLen)
	var i, j int
	for i < len(a) && j < len(b) {
		if a[i] < b[j] {
			i++
		} else if a[i] > b[j] {
			j++
		} else {
			r = append(r, a[i])
			i++
			j++
		}
	}
	return r
}

type indexer interface {
	Document(i int) entities.Doc
}

type linearSearchResultRow struct {
	doc entities.Doc
}

func (r linearSearchResultRow) Doc() entities.Doc {
	return r.doc
}

func (r linearSearchResultRow) Ser(serializer entities.Serializer) []byte {
	return serializer.Serialize(r)
}

type linearSearchResult struct {
	indexes []int
	indexer indexer
}

func (r linearSearchResult) Len() int {
	return len(r.indexes)
}

func (r linearSearchResult) Get(i int) entities.Row {
	if r.indexes == nil || i >= len(r.indexes) {
		return nil
	}
	doc := r.indexer.Document(i)
	return linearSearchResultRow{doc: doc}
}

func LinearSearchEngine(tokens []string, indexable Indexer) entities.Iterator {
	var indexedGroup []int
	for _, token := range tokens {
		if indexed := indexable.Indexed(token); indexed != nil {
			if indexedGroup == nil {
				indexedGroup = indexed
			} else {
				indexedGroup = intersection(indexedGroup, indexed)
			}
		}
	}
	return container.NewResultIterator(linearSearchResult{indexes: indexedGroup, indexer: indexable})
}
