package visigoth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndex_Search_One(t *testing.T) {
	tokenizr := NewKeepAlphanumericTokenizer()
	analyzer := NewTokenizationPipeline(
		tokenizr,
		NewLowerCaseTokenizer(),
		NewStopWordsFilter(SpanishStopWords),
		NewSpanishStemmer(true),
	)
	in := NewMemoryIndex("testing", analyzer)
	in.Put(NewDocRequest("/course/java", `Curso de programación en Java (León)`))
	in.Put(NewDocRequest("/course/php", `Curso de programación en PHP (León)`))

	// Test searching for "java"
	results := in.Search("java", HitsSearch)
	assert.Equal(t, 1, results.Len(), "unexpected search result size for term 'java'")

	// Iterate over results to get the first one
	var foundDoc *SearchResult
	for _, result := range results {
		foundDoc = &result
		break
	}

	assert.NotNil(t, foundDoc, "no results found")
	assert.Equal(t, "/course/java", foundDoc.Document.ID(), "unexpected document returned")
}

func TestIndex_Search_TwoDocuments(t *testing.T) {
	tokenizr := NewKeepAlphanumericTokenizer()
	analyzer := NewTokenizationPipeline(
		tokenizr,
		NewLowerCaseTokenizer(),
		NewStopWordsFilter(SpanishStopWords),
		NewSpanishStemmer(true),
	)
	in := NewMemoryIndex("testing", analyzer)
	in.Put(NewDocRequest("/course/java", `Curso de programacion en Java (León)`))
	in.Put(NewDocRequest("/course/php", `Curso de programacion en PHP (León)`))

	results := in.Search("programacion", HitsSearch)
	assert.Equal(t, 2, results.Len(), "unexpected search result size for term 'programacion'")

	// Verify both documents are returned
	foundJava := false
	foundPHP := false
	for _, result := range results {
		if result.Document.ID() == "/course/java" {
			foundJava = true
		}
		if result.Document.ID() == "/course/php" {
			foundPHP = true
		}
	}

	assert.True(t, foundJava, "Java course document is missing from search results")
	assert.True(t, foundPHP, "PHP course document is missing from search results")
}

func TestIndex_Search_Deterministic(t *testing.T) {
	tokenizr := NewKeepAlphanumericTokenizer()
	analyzer := NewTokenizationPipeline(
		tokenizr,
		NewLowerCaseTokenizer(),
		NewStopWordsFilter(SpanishStopWords),
		NewSpanishStemmer(true),
	)
	in := NewMemoryIndex("testing", analyzer)

	// Add multiple documents that will all match the search term
	in.Put(NewDocRequest("java-course", "programming course java"))
	in.Put(NewDocRequest("python-course", "programming course python"))
	in.Put(NewDocRequest("go-course", "programming course golang"))
	in.Put(NewDocRequest("js-course", "programming course javascript"))

	// Run the same search multiple times to verify deterministic results
	var allResults [][]string

	for i := 0; i < 5; i++ {
		results := in.Search("programming", HitsSearch)

		var docIDs []string
		for _, result := range results {
			docIDs = append(docIDs, result.Document.ID())
		}

		allResults = append(allResults, docIDs)
	}

	// Verify all runs produced the same results in the same order
	firstResult := allResults[0]
	for i := 1; i < len(allResults); i++ {
		assert.Equal(t, firstResult, allResults[i],
			"Search results should be deterministic across multiple runs")
	}

	// Verify we got all expected documents
	assert.Len(t, firstResult, 4, "Should find all 4 documents")
	assert.Contains(t, firstResult, "java-course")
	assert.Contains(t, firstResult, "python-course")
	assert.Contains(t, firstResult, "go-course")
	assert.Contains(t, firstResult, "js-course")
}
