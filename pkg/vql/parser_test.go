package vql

import (
	"fmt"
	"testing"
)

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

func TestParser_IndexStatement(t *testing.T) {
	payload := "INDEX index 'document content'"
	lexer := NewLexer(payload)
	parser := NewParser(lexer)
	query := parser.ParseQuery()
	fmt.Println(query.String())
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
