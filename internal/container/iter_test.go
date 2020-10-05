package container

import (
	"log"
	"testing"

	"github.com/sonirico/visigoth/pkg/entities"
)

type testItem struct {
	name string
}

func (t *testItem) Doc() entities.Doc {
	return entities.Doc{Name: t.name}
}

func (t *testItem) Ser(ser entities.Serializer) []byte {
	return nil
}

type testResult struct {
	Items []string
}

func (t *testResult) Len() int {
	return len(t.Items)
}

func (t *testResult) Get(i int) entities.Row {
	return &testItem{name: t.Items[i]}
}

func TestResultIterator_Chain(t *testing.T) {
	iterABC := NewResultIterator(&testResult{Items: []string{"a", "b", "c"}})
	iterXYZ := NewResultIterator(&testResult{Items: []string{"z", "y", "x"}})
	iter123 := NewResultIterator(&testResult{Items: []string{"1", "2", "3"}})
	finalIter := iterABC.Chain(iterXYZ).Chain(iter123)
	expected := []string{"a", "b", "c", "z", "y", "x", "1", "2", "3"}
	ei := 0
	for {
		item, done := finalIter.Next()
		if item != nil {
			value := item.Doc().Id()
			log.Println(value)
			if value != expected[ei] {
				t.Fatalf("unexpected yielded result. want '%s', have '%s'",
					expected[ei], value)
			}
			ei++
		}
		if done {
			break
		}
	}
}
