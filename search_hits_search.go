package visigoth

import (
	"sort"

	"github.com/sonirico/vago/slices"
)

// HitsSearch performs a search and returns results sorted by hit count
func HitsSearch(tokens []string, indexable Indexer) slices.Slice[SearchResult] {
	threshold := len(tokens)
	docHits := make(map[HashKey]SearchResult)

	for _, token := range tokens {
		indexed := indexable.Indexed(token)
		if indexed == nil {
			continue
		}

		for _, index := range indexed {
			doc := indexable.Document(index)
			hashKey := doc.Hash()

			if result, exists := docHits[hashKey]; exists {
				result.Hits++
				docHits[hashKey] = result
			} else {
				docHits[hashKey] = SearchResult{
					Document: doc,
					Hits:     1,
				}
			}
		}
	}

	// Filter results that meet the threshold and convert to slice
	var results SearchResults
	for _, result := range docHits {
		if result.Hits >= threshold {
			results = append(results, result)
		}
	}

	// Sort by hits (descending)
	sort.Sort(results)

	return slices.Slice[SearchResult](results)
}
