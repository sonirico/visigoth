package visigoth

import "github.com/sonirico/vago/slices"

// intersection returns the elements in common between two sorted slices.
// This function assumes both input slices are sorted in ascending order.
// It uses a two-pointer technique to find the intersection in O(n+m) time complexity.
//
// Example:
//
//	intersection([1, 3, 5, 7], [3, 5, 8, 9]) = [3, 5]
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

// LinearSearch performs an intersection-based search across all query tokens.
//
// Algorithm:
// 1. For each token in the query, get the list of documents that contain it
// 2. Find the intersection of all these document lists (documents that contain ALL tokens)
// 3. Return only documents that contain every single token in the query
//
// Key characteristics:
// - Uses AND logic: ALL tokens must be present in a document for it to match
// - Same logic as HitsSearch but different algorithm (intersection vs. hit counting)
// - Results have Hits = len(tokens) since all tokens are guaranteed to be found
// - More efficient for queries with many tokens due to early termination
// - Deterministic order based on document index order
//
// Comparison with HitsSearch:
// - Both implement AND logic (only documents with ALL tokens are returned)
// - LinearSearch: Uses set intersection operations (more efficient for large queries)
// - HitsSearch: Uses hit counting with threshold filtering (more flexible for scoring)
//
// Example:
//
//	Query: "programming java"
//	- Only returns documents that contain BOTH "programming" AND "java"
//	- A document with only "programming" will NOT be returned
//	- A document with only "java" will NOT be returned
func LinearSearch(tokens []string, indexable Indexer) slices.Slice[SearchResult] {
	if len(tokens) == 0 {
		return nil
	}

	// Start with the first token's documents
	docs := indexable.Indexed(tokens[0])
	if docs == nil {
		return nil
	}

	// Intersect with each subsequent token's documents
	for i := 1; i < len(tokens); i++ {
		nextDocs := indexable.Indexed(tokens[i])
		if nextDocs == nil {
			return nil
		}
		docs = intersection(docs, nextDocs)
		if len(docs) == 0 {
			return nil
		}
	}

	// Convert document indices to search results
	var results []SearchResult
	for _, docIndex := range docs {
		doc := indexable.Document(docIndex)
		results = append(results, SearchResult{
			Document: doc,
			Hits:     len(tokens), // All tokens were found
		})
	}

	return results
}
