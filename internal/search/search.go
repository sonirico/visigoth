package search

import (
	"github.com/sonirico/visigoth/internal"
)

type EngineType byte

const (
	NoopZero EngineType = iota
	NoopAll
	Hits
	SmartsHits
)

type Indexer interface {
	Len() int
	Indexed(key string) []int
	Document(index int) internal.Doc
}

type Engine func(tokens [][]byte, indexable Indexer) internal.Iterator
