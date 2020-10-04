package vql

import (
	"fmt"
	"testing"
)

type testCaseIndex struct {
	query        string
	indexPayload string
	indexAka     string
	indexFormat  string
	indexName    string
}

type testCaseAlias struct {
	query       string
	indexName   string
	aliasName   string
	totalErrors int
}

func testAliasStatement(t *testing.T, actual Statement, expected testCaseAlias) bool {
	t.Helper()

	stmt, ok := actual.(*AliasStatement)

	if !ok {
		t.Errorf("unexpected statement type. want AliasStatement, have %T(%v)", stmt, stmt)
	}
	if stmt.Index.Literal() != expected.indexName {
		t.Errorf("unexpected index name. want '%s', have '%s'",
			expected.indexName, stmt.Index.Literal())
	}
	if stmt.Alias.Literal() != expected.aliasName {
		t.Errorf("unexpected index format. want '%s', have '%s'",
			expected.aliasName, stmt.Alias.Literal())
	}

	return true
}

func testIndexStatement(t *testing.T, actual Statement, expected testCaseIndex) bool {
	t.Helper()

	stmt, ok := actual.(*IndexStatement)

	if !ok {
		t.Errorf("unexpected statement type. want IndexStatement, have %T(%v)", stmt, stmt)
	}
	if stmt.Index.Literal() != expected.indexName {
		t.Errorf("unexpected index name. want '%s', have '%s'",
			expected.indexName, stmt.Index.Literal())
	}
	if stmt.Payload.Literal() != expected.indexPayload {
		t.Errorf("unexpected index payload. want '%s', have '%s'",
			expected.indexPayload, stmt.Payload.Literal())
	}
	if stmt.Format.Literal() != expected.indexFormat {
		t.Errorf("unexpected index format. want '%s', have '%s'",
			expected.indexFormat, stmt.Format.Literal())
	}

	return true
}

func TestParser_SearchStatement(t *testing.T) {
	payload := "SEARCH index 'string literal'"
	lexer := NewLexer(payload)
	parser := NewParser(lexer)
	query := parser.ParseQuery()
	fmt.Println(query.String())
}

func TestParser_SearchStatementWithUsing(t *testing.T) {
	payload := "SEARCH 'string literal' USING hits"
	lexer := NewLexer(payload)
	parser := NewParser(lexer)
	query := parser.ParseQuery()
	fmt.Println(query.String())
}

func TestParser_AliasStatement(t *testing.T) {
	tests := []testCaseAlias{
		{
			query:     "ALIAS 'index name' 'alias name'",
			indexName: "index name",
			aliasName: "alias name",
		},
		{
			query:     "ALIAS index_name index_alias",
			indexName: "index_name",
			aliasName: "index_alias",
		},
		{
			query:       "ALIAS 'alias name'",
			indexName:   "index",
			totalErrors: 1,
		},
	}

	for _, test := range tests {
		lexer := NewLexer(test.query)
		parser := NewParser(lexer)
		query := parser.ParseQuery()
		if len(parser.errors) == test.totalErrors {
			continue
		}
		t.Errorf("unexpected number of errors: %d, parser errors for query '%s'",
			len(parser.errors), test.query)
		for _, e := range parser.Errors() {
			t.Errorf(e)
		}
		if len(query.Statements) < 1 {
			t.Fatal("unexpected query with zero statements")
		}

		if !testAliasStatement(t, query.Statements[0], test) {
			t.Fatal()
		}
	}
}

func TestParser_IndexStatement(t *testing.T) {
	tests := []testCaseIndex{
		{
			query:        "INDEX 'document content' AKA 'content' AS JSON INTO 'index with spaces'",
			indexAka:     "content",
			indexPayload: "document content",
			indexFormat:  "JSON",
			indexName:    "index with spaces",
		},
		{
			query:        "INDEX 'document content' INTO 'index with spaces'",
			indexAka:     "",
			indexPayload: "document content",
			indexFormat:  "TEXT",
			indexName:    "index with spaces",
		},
		{
			query:        "INDEX 'document content' INTO index",
			indexAka:     "",
			indexPayload: "document content",
			indexFormat:  "TEXT",
			indexName:    "index",
		},
		{
			query:        "INDEX 'document content'",
			indexAka:     "",
			indexPayload: "document content",
			indexFormat:  "TEXT",
			indexName:    "",
		},
	}

	for _, test := range tests {
		lexer := NewLexer(test.query)
		parser := NewParser(lexer)
		query := parser.ParseQuery()
		if len(parser.errors) > 0 {
			t.Errorf("%d, parser errors for query '%s'",
				len(parser.errors), test.query)
			for _, e := range parser.Errors() {
				t.Errorf(e)
			}
		}
		if len(query.Statements) < 1 {
			t.Errorf("unexpected query with zero statements")
		}

		if !testIndexStatement(t, query.Statements[0], test) {
			t.Fatal()
		}
	}
}

func TestParser_UseStatement(t *testing.T) {
	payload := "USE index"
	lexer := NewLexer(payload)
	parser := NewParser(lexer)
	query := parser.ParseQuery()
	fmt.Println(query.String())
}

func TestParser_UseStatementSpaces(t *testing.T) {
	payload := "USE 'index with spaces'"
	lexer := NewLexer(payload)
	parser := NewParser(lexer)
	query := parser.ParseQuery()
	fmt.Println(query.String())
}

func TestParser_ShowIndicesStatement(t *testing.T) {
	payload := "SHOW indices"
	lexer := NewLexer(payload)
	parser := NewParser(lexer)
	query := parser.ParseQuery()
	fmt.Println(query.String())
}

func TestParser_DropIndexStatement(t *testing.T) {
	payload := "DROP index"
	lexer := NewLexer(payload)
	parser := NewParser(lexer)
	query := parser.ParseQuery()
	fmt.Println(query.String())
}
