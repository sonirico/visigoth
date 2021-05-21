package index

import (
	"testing"

	"github.com/sonirico/visigoth/internal/search"
	"github.com/sonirico/visigoth/pkg/analyze"
	"github.com/sonirico/visigoth/pkg/entities"
)

type testResultRow struct {
	id string
}

func newTestResultRow(id string) *testResultRow {
	return &testResultRow{id: id}
}

func (tr *testResultRow) Doc() entities.Doc {
	return entities.Doc{Name: tr.id}
}

func (tr *testResultRow) Ser(serializer entities.Serializer) []byte {
	return nil
}

type testSearchResult struct {
	rows []entities.Row
}

func newTestResult() *testSearchResult {
	return &testSearchResult{
		rows: make([]entities.Row, 0),
	}
}

func (tr *testSearchResult) Add(row ...entities.Row) entities.Result {
	tr.rows = append(tr.rows, row...)
	return tr
}

func (tr *testSearchResult) Get(index int) entities.Row {
	return tr.rows[index]
}

func (tr *testSearchResult) Len() int {
	return len(tr.rows)
}

type testSearch struct {
	term   string
	result entities.Result
	engine search.Engine
}

func assertSearchReturns(t *testing.T, index Index, tests []testSearch) {
	for _, test := range tests {
		ares := index.Search(test.term, test.engine)
		counter := 0
		docs := make([]entities.Doc, 0)
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
			t.Fatalf("unexpected search result size for term '%s'. want %d, have %d results",
				test.term, test.result.Len(), counter)
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
					erow.Doc().ID())
			}
		}
	}
}

func Test_Index_Search_One(t *testing.T) {
	tokenizr := analyze.NewKeepAlphanumericTokenizer()
	analyzer := analyze.NewTokenizationPipeline(&tokenizr,
		analyze.NewLowerCaseTokenizer(),
		analyze.NewStopWordsFilter(analyze.SpanishStopWords),
		analyze.NewSpanishStemmer(true))
	in := NewMemoryIndex("testing", &analyzer)
	in.Put(entities.NewDocRequest("/course/java", `Curso de programación en Java (León)`))
	in.Put(entities.NewDocRequest("/course/php", `Curso de programación en PHP (León)`))
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
	tokenizr := analyze.NewKeepAlphanumericTokenizer()
	analyzer := analyze.NewTokenizationPipeline(&tokenizr,
		analyze.NewLowerCaseTokenizer(),
		analyze.NewStopWordsFilter(analyze.SpanishStopWords),
		analyze.NewSpanishStemmer(true))
	in := NewMemoryIndex("testing", &analyzer)
	in.Put(entities.NewDocRequest("/course/java", `Curso de programacion en Java (León)`))
	in.Put(entities.NewDocRequest("/course/php", `Curso de programacion en PHP (León)`))
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
			engine: search.LinearSearchEngine,
		},
	}

	assertSearchReturns(t, in, tests)
}
