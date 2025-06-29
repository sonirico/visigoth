package visigoth

import "github.com/sonirico/vago/slices"

// NoopZeroSearch returns empty results
func NoopZeroSearch(tokens []string, indexable Indexer) slices.Slice[SearchResult] {
	return nil
}

// NoopAllSearch returns all documents as results
func NoopAllSearch(tokens []string, indexable Indexer) slices.Slice[SearchResult] {
	var results SearchResults
	for i := 0; i < indexable.Len(); i++ {
		doc := indexable.Document(i)
		results = append(results, SearchResult{
			Document: doc,
			Hits:     0, // No actual hits since this is a noop
		})
	}
	return slices.Slice[SearchResult](results)
}
