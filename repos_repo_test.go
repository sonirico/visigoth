package visigoth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestIndexRepo() Repo {
	tokenizer := NewKeepAlphanumericTokenizer()
	pipeline := NewTokenizationPipeline(tokenizer, NewLowerCaseTokenizer())
	return NewIndexRepo(func(name string) Index {
		return NewMemoryIndex(name, pipeline)
	})
}

func Test_IndexRepo_Alias_Index_Exists(t *testing.T) {
	repo := newTestIndexRepo()

	repo.Put("dedos", NewDocRequest("pulgar", "este fue a por huevos"))
	repo.Put("colores", NewDocRequest("naranjito", "este es del 92"))

	ok := repo.Alias("dedos:latest", "dedos")
	assert.True(t, ok, "alias should be created successfully")
}

func Test_IndexRepo_Alias_Index_DoesNotExist(t *testing.T) {
	repo := newTestIndexRepo()

	repo.Put("dedos", NewDocRequest("pulgar", "este fue a por huevos"))
	repo.Put("colores", NewDocRequest("naranjito", "este es del 92"))

	ok := repo.Alias("dedos:latest", "sabores")
	assert.False(t, ok, "alias should not be created for non-existent index")
}

func Test_IndexRepo_UnAlias_All_Alias_Exists(t *testing.T) {
	repo := newTestIndexRepo()

	repo.Put("dedos", NewDocRequest("pulgar", "este fue a por huevos"))
	repo.Alias("dedos:latest", "dedos")

	ok := repo.UnAlias("dedos:latest", "")
	assert.True(t, ok, "alias should exist and be removed successfully")
}

func Test_IndexRepo_UnAlias_All_Alias_DoesNotExist(t *testing.T) {
	repo := newTestIndexRepo()

	repo.Put("dedos", NewDocRequest("pulgar", "este fue a por huevos"))

	ok := repo.UnAlias("dedos:latest", "")
	assert.False(t, ok, "alias should not exist")
}

func Test_IndexRepo_Search_By_Alias(t *testing.T) {
	repo := newTestIndexRepo()

	repo.Put("dedos", NewDocRequest("pulgar", "este fue a por huevos"))
	repo.Alias("dedos:latest", "dedos")

	_, err := repo.Search("dedos:latest", "huevos", NoopAllSearch)
	assert.NoError(t, err, "search by alias should return result without error")
}

func Test_IndexRepo_Search_By_AliasSeveralPointedIndices(t *testing.T) {
	repo := newTestIndexRepo()
	repo.Put("dedos", NewDocRequest("pulgar", "este fue a por huevos"))
	repo.Put("comida", NewDocRequest("huevos", "los huevos son cuerpos redondeados"))
	repo.Alias("huevos:latest", "dedos")
	repo.Alias("huevos:latest", "comida")

	res, err := repo.Search("huevos:latest", "huevos", HitsSearch)
	assert.NoError(t, err, "search by alias should return result without error")

	expectedDocuments := map[string]bool{"pulgar": false, "huevos": false}
	for res.Next() {
		item := res.Data()
		doc := item.Doc().ID()
		_, ok := expectedDocuments[doc]
		expectedDocuments[doc] = ok
	}

	for index, found := range expectedDocuments {
		assert.True(t, found, "expected document '%s' to be seen in result", index)
	}
}

func Test_IndexRepo_Put_By_Alias(t *testing.T) {
	repo := newTestIndexRepo()
	repo.Put("dedos", NewDocRequest("pulgar", "este fue a por huevos"))
	repo.Alias("dedos:latest", "dedos")
	repo.Put("dedos:latest", NewDocRequest("indice", "y este los casco"))

	_, err := repo.Search("dedos:latest", "casco", NoopAllSearch)
	assert.NoError(t, err, "search by alias should return result without error")
}

