package search

import (
	"github.com/sonirico/visigoth/internal"
	"github.com/sonirico/visigoth/internal/container"
)

type noopResult struct {
	size int
}

func (n noopResult) Len() int { return n.size }

func (n noopResult) Get(index int) internal.Row {
	return nil
}

func NoopZeroSearchEngine(tokens [][]byte, indexable Indexer) internal.Iterator {
	return container.NewResultIterator(&noopResult{size: 0})
}

func NoopAllSearchEngine(tokens [][]byte, indexable Indexer) internal.Iterator {
	result := &noopResult{size: indexable.Len()}
	return container.NewResultIterator(result)
}
