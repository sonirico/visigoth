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

func TestDebugEstuviesIssue(t *testing.T) {
	tokenizr := NewKeepAlphanumericTokenizer()
	analyzer := NewTokenizationPipeline(
		tokenizr,
		NewLowerCaseTokenizer(),
		NewStopWordsFilter(SpanishStopWords),
		NewSpanishStemmer(true),
	)
	in := NewMemoryIndex("testing", analyzer)

	// Add only the documents we know about
	in.Put(NewDocRequest("/course/java", `Curso de programación en Java (León)`))
	in.Put(NewDocRequest("/course/php", `Curso de programación en PHP (León)`))

	// Check what documents are actually stored
	t.Logf("Number of documents in index: %d", len(in.Docs))
	for i, doc := range in.Docs {
		t.Logf("Doc[%d]: ID='%s', Content='%s'", i, doc.ID(), doc.Raw())
	}

	// Check the inverted index
	t.Logf("Inverted index contents:")
	for token, docIndices := range in.InvertedIndex {
		t.Logf("Token '%s' -> documents %v", token, docIndices)
	}

	// Perform a search
	results := in.Search("java", HitsSearch)
	t.Logf("Search results for 'java': %d results", results.Len())
	for _, result := range results {
		t.Logf("Result: ID='%s', Hits=%d, Content='%s'", result.Document.ID(), result.Hits, result.Document.Raw())
	}
}