func Test_IndexRepo_Rename_IndexExists(t *testing.T) {
	repo := newTestIndexRepo()
	repo.Put("dedos", NewDocRequest("pulgar", "este fue a por huevos"))
	repo.Alias("dedos:latest", "dedos")

	ok := repo.Rename("dedos", "dedos_v2")
	assert.True(t, ok, "expected index 'dedos' to exist and be renamed")

	_, err := repo.Search("dedos:latest", "huevos", NoopAllSearch)
	assert.NoError(t, err, "search by alias should return result without error")
}

func Test_IndexRepo_Rename_IndexDoesNotExist(t *testing.T) {
	repo := newTestIndexRepo()
	repo.Put("dedos", NewDocRequest("pulgar", "este fue a por huevos"))
	repo.Alias("dedos:latest", "dedos")

	ok := repo.Rename("deditos", "dedos_v2")
	assert.False(t, ok, "expected index 'deditos' to be non-existent")
}

func Test_IndexRepo_HotSwap(t *testing.T) {
	repo := newTestIndexRepo()
	repo.Put("dedos", NewDocRequest("pulgar", "este fue a por huevos"))
	repo.Alias("dedos:latest", "dedos")
	repo.Put("dedos_v2", NewDocRequest("menique", "este los zampo"))
	repo.Alias("dedos:latest", "dedos_v2")

	r, err := repo.Search("dedos:latest", "zampo", NoopAllSearch)
	assert.NoError(t, err, "search by alias should return result without error")
	assert.NotNil(t, r, "result should not be nil")
}

func Test_IndexRepo_Drop_IndexExists(t *testing.T) {
	repo := newTestIndexRepo()
	repo.Put("dedos", NewDocRequest("pulgar", "este fue a por huevos"))

	ok := repo.Drop("dedos")
	assert.True(t, ok, "expected index 'dedos' to exist and be dropped")

	hasIndex := repo.Has("dedos")
	assert.False(t, hasIndex, "expected 'dedos' to have been dropped")
}

func Test_IndexRepo_Drop_IndexDoesNotExist(t *testing.T) {
	repo := newTestIndexRepo()

	ok := repo.Drop("dedos")
	assert.False(t, ok, "expected index 'dedos' to not exist")
}

func Test_IndexRepo_Drop_IndexWithAliasExists_ShouldDropAlias(t *testing.T) {
	repo := newTestIndexRepo()
	repo.Put("dedos", NewDocRequest("pulgar", "este fue a por huevos"))
	repo.Alias("dedos:latest", "dedos")

	ok := repo.Drop("dedos")
	assert.True(t, ok, "expected index 'dedos' to have been dropped")

	hasAlias := repo.HasAlias("dedos:latest")
	assert.False(t, hasAlias, "expected alias 'dedos:latest' to have been erased too")
}

func Test_IndexRepo_Search_Deterministic(t *testing.T) {
	repo := newTestIndexRepo()

	// Add multiple documents to the same index
	repo.Put("courses", NewDocRequest("java-course", "programming course java"))
	repo.Put("courses", NewDocRequest("python-course", "programming course python"))
	repo.Put("courses", NewDocRequest("go-course", "programming course golang"))
	repo.Put("courses", NewDocRequest("js-course", "programming course javascript"))

	// Run the same search multiple times to verify deterministic results
	var allResults [][]string

	for i := 0; i < 5; i++ {
		stream, err := repo.Search("courses", "programming", HitsSearch)
		assert.NoError(t, err)

		var docIDs []string
		for stream.Next() {
			result := stream.Data()
			docIDs = append(docIDs, result.Document.ID())
		}

		allResults = append(allResults, docIDs)
	}

	// Verify all runs produced the same results in the same order
	firstResult := allResults[0]
	for i := 1; i < len(allResults); i++ {
		assert.Equal(t, firstResult, allResults[i],
			"Repo search results should be deterministic across multiple runs")
	}

	// Verify we got all expected documents
	assert.Len(t, firstResult, 4, "Should find all 4 documents")
}
