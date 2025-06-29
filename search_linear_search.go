package visigoth

import "github.com/sonirico/vago/slices"

// intersection returns the elements in common between two slices. Considers that they have been previously sorted ASC
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

// LinearSearch performs a linear search and returns results directly
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
