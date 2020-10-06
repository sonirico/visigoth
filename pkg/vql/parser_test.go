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

type testCaseUnAlias = testCaseAlias

type testCaseUse struct {
	query     string
	indexName string
}

type testCaseShow struct {
	query string
	shown string
}

func testShowStatement(t *testing.T, actual Statement, expected testCaseShow) bool {
	t.Helper()

	stmt, ok := actual.(*ShowStatement)

	if !ok {
		t.Errorf("unexpected statement type. want AliasStatement, have %T(%v)", stmt, stmt)
		return false
	}
	if stmt.Shown.Literal() != expected.shown {
		t.Errorf("unexpected index name. want '%s', have '%s'",
			expected.shown, stmt.Shown.Literal())
		return false
	}

	return true
}

func testUseStatement(t *testing.T, actual Statement, expected testCaseUse) bool {
	t.Helper()

	stmt, ok := actual.(*UseStatement)

	if !ok {
		t.Errorf("unexpected statement type. want AliasStatement, have %T(%v)", stmt, stmt)
		return false
	}

	if stmt.Token.Type != UseTokenType {
		t.Errorf("unexpected token type. want '%s', have '%s'", UseTokenType, stmt.Token.Type)
	}

	if stmt.Used.Literal() != expected.indexName {
		t.Errorf("unexpected index name. want '%s', have '%s'",
			expected.indexName, stmt.Used.Literal())
		return false
	}

	return true
}

func testUnAliasStatement(t *testing.T, actual Statement, expected testCaseUnAlias) bool {
	t.Helper()

	stmt, ok := actual.(*UnAliasStatement)

	if !ok {
		t.Errorf("unexpected statement type. want UnAliasStatement, have %T(%v)", stmt, stmt)
		return false
	}
	if stmt.Index == nil {
		if len(expected.indexName) > 0 {
			t.Errorf("unexpected index name. want %s, have nil", expected.indexName)
			return false
		}
	} else if stmt.Index.Literal() != expected.indexName {
		t.Errorf("unexpected index name. want '%s', have '%s'",
			expected.indexName, stmt.Index.Literal())
		return false
	}

	if stmt.Alias != nil && stmt.Alias.Literal() != expected.aliasName {
		t.Errorf("unexpected index format. want '%s', have '%s'",
			expected.aliasName, stmt.Alias.Literal())
		return false
	}

	return true
}

func testAliasStatement(t *testing.T, actual Statement, expected testCaseAlias) bool {
	t.Helper()

	stmt, ok := actual.(*AliasStatement)

	if !ok {
		t.Errorf("unexpected statement type. want AliasStatement, have %T(%v)", stmt, stmt)
		return false
	}
	if stmt.Index.Literal() != expected.indexName {
		t.Errorf("unexpected index name. want '%s', have '%s'",
			expected.indexName, stmt.Index.Literal())
		return false
	}
	if stmt.Alias.Literal() != expected.aliasName {
		t.Errorf("unexpected index format. want '%s', have '%s'",
			expected.aliasName, stmt.Alias.Literal())
		return false
	}

	return true
}

func testIndexStatement(t *testing.T, actual Statement, expected testCaseIndex) bool {
	t.Helper()

	stmt, ok := actual.(*IndexStatement)

	if !ok {
		t.Errorf("unexpected statement type. want IndexStatement, have %T(%v)", stmt, stmt)
		return false
	}
	if stmt.Index.Literal() != expected.indexName {
		t.Errorf("unexpected index name. want '%s', have '%s'",
			expected.indexName, stmt.Index.Literal())
		return false
	}
	if stmt.Payload.Literal() != expected.indexPayload {
		t.Errorf("unexpected index payload. want '%s', have '%s'",
			expected.indexPayload, stmt.Payload.Literal())
		return false
	}
	if stmt.Format.Literal() != expected.indexFormat {
		t.Errorf("unexpected index format. want '%s', have '%s'",
			expected.indexFormat, stmt.Format.Literal())
		return false
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

func TestParser_UnAliasStatement(t *testing.T) {
	tests := []testCaseUnAlias{
		{
			query:     "unalias index as alias",
			indexName: "index name",
			aliasName: "alias name",
		},
		{
			query:     "UNALIAS 'index name' AS 'alias name'",
			indexName: "index name",
			aliasName: "alias name",
		},
		{
			query:     "UNALIAS index as alias",
			indexName: "index name",
			aliasName: "alias name",
		},
		{
			query:     "UNALIAS 'index name' 'alias name'",
			indexName: "index name",
			aliasName: "alias name",
		},
		{
			query:     "UNALIAS index_name index_alias",
			indexName: "index_name",
			aliasName: "index_alias",
		},
		{
			query:     "UNALIAS 'alias name'",
			indexName: "",
			aliasName: "alias name",
		},
		{
			query:       "UNALIAS",
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

		if !testUnAliasStatement(t, query.Statements[0], test) {
			t.Fatal()
		}
	}
}

func TestParser_AliasStatement(t *testing.T) {
	tests := []testCaseAlias{
		{
			query:     "alias index as alias",
			indexName: "index name",
			aliasName: "alias name",
		},
		{
			query:     "ALIAS 'index name' AS 'alias name'",
			indexName: "index name",
			aliasName: "alias name",
		},
		{
			query:     "ALIAS 'index name' 'alias name'",
			indexName: "index name",
			aliasName: "alias name",
		},
		{
			query:     "ALIAS index as alias",
			indexName: "index_name",
			aliasName: "index_alias",
		},
		{
			query:     "ALIAS index_name index_alias",
			indexName: "index_name",
			aliasName: "index_alias",
		},
		{
			query:       "ALIAS 'alias name'",
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
			query:        "index 'document content' aka 'content' as json into 'index with spaces'",
			indexAka:     "content",
			indexPayload: "document content",
			indexFormat:  "json",
			indexName:    "index with spaces",
		},
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
			t.Fatalf("unexpected query with zero statements")
		}

		if !testIndexStatement(t, query.Statements[0], test) {
			t.Fatal()
		}
	}
}

func TestParser_UseStatement(t *testing.T) {
	tests := []testCaseUse{
		{
			query:     "USE index",
			indexName: "index",
		},
		{
			query:     "use index",
			indexName: "index",
		},
		{
			query:     "use 'index with spaces'",
			indexName: "index with spaces",
		},
		{
			query:     "USE 'index'",
			indexName: "index",
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
			t.Fatalf("unexpected query with zero statements")
		}

		if !testUseStatement(t, query.Statements[0], test) {
			t.Fatal()
		}
	}
}

func TestParser_ShowIndicesStatement(t *testing.T) {
	tests := []testCaseShow{
		{
			query: "SHOW indices",
			shown: "indices",
		},
		{
			query: "show 'indices'",
			shown: "indices",
		},
		{
			query: "SHOW aliases",
			shown: "aliases",
		},
		{
			query: "show 'aliases'",
			shown: "aliases",
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
			t.Fatalf("unexpected query with zero statements")
		}

		if !testShowStatement(t, query.Statements[0], test) {
			t.Fatal()
		}
	}
}

func TestParser_DropIndexStatement(t *testing.T) {
	payload := "DROP index"
	lexer := NewLexer(payload)
	parser := NewParser(lexer)
	query := parser.ParseQuery()
	fmt.Println(query.String())
}
