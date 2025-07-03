package visigoth

import (
	"sort"

	"github.com/sonirico/vago/slices"
)

// HitsSearch implements a hit-counting based search algorithm with AND logic.
//
// Algorithm:
// 1. For each search token, find all documents containing that token
// 2. Count hits per document (number of unique search tokens each document contains)
// 3. Filter documents that have ALL tokens (hits >= number of search tokens)
// 4. Sort results by hit count in descending order (most relevant first)
// 5. Return results in deterministic order by preserving document index order for ties
//
// Behavior:
// - Uses AND logic: only returns documents that contain ALL search tokens
// - Hit count = number of unique search tokens found in document (not total occurrences)
// - Results are sorted by relevance (hit count), then by document order for determinism
// - For multi-token queries, only documents with all tokens are returned
// - Time complexity: O(T * D + R log R) where T=tokens, D=avg docs per token, R=results
// - Space complexity: O(R) where R=number of matching documents
//
// Differences from LinearSearch:
// - HitsSearch: Uses hit counting with hash map lookup, then sorts by relevance
// - LinearSearch: Uses set intersection with early termination, preserves document order
// - Both implement AND logic but with different performance characteristics
// - HitsSearch is better for relevance ranking, LinearSearch for simple boolean matching
//
// Example:
//
//	Query: "java programming"
//	Doc1: "java tutorial" (hits=1, excluded - doesn't have "programming")
//	Doc2: "java programming guide" (hits=2, included)
//	Doc3: "advanced java programming concepts" (hits=2, included)
//	Result: [Doc2, Doc3] (both have hits=2, ordered by document index)
//
// Note: Hit counting is per unique token, not total occurrences:
//
//	"java java programming" with query "java programming" = 2 hits (not 3)
func HitsSearch(tokens []string, indexer Indexer) slices.Slice[SearchResult] {
	// Set threshold to number of tokens - implements AND logic
	// A document must contain ALL tokens to be included in results
	threshold := len(tokens)

	// Map to count hits per document (using document hash as key for uniqueness)
	docHits := make(map[HashKey]SearchResult)

	// Phase 1: Count hits for each document
	for _, token := range tokens {
		// Get all document indices that contain this token
		indexed := indexer.Indexed(token)
		if indexed == nil {
			continue
		}

		// For each document containing this token, increment its hit count
		for _, index := range indexed {
			doc := indexer.Document(index)
			hashKey := doc.Hash()

			if result, exists := docHits[hashKey]; exists {
				// Document already seen - increment hit count
				result.Hits++
				docHits[hashKey] = result
			} else {
				// First time seeing this document - initialize with 1 hit
				docHits[hashKey] = SearchResult{
					Document: doc,
					Hits:     1,
				}
			}
		}
	}

	// Phase 2: Filter documents that meet threshold and create results slice
	var results SearchResults

	// Iterate through documents in index order to ensure deterministic results
	// This prevents random ordering when documents have the same hit count
	for i := 0; i < indexer.Len(); i++ {
		doc := indexer.Document(i)
		hashKey := doc.Hash()

		// Only include documents that have ALL tokens (hits >= threshold)
		if result, exists := docHits[hashKey]; exists && result.Hits >= threshold {
			results = append(results, result)
		}
	}

	// Phase 3: Sort by relevance (hit count descending, then by document order for ties)
	sort.Sort(results)

	return slices.Slice[SearchResult](results)
}
