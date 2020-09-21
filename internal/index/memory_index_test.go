package index

import (
	"github.com/sonirico/visigoth/internal"
	"github.com/sonirico/visigoth/internal/search"
	"testing"

	visigoth "github.com/sonirico/visigoth/internal/tokenizer"
)

type testResultRow struct {
	id string
}

func newTestResultRow(id string) *testResultRow {
	return &testResultRow{id: id}
}

func (tr *testResultRow) Doc() internal.Doc {
	return internal.Doc{Name: tr.id}
}

func (tr *testResultRow) Ser(serializer internal.Serializer) []byte {
	return nil
}

type testSearchResult struct {
	rows []internal.Row
}

func newTestResult() *testSearchResult {
	return &testSearchResult{
		rows: make([]internal.Row, 0),
	}
}

func (tr *testSearchResult) Add(row ...internal.Row) internal.Result {
	tr.rows = append(tr.rows, row...)
	return tr
}

func (tr *testSearchResult) Get(index int) internal.Row {
	return tr.rows[index]
}

func (tr *testSearchResult) Len() int {
	return len(tr.rows)
}

type testSearch struct {
	term   string
	result internal.Result
	engine search.Engine
}

func assertSearchReturns(t *testing.T, index Index, tests []testSearch) {
	for _, test := range tests {
		ares := index.Search(test.term, test.engine)
		counter := 0
		docs := make([]internal.Doc, 0)
		for {
			row, done := ares.Next()
			if row != nil {
				docs = append(docs, row.Doc())
				counter++
			}
			if done {
				break
			}
		}

		if counter != test.result.Len() {
			t.Fatalf("unexpected search result size. want %d, have %d results",
				test.result.Len(), counter)
		}

		for i := 0; i < counter; i++ {
			erow := test.result.Get(i)
			found := false
			for _, adoc := range docs {
				if adoc.Hash() == erow.Doc().Hash() {
					found = true
				}
			}
			if !found {
				t.Fatalf("'%s' document is missing, but should be present",
					erow.Doc().Id())
			}
		}
	}
}

func Test_Index_Search_One(t *testing.T) {
	in := NewMemoryIndex("testing", visigoth.NewSimpleTokenizer())
	in.Put(internal.NewDocRequest("/course/java", `Curso de programación en Java (León)`))
	in.Put(internal.NewDocRequest("/course/php", `Curso de programación en PHP (León)`))
	tests := []testSearch{
		{
			term:   "java",
			result: newTestResult().Add(newTestResultRow("/course/java")),
			engine: search.HitsSearchEngine,
		},
	}

	assertSearchReturns(t, in, tests)
}

func Test_Index_Search_Several(t *testing.T) {
	in := NewMemoryIndex("testing", visigoth.NewSimpleTokenizer())
	in.Put(internal.NewDocRequest("/course/java", `Curso de programación en Java (León)`))
	in.Put(internal.NewDocRequest("/course/php", `Curso de programación en PHP (León)`))
	tests := []testSearch{
		{
			term:   "java",
			result: newTestResult().Add(newTestResultRow("/course/java")),
			engine: search.HitsSearchEngine,
		},
		{
			term: "programacion",
			result: newTestResult().Add(
				newTestResultRow("/course/java"),
				newTestResultRow("/course/php")),
			engine: search.HitsSearchEngine,
		},
	}

	assertSearchReturns(t, in, tests)
}
