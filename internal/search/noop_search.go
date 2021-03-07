package search

import (
	"github.com/sonirico/visigoth/internal/container"
	"github.com/sonirico/visigoth/pkg/entities"
)

type noopResult struct {
	size int
}

func (n noopResult) Len() int { return n.size }

func (n noopResult) Get(index int) entities.Row {
	return nil
}

func NoopZeroSearchEngine(tokens []string, indexable Indexer) entities.Iterator {
	return container.NewResultIterator(&noopResult{size: 0})
}

func NoopAllSearchEngine(tokens []string, indexable Indexer) entities.Iterator {
	result := &noopResult{size: indexable.Len()}
	return container.NewResultIterator(result)
}
