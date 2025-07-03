package visigoth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHitsSearch(t *testing.T) {
	// Create a fresh tokenizer and analyzer for each test
	tokenizr := NewKeepAlphanumericTokenizer()
	analyzer := NewTokenizationPipeline(
		tokenizr,
		NewLowerCaseTokenizer(),
		NewStopWordsFilter(SpanishStopWords),
		NewSpanishStemmer(true),
	)

	t.Run("Single token search", func(t *testing.T) {
		in := NewMemoryIndex("hits_single", analyzer)
		in.Put(NewDocRequest("doc1", "programacion en java"))
		in.Put(NewDocRequest("doc2", "programacion en php"))
		in.Put(NewDocRequest("doc3", "desarrollo web"))

		results := in.Search("programacion", HitsSearch)

		assert.Equal(t, 2, results.Len(), "Should find 2 documents with 'programacion'")

		// Verify all results have hits = 1 (single token)
		for _, result := range results {
			assert.Equal(t, 1, result.Hits, "Single token search should have Hits=1")
			assert.Contains(t, []string{"doc1", "doc2"}, result.Document.ID())
		}
	})

	t.Run("Multiple token search - AND logic", func(t *testing.T) {
		in := NewMemoryIndex("hits_multiple", analyzer)
		in.Put(
			NewDocRequest("doc1", "curso de programacion en java"),
		) // has both "programacion" and "java"
		in.Put(
			NewDocRequest("doc2", "curso de programacion en php"),
		) // has "programacion" but not "java"
		in.Put(
			NewDocRequest("doc3", "tutorial de java para principiantes"),
		) // has "java" but not "programacion"
		in.Put(NewDocRequest("doc4", "desarrollo web frontend")) // has neither

		results := in.Search("programacion java", HitsSearch)

		assert.Equal(t, 1, results.Len(), "Should find only 1 document with BOTH terms")

		// Get first result
		var foundResult *SearchResult
		for _, result := range results {
			foundResult = &result
			break
		}
		assert.NotNil(t, foundResult)
		assert.Equal(t, "doc1", foundResult.Document.ID(), "Should find doc1 which has both terms")
		assert.Equal(t, 2, foundResult.Hits, "Should have Hits=2 for two tokens")
	})

	t.Run("Hit counting and threshold filtering", func(t *testing.T) {
		in := NewMemoryIndex("hits_counting", analyzer)
		in.Put(
			NewDocRequest("doc1", "java java java"),
		) // 3 occurrences of "java" but hits = 1 per unique token
		in.Put(
			NewDocRequest("doc2", "java programming java"),
		) // has both "java" and "programming" = 2 hits
		in.Put(NewDocRequest("doc3", "programming tutorial")) // 1 hit for "programming"
		in.Put(NewDocRequest("doc4", "python programming"))   // 1 hit for "programming"

		// Single token search should return all documents with that token
		results := in.Search("java", HitsSearch)
		assert.Equal(t, 2, results.Len(), "Should find 2 documents with 'java'")

		// Multiple token search requires ALL tokens
		results = in.Search("java programming", HitsSearch)
		assert.Equal(
			t,
			1,
			results.Len(),
			"Should find only 1 document with BOTH 'java' and 'programming'",
		)

		var foundResult *SearchResult
		for _, result := range results {
			foundResult = &result
			break
		}
		assert.NotNil(t, foundResult)
		assert.Equal(t, "doc2", foundResult.Document.ID())
		// Hits = number of unique tokens found, not total occurrences
		assert.Equal(
			t,
			2,
			foundResult.Hits,
			"Should have Hits=2 (1 for 'java' + 1 for 'programming')",
		)
	})

	t.Run("Relevance sorting by hit count", func(t *testing.T) {
		in := NewMemoryIndex("hits_sorting", analyzer)
		in.Put(
			NewDocRequest("doc1", "java programming guide"),
		) // 2 hits (1 for each token)
		in.Put(
			NewDocRequest("doc2", "java programming tutorial java"),
		) // 2 hits (1 per unique token, not 3)
		in.Put(
			NewDocRequest("doc3", "advanced java programming concepts"),
		) // 2 hits (1 for each token)

		results := in.Search("java programming", HitsSearch)
		assert.Equal(t, 3, results.Len(), "Should find all 3 documents")

		// Convert to slice to check order
		var resultSlice []SearchResult
		for _, result := range results {
			resultSlice = append(resultSlice, result)
		}

		// All documents have same hit count (2), so order should be by document index
		assert.Equal(t, 2, resultSlice[0].Hits, "First result should have 2 hits")
		assert.Equal(t, 2, resultSlice[1].Hits, "Second result should have 2 hits")
		assert.Equal(t, 2, resultSlice[2].Hits, "Third result should have 2 hits")

		// Order should be by document index (deterministic)
		assert.Equal(
			t,
			"doc1",
			resultSlice[0].Document.ID(),
			"doc1 should be first (lowest document index)",
		)
		assert.Equal(t, "doc2", resultSlice[1].Document.ID(), "doc2 should be second")
		assert.Equal(t, "doc3", resultSlice[2].Document.ID(), "doc3 should be third")
	})

	t.Run("Empty query returns no results", func(t *testing.T) {
		in := NewMemoryIndex("hits_empty", analyzer)
		in.Put(NewDocRequest("doc1", "some content"))

		results := in.Search("", HitsSearch)
		assert.Equal(t, 0, results.Len(), "Empty query should return no results")
	})

	t.Run("Non-existent token returns no results", func(t *testing.T) {
		in := NewMemoryIndex("hits_nonexistent", analyzer)
		in.Put(NewDocRequest("doc1", "programacion en java"))

		results := in.Search("python", HitsSearch)
		assert.Equal(t, 0, results.Len(), "Non-existent token should return no results")
	})

	t.Run("Partial match returns no results (AND logic)", func(t *testing.T) {
		in := NewMemoryIndex("hits_partial", analyzer)
		in.Put(NewDocRequest("doc1", "programacion en php")) // has "programacion" but not "java"
		in.Put(NewDocRequest("doc2", "desarrollo java"))     // has "java" but not "programacion"

		results := in.Search("programacion java", HitsSearch)
		assert.Equal(t, 0, results.Len(), "No document has both terms, should return no results")
	})

	t.Run("Three token search with threshold", func(t *testing.T) {
		in := NewMemoryIndex("hits_three", analyzer)
		in.Put(NewDocRequest("doc1", "curso completo de programacion en java")) // has all three
		in.Put(
			NewDocRequest("doc2", "curso de programacion en php"),
		) // missing "completo"
		in.Put(
			NewDocRequest("doc3", "curso completo de desarrollo"),
		) // missing "programacion"
		in.Put(NewDocRequest("doc4", "programacion completo tutorial")) // missing "curso"

		results := in.Search("curso completo programacion", HitsSearch)

		assert.Equal(t, 1, results.Len(), "Should find only document with all three terms")

		var foundResult *SearchResult
		for _, result := range results {
			foundResult = &result
			break
		}
		assert.NotNil(t, foundResult)
		assert.Equal(t, "doc1", foundResult.Document.ID())
		assert.Equal(t, 3, foundResult.Hits, "Should have Hits=3 for three tokens")
	})

	t.Run("Deterministic results across multiple searches", func(t *testing.T) {
		in := NewMemoryIndex("hits_deterministic", analyzer)
		in.Put(NewDocRequest("doc1", "programacion java"))
		in.Put(NewDocRequest("doc2", "java programacion"))
		in.Put(NewDocRequest("doc3", "curso de programacion en java"))

		// Run the same search multiple times
		var allResults [][]SearchResult
		searchTerm := "programacion java"

		for i := 0; i < 10; i++ {
			results := in.Search(searchTerm, HitsSearch)

			var resultSlice []SearchResult
			for _, result := range results {
				resultSlice = append(resultSlice, result)
			}
			allResults = append(allResults, resultSlice)
		}

		// Verify all searches returned the same results in the same order
		expectedCount := len(allResults[0])
		for i := 1; i < len(allResults); i++ {
			assert.Equal(t, expectedCount, len(allResults[i]),
				"Search %d returned different number of results", i+1)

			// Verify exact order (HitsSearch should be deterministic)
			for j := 0; j < expectedCount; j++ {
				assert.Equal(t, allResults[0][j].Document.ID(), allResults[i][j].Document.ID(),
					"Search %d returned different document order at position %d", i+1, j)
				assert.Equal(t, allResults[0][j].Hits, allResults[i][j].Hits,
					"Search %d returned different hit count at position %d", i+1, j)
			}
		}
	})

	t.Run("HitsSearch vs LinearSearch comparison", func(t *testing.T) {
		in := NewMemoryIndex("comparison", analyzer)
		in.Put(NewDocRequest("doc1", "programacion java"))         // has both terms
		in.Put(NewDocRequest("doc2", "programacion php"))          // has only first term
		in.Put(NewDocRequest("doc3", "java tutorial"))             // has only second term
		in.Put(NewDocRequest("doc4", "advanced java programming")) // different tokens

		// Both should implement AND logic and return only documents with ALL tokens
		hitsResults := in.Search("programacion java", HitsSearch)
		linearResults := in.Search("programacion java", LinearSearch)

		assert.Equal(t, 1, hitsResults.Len(), "HitsSearch should find 1 document")
		assert.Equal(t, 1, linearResults.Len(), "LinearSearch should find 1 document")

		// Both should find the same document
		var hitsResult, linearResult *SearchResult
		for _, result := range hitsResults {
			hitsResult = &result
			break
		}
		for _, result := range linearResults {
			linearResult = &result
			break
		}

		assert.NotNil(t, hitsResult)
		assert.NotNil(t, linearResult)
		assert.Equal(t, "doc1", hitsResult.Document.ID())
		assert.Equal(t, "doc1", linearResult.Document.ID())
		assert.Equal(t, hitsResult.Document.ID(), linearResult.Document.ID(),
			"Both algorithms should find the same document")
	})

	t.Run("Index state preservation", func(t *testing.T) {
		in := NewMemoryIndex("state_preservation", analyzer)
		in.Put(NewDocRequest("doc1", "test document one"))
		in.Put(NewDocRequest("doc2", "test document two"))

		// Capture initial state
		initialDocCount := len(in.Docs)
		initialIndexSize := len(in.InvertedIndex)

		// Perform multiple searches
		for i := 0; i < 5; i++ {
			results := in.Search("test", HitsSearch)
			assert.Equal(t, 2, results.Len(), "Search %d returned wrong number of results", i+1)

			// Verify index state hasn't changed
			assert.Equal(t, initialDocCount, len(in.Docs),
				"Document count changed after search %d", i+1)
			assert.Equal(t, initialIndexSize, len(in.InvertedIndex),
				"Inverted index size changed after search %d", i+1)
		}
	})

	t.Run("No phantom documents appear", func(t *testing.T) {
		in := NewMemoryIndex("phantom_test", analyzer)

		// Add exactly these documents
		expectedDocs := map[string]string{
			"course_java": "Curso de programacion en Java",
			"course_php":  "Curso de programacion en PHP",
			"course_go":   "Curso de programacion en Go",
		}

		for id, content := range expectedDocs {
			in.Put(NewDocRequest(id, content))
		}

		// Search for a term that should match all documents
		results := in.Search("programacion", HitsSearch)

		// Verify each result is one of our expected documents
		foundDocs := make(map[string]bool)
		for _, result := range results {
			docID := result.Document.ID()
			_, exists := expectedDocs[docID]
			assert.True(t, exists, "Found unexpected document with ID: %s", docID)
			foundDocs[docID] = true
		}

		// Verify we found all expected documents
		assert.Equal(t, len(expectedDocs), len(foundDocs),
			"Did not find all expected documents")
	})
}
