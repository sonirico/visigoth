package repos

import (
	"testing"

	"github.com/sonirico/visigoth/pkg/analyze"

	vindex "github.com/sonirico/visigoth/internal/index"

	"github.com/sonirico/visigoth/internal/search"
	"github.com/sonirico/visigoth/pkg/entities"
)

func newIndexRepo() IndexRepo {
	tokenizer := analyze.NewKeepAlphanumericTokenizer()
	pipeline := analyze.NewTokenizationPipeline(&tokenizer, analyze.NewLowerCaseTokenizer())
	return NewIndexRepo(func(name string) vindex.Index {
		return vindex.NewMemoryIndex(name, &pipeline)
	})
}

func Test_IndexRepo_Alias_Index_Exists(t *testing.T) {
	repo := newIndexRepo()

	repo.Put("dedos", entities.NewDocRequest("pulgar", "este fue a por huevos"))
	repo.Put("colores", entities.NewDocRequest("naranjito", "este es del 92"))

	if ok := repo.Alias("dedos:latest", "dedos"); !ok {
		t.Errorf("alias failed. want alias created, have otherwise")
	}
}

func Test_IndexRepo_Alias_Index_DoesNotExist(t *testing.T) {
	repo := newIndexRepo()

	repo.Put("dedos", entities.NewDocRequest("pulgar", "este fue a por huevos"))
	repo.Put("colores", entities.NewDocRequest("naranjito", "este es del 92"))

	if ok := repo.Alias("dedos:latest", "sabores"); ok {
		t.Errorf("alias failed. want alias not created, have otherwise")
	}
}

func Test_IndexRepo_UnAlias_All_Alias_Exists(t *testing.T) {
	repo := newIndexRepo()

	repo.Put("dedos", entities.NewDocRequest("pulgar", "este fue a por huevos"))
	repo.Alias("dedos:latest", "dedos")

	if ok := repo.UnAlias("dedos:latest", ""); !ok {
		t.Errorf("alias failed. alias should exist, have otherwise")
	}
}

func Test_IndexRepo_UnAlias_All_Alias_DoesNotExist(t *testing.T) {
	repo := newIndexRepo()

	repo.Put("dedos", entities.NewDocRequest("pulgar", "este fue a por huevos"))

	if ok := repo.UnAlias("dedos:latest", ""); ok {
		t.Errorf("alias failed. received alias that should not exist, have otherwise")
	}
}

func Test_IndexRepo_Search_By_Alias(t *testing.T) {
	repo := newIndexRepo()

	repo.Put("dedos", entities.NewDocRequest("pulgar", "este fue a por huevos"))
	repo.Alias("dedos:latest", "dedos")
	_, err := repo.Search("dedos:latest", "huevos", search.NoopAllSearchEngine)
	if err != nil {
		t.Errorf("unexpected error. want search by alias return result, have error %s", err)
	}
}

func Test_IndexRepo_Search_By_AliasSeveralPointedIndices(t *testing.T) {
	repo := newIndexRepo()
	repo.Put("dedos", entities.NewDocRequest("pulgar", "este fue a por huevos"))
	repo.Put("comida", entities.NewDocRequest("huevos", "los huevos son cuerpos redondeados"))
	repo.Alias("huevos:latest", "dedos")
	repo.Alias("huevos:latest", "comida")
	res, err := repo.Search("huevos:latest", "huevos", search.HitsSearchEngine)
	if err != nil {
		t.Errorf("unexpected error. want search by alias return result, have error %s", err)
	}
	expectedDocuments := map[string]bool{"pulgar": false, "huevos": false}
	for {
		item, done := res.Next()
		if item != nil {
			doc := item.Doc().ID()
			_, ok := expectedDocuments[doc]
			expectedDocuments[doc] = ok
		}
		if done {
			break
		}
	}
	for index, ok := range expectedDocuments {
		if !ok {
			t.Errorf("expected document '%s' to be seen in result", index)
		}
	}
}

func Test_IndexRepo_Put_By_Alias(t *testing.T) {
	repo := newIndexRepo()
	repo.Put("dedos", entities.NewDocRequest("pulgar", "este fue a por huevos"))
	repo.Alias("dedos:latest", "dedos")
	repo.Put("dedos:latest", entities.NewDocRequest("indice", "y este los casco"))
	_, err := repo.Search("dedos:latest", "casco", search.NoopAllSearchEngine)
	if err != nil {
		t.Errorf("unexpected error. want search by alias return result, have error %s", err.Error())
		return
	}
}

func Test_IndexRepo_Rename_IndexExists(t *testing.T) {
	repo := newIndexRepo()
	repo.Put("dedos", entities.NewDocRequest("pulgar", "este fue a por huevos"))
	repo.Alias("dedos:latest", "dedos")
	if ok := repo.Rename("dedos", "dedos_v2"); !ok {
		t.Errorf("expected index 'dedos' to exist")
		return
	}
	_, err := repo.Search("dedos:latest", "huevos", search.NoopAllSearchEngine)
	if err != nil {
		t.Errorf("unexpected error. want search by alias return result, have error %s", err.Error())
		return
	}
}

func Test_IndexRepo_Rename_IndexDoesNotExist(t *testing.T) {
	repo := newIndexRepo()
	repo.Put("dedos", entities.NewDocRequest("pulgar", "este fue a por huevos"))
	repo.Alias("dedos:latest", "dedos")
	if ok := repo.Rename("deditos", "dedos_v2"); ok {
		t.Errorf("expected index 'deditos' to be non-existent")
		return
	}
}

func Test_IndexRepo_HotSwap(t *testing.T) {
	repo := newIndexRepo()
	repo.Put("dedos", entities.NewDocRequest("pulgar", "este fue a por huevos"))
	repo.Alias("dedos:latest", "dedos")
	repo.Put("dedos_v2", entities.NewDocRequest("menique", "este los zampo"))
	repo.Alias("dedos:latest", "dedos_v2")
	r, err := repo.Search("dedos:latest", "zampo", search.NoopAllSearchEngine)
	if err != nil {
		t.Errorf("unexpected error. want search by alias return result, have error %s %s", err.Error(), r)
		return
	}
}

func Test_IndexRepo_Drop_IndexExists(t *testing.T) {
	repo := newIndexRepo()
	repo.Put("dedos", entities.NewDocRequest("pulgar", "este fue a por huevos"))
	if ok := repo.Drop("dedos"); !ok {
		t.Errorf("expected index 'dedos' to exist")
		return
	}
	if repo.Has("dedos") {
		t.Errorf("unexpected drop result. expected '%s' to have been dropped, but wasn't", "dedos")
		return
	}
}

func Test_IndexRepo_Drop_IndexDoesNotExist(t *testing.T) {
	repo := newIndexRepo()
	if ok := repo.Drop("dedos"); ok {
		t.Errorf("expected index 'dedos' to have been erased")
		return
	}
}

func Test_IndexRepo_Drop_IndexWithAliasExists_ShouldDropAlias(t *testing.T) {
	repo := newIndexRepo()
	repo.Put("dedos", entities.NewDocRequest("pulgar", "este fue a por huevos"))
	repo.Alias("dedos:latest", "dedos")
	if ok := repo.Drop("dedos"); !ok {
		t.Errorf("expected index 'dedos' to have been dropped")
		return
	}
	if repo.HasAlias("dedos:latest") {
		t.Errorf("expected alias '%s' to have been erased too", "dedos:latest")
		return
	}
}
