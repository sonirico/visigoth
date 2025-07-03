package visigoth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_intersection(t *testing.T) {
	tests := []struct {
		name     string
		L1, L2   []int
		Expected []int
	}{
		{
			name:     "Both nil",
			L1:       nil,
			L2:       nil,
			Expected: nil,
		},
		{
			name:     "L1 nil, L2 has elements",
			L1:       []int{0, 1, 2, 3, 4, 5},
			L2:       nil,
			Expected: nil,
		},
		{
			name:     "L1 has elements, L2 nil",
			L1:       nil,
			L2:       []int{0, 1, 2, 3, 4, 5},
			Expected: nil,
		},
		{
			name:     "Complete intersection",
			L1:       []int{0, 1, 2, 3, 4, 5},
			L2:       []int{0, 1, 2, 3, 4, 5},
			Expected: []int{0, 1, 2, 3, 4, 5},
		},
		{
			name:     "Partial intersection",
			L1:       []int{0, 1, 2, 3, 4, 5},
			L2:       []int{0, 5, 6},
			Expected: []int{0, 5},
		},
		{
			name:     "No intersection",
			L1:       []int{0},
			L2:       []int{1},
			Expected: []int{},
		},
		{
			name:     "Single element intersection",
			L1:       []int{0},
			L2:       []int{0},
			Expected: []int{0},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := intersection(test.L1, test.L2)
			if len(test.Expected) == 0 {
				assert.Empty(t, actual, "Expected empty intersection")
			} else {
				assert.Equal(t, test.Expected, actual)
			}
		})
	}
}

func TestLinearSearch(t *testing.T) {
	// Create a fresh index for each test
	tokenizr := NewKeepAlphanumericTokenizer()
	analyzer := NewTokenizationPipeline(
		tokenizr,
		NewLowerCaseTokenizer(),
		NewStopWordsFilter(SpanishStopWords),
		NewSpanishStemmer(true),
	)

	t.Run("Single token search", func(t *testing.T) {
		in := NewMemoryIndex("linear_single", analyzer)
		in.Put(NewDocRequest("doc1", "programacion en java"))
		in.Put(NewDocRequest("doc2", "programacion en php"))
		in.Put(NewDocRequest("doc3", "desarrollo web"))

		results := in.Search("programacion", LinearSearch)

		assert.Equal(t, 2, results.Len(), "Should find 2 documents with 'programacion'")

		// Verify all results have hits = 1 (single token)
		for _, result := range results {
			assert.Equal(t, 1, result.Hits, "Single token search should have Hits=1")
			assert.Contains(t, []string{"doc1", "doc2"}, result.Document.ID())
		}
	})

	t.Run("Multiple token search - AND logic", func(t *testing.T) {
		in := NewMemoryIndex("linear_multiple", analyzer)
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

		results := in.Search("programacion java", LinearSearch)

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

	t.Run("No matches when one token missing", func(t *testing.T) {
		in := NewMemoryIndex("linear_no_match", analyzer)
		in.Put(NewDocRequest("doc1", "programacion en php"))
		in.Put(NewDocRequest("doc2", "desarrollo java"))

		// Search for two terms where no document has both
		results := in.Search("programacion java", LinearSearch)

		assert.Equal(t, 0, results.Len(), "Should find no documents when no document has all terms")
	})

	t.Run("Empty query", func(t *testing.T) {
		in := NewMemoryIndex("linear_empty", analyzer)
		in.Put(NewDocRequest("doc1", "some content"))

		results := in.Search("", LinearSearch)

		assert.Equal(t, 0, results.Len(), "Empty query should return no results")
	})

	t.Run("Non-existent token", func(t *testing.T) {
		in := NewMemoryIndex("linear_nonexistent", analyzer)
		in.Put(NewDocRequest("doc1", "programacion en java"))

		results := in.Search("python", LinearSearch)

		assert.Equal(t, 0, results.Len(), "Non-existent token should return no results")
	})

	t.Run("Three token search", func(t *testing.T) {
		in := NewMemoryIndex("linear_three", analyzer)
		in.Put(NewDocRequest("doc1", "curso completo de programacion en java")) // has all three
		in.Put(
			NewDocRequest("doc2", "curso de programacion en php"),
		) // missing "completo"
		in.Put(
			NewDocRequest("doc3", "curso completo de desarrollo"),
		) // missing "programacion"

		results := in.Search("curso completo programacion", LinearSearch)

		assert.Equal(t, 1, results.Len(), "Should find only document with all three terms")

		// Get first result
		var foundResult *SearchResult
		for _, result := range results {
			foundResult = &result
			break
		}
		assert.NotNil(t, foundResult)
		assert.Equal(t, "doc1", foundResult.Document.ID())
		assert.Equal(t, 3, foundResult.Hits, "Should have Hits=3 for three tokens")
	})

	t.Run("Deterministic results", func(t *testing.T) {
		in := NewMemoryIndex("linear_deterministic", analyzer)
		in.Put(NewDocRequest("doc1", "programacion java"))
		in.Put(NewDocRequest("doc2", "java programacion"))
		in.Put(NewDocRequest("doc3", "curso de programacion en java"))

		// Run the same search multiple times
		var allResults [][]SearchResult
		searchTerm := "programacion java"

		for i := 0; i < 5; i++ {
			results := in.Search(searchTerm, LinearSearch)

			var resultSlice []SearchResult
			for _, result := range results {
				resultSlice = append(resultSlice, result)
			}
			allResults = append(allResults, resultSlice)
		}

		// Verify all searches returned the same results
		expectedCount := len(allResults[0])
		for i := 1; i < len(allResults); i++ {
			assert.Equal(t, expectedCount, len(allResults[i]),
				"Search %d returned different number of results", i+1)

			// LinearSearch should return results in document index order (deterministic)
			for j := 0; j < expectedCount; j++ {
				assert.Equal(t, allResults[0][j].Document.ID(), allResults[i][j].Document.ID(),
					"Search %d returned different document order", i+1)
			}
		}
	})

	t.Run("LinearSearch vs HitsSearch behavior difference", func(t *testing.T) {
		in := NewMemoryIndex("comparison", analyzer)
		in.Put(NewDocRequest("doc1", "programacion java")) // has both terms
		in.Put(NewDocRequest("doc2", "programacion php"))  // has only first term
		in.Put(NewDocRequest("doc3", "java tutorial"))     // has only second term

		// Both LinearSearch and HitsSearch should only return doc1 (AND logic)
		// The difference is in the algorithm, not the logic
		linearResults := in.Search("programacion java", LinearSearch)
		assert.Equal(t, 1, linearResults.Len(), "LinearSearch should find only 1 document")

		// Get first result from LinearSearch
		var linearResult *SearchResult
		for _, result := range linearResults {
			linearResult = &result
			break
		}
		assert.NotNil(t, linearResult)
		assert.Equal(t, "doc1", linearResult.Document.ID())

		// HitsSearch should also return only doc1 (same AND logic, different algorithm)
		hitsResults := in.Search("programacion java", HitsSearch)
		assert.Equal(t, 1, hitsResults.Len(), "HitsSearch should also find only 1 document")

		// Get first result from HitsSearch
		var hitsResult *SearchResult
		for _, result := range hitsResults {
			hitsResult = &result
			break
		}
		assert.NotNil(t, hitsResult)
		assert.Equal(t, "doc1", hitsResult.Document.ID())

		// Both should find the same document but LinearSearch uses intersection,
		// while HitsSearch uses hit counting with threshold filtering
		assert.Equal(t, linearResult.Document.ID(), hitsResult.Document.ID(),
			"Both algorithms should find the same document")
	})
}
