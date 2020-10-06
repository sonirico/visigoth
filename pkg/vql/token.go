package vql

import "strings"

const (
	// MISC
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"
	// Identifiers
	IdentifierTokenType = "ident"
	IndicesIdent        = "indices"
	IndexesIdent        = "indexes"

	// Literals
	INT    = "int"
	STRING = "string"
	TRUE   = "true"
	FALSE  = "false"

	// operators
	PLUS     = "+"
	MINUS    = "-"
	EQ       = "=="
	NOT_EQ   = "!="
	GT       = ">"
	LT       = "<"
	GTE      = ">="
	LTE      = "<="
	BANG     = "!"
	SLASH    = "/"
	ASTERISK = "*"
	PERCENT  = "%"
	POWER    = "^"

	//
	ASSIGNMENT = "="

	// keywords
	AsTokenType      = "AS"
	AkaTokenType     = "AKA"
	DropTokenType    = "DROP"
	ShowTokenType    = "SHOW"
	SearchTokenType  = "SEARCH"
	FromTokenType    = "FROM"
	UseTokenType     = "USE"
	IndexTokenType   = "INDEX"
	IntoTokenType    = "INTO"
	UsingTokenType   = "USING"
	AliasTokenType   = "ALIAS"
	UnAliasTokenType = "UNALIAS"

	// Delimiters
	COMMA     = ","
	COLON     = ":"
	SEMICOLON = ";"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"
	LBRACKET  = "["
	RBRACKET  = "]"
)

type TokenType string

var keywords = map[string]TokenType{
	"ALIAS":   AliasTokenType,
	"UNALIAS": UnAliasTokenType,
	"INTO":    IntoTokenType,
	"AS":      AsTokenType,
	"AKA":     AkaTokenType,
	"USING":   UsingTokenType,
	"DROP":    DropTokenType,
	"SHOW":    ShowTokenType,
	"SEARCH":  SearchTokenType,
	"FROM":    FromTokenType,
	"USE":     UseTokenType,
	"INDEX":   IndexTokenType,
	"true":    TRUE,
	"false":   FALSE,
}

func TokenIsKeyword(t Token) bool {
	if _, ok := keywords[strings.ToLower(t.Literal)]; ok {
		return true
	}

	if _, ok := keywords[strings.ToUpper(t.Literal)]; ok {
		return true
	}

	return false
}

func LookupKeyword(literal string) TokenType {
	if tt, ok := keywords[literal]; ok {
		return tt
	}
	if tt, ok := keywords[strings.ToUpper(literal)]; ok {
		return tt
	}
	return IdentifierTokenType
}

type Token struct {
	Type    TokenType
	Literal string
}

func (t Token) IsKeyword() bool {
	return TokenIsKeyword(t)
}

func NewToken(tokenType TokenType, literal string) *Token {
	return &Token{Type: tokenType, Literal: literal}
}

var (
	IllegalToken = Token{Type: ILLEGAL, Literal: ""}
)
