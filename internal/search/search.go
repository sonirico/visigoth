package search

import (
	"github.com/sonirico/visigoth/pkg/entities"
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
	Document(index int) entities.Doc
}

type Engine func(tokens [][]byte, indexable Indexer) entities.Iterator
