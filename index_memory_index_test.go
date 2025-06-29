package visigoth

import (
	"testing"
)

type testResultRow struct {
	id string
}

func newTestResultRow(id string) *testResultRow {
	return &testResultRow{id: id}
}

func (tr *testResultRow) Doc() Doc {
	return Doc{Name: tr.id}
}

type testSearchResult struct {
	rows []Row
}

func newTestResult() *testSearchResult {
	return &testSearchResult{
		rows: make([]Row, 0),
	}
}

func (tr *testSearchResult) Add(row ...Row) Result {
	tr.rows = append(tr.rows, row...)
	return tr
}

func (tr *testSearchResult) Get(index int) Row {
	return tr.rows[index]
}

func (tr *testSearchResult) Len() int {
	return len(tr.rows)
}

type testSearch struct {
	term   string
	result Result
	engine Engine
}

func assertSearchReturns(t *testing.T, index Index, tests []testSearch) {
	for _, test := range tests {
		ares := index.Search(test.term, test.engine)
		counter := 0
		docs := make([]Doc, 0)
		for _, row := range ares {
			docs = append(docs, row.Doc())
			counter++
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
	tests := []testSearch{
		{
			term:   "java",
			result: newTestResult().Add(newTestResultRow("/course/java")),
			engine: HitsSearch,
		},
	}

	assertSearchReturns(t, in, tests)
}

func Test_Index_Search_Several(t *testing.T) {
	tokenizr := NewKeepAlphanumericTokenizer()
	analyzer := NewTokenizationPipeline(
		tokenizr,
		NewLowerCaseTokenizer(),
		NewStopWordsFilter(SpanishStopWords),
		NewSpanishStemmer(true))
	in := NewMemoryIndex("testing", analyzer)
	in.Put(NewDocRequest("/course/java", `Curso de programacion en Java (León)`))
	in.Put(NewDocRequest("/course/php", `Curso de programacion en PHP (León)`))
	tests := []testSearch{
		{
			term:   "java",
			result: newTestResult().Add(newTestResultRow("/course/java")),
			engine: HitsSearch,
		},
		{
			term: "programacion",
			result: newTestResult().Add(
				newTestResultRow("/course/java"),
				newTestResultRow("/course/php")),
			engine: LinearSearch,
		},
	}

	assertSearchReturns(t, in, tests)
}
