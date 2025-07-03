package visigoth

import "github.com/sonirico/vago/slices"

type EngineType byte

const (
	NoopZero EngineType = iota
	NoopAll
	Hits
	SmartsHits
	Linear
)

type Indexer interface {
	Len() int
	Indexed(key string) []int
	Document(index int) Doc
}

// Engine defines the function signature for search functions
type Engine func(tokens []string, indexable Indexer) slices.Slice[SearchResult]
