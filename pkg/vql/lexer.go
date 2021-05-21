package vql

import (
	"fmt"
	"strings"
)

var WHITESPACES = map[byte]int{
	'\n': 1,
	'\r': 1,
	' ':  1,
	'\t': 1,
}

type Lexer struct {
	currentChar     byte
	currentPosition int64
	nextPosition    int64
	input           string
	inputLength     int64
}

func NewLexer(code string) *Lexer {
	lexer := Lexer{
		currentPosition: 0,
		nextPosition:    0,
		input:           code,
		inputLength:     int64(len(code)),
	}
	lexer.readChar()
	return &lexer
}

func (l *Lexer) readChar() bool {
	end := false
	if l.nextPosition >= l.inputLength {
		l.currentChar = 0
		end = true
	} else {
		l.currentChar = l.input[l.nextPosition]
	}

	l.currentPosition = l.nextPosition
	l.nextPosition++
	return end
}

func (l *Lexer) peekChar() byte {
	if l.nextPosition >= l.inputLength {
		return 0
	}
	return l.input[l.nextPosition]
}

func (l *Lexer) NextToken() (Token, error) {
	var tok Token

	l.consumeWhitespace()
	currentPos := l.currentPosition

	switch l.currentChar {
	// Delimiters
	case ',':
		tok = newToken(COMMA, l.currentChar)
	case '{':
		tok = newToken(LBRACE, l.currentChar)
	case '}':
		tok = newToken(RBRACE, l.currentChar)
	case '(':
		tok = newToken(LPAREN, l.currentChar)
	case ')':
		tok = newToken(RPAREN, l.currentChar)
	case ';':
		tok = newToken(SEMICOLON, l.currentChar)
	case '[':
		tok = newToken(LBRACKET, l.currentChar)
	case ']':
		tok = newToken(RBRACKET, l.currentChar)
	case ':':
		tok = newToken(COLON, l.currentChar)
	case '"', '\'':
		prevChar := l.currentChar
		if l.readChar() {
			return IllegalToken, newErrorAtPosition(l.input, int(currentPos))
		}
		tok.Type = STRING
		literal, done := l.readString(prevChar)
		if done {
			return IllegalToken, newErrorAtPosition(l.input, int(currentPos))
		}
		tok.Literal = literal
	// Operators
	case '<':
		{
			if l.peekChar() == '=' {
				ch := l.currentChar
				l.readChar()
				tok.Type = LTE
				tok.Literal = string(ch) + string(l.currentChar)
			} else {
				tok = newToken(LT, l.currentChar)
			}
		}
	case '>':
		{
			if l.peekChar() == '=' {
				ch := l.currentChar
				l.readChar()
				tok.Type = GTE
				tok.Literal = string(ch) + string(l.currentChar)
			} else {
				tok = newToken(GT, l.currentChar)
			}
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.currentChar
			l.readChar()
			tok.Type = NotEq
			tok.Literal = string(ch) + string(l.currentChar)
		} else {
			tok = newToken(BANG, l.currentChar)
		}
	case '/':
		tok = newToken(SLASH, l.currentChar)
	case '+':
		tok = newToken(PLUS, l.currentChar)
	case '-':
		tok = newToken(MINUS, l.currentChar)
	case '*':
		tok = newToken(ASTERISK, l.currentChar)
	case '%':
		tok = newToken(PERCENT, l.currentChar)
	case '^':
		tok = newToken(POWER, l.currentChar)
	case '=':
		{
			ch := l.currentChar
			switch l.peekChar() {
			case '=':
				l.readChar()
				tok.Type = EQ
				tok.Literal = string(ch) + string(l.currentChar)
			default:
				tok = newToken(ASSIGNMENT, l.currentChar)
			}
		}
	case 0:
		tok.Literal = ""
		tok.Type = EOF
	default:
		if isDigit(l.currentChar) {
			return Token{Type: INT, Literal: l.readNumber()}, nil
		} else if isLetter(l.currentChar) {
			literal := l.readWord()
			ttype := LookupKeyword(literal)
			return Token{Type: ttype, Literal: literal}, nil
		}
		tok = newToken(ILLEGAL, l.currentChar)
	}
	l.readChar()
	return tok, nil
}

func (l *Lexer) consumeWhitespace() {
	for isWhiteSpace(l.currentChar) {
		if l.readChar() {
			return
		}
	}
}

func (l *Lexer) readNumber() string {
	pos := l.currentPosition
	for isDigit(l.currentChar) {
		if l.readChar() {
			break
		}
	}
	return l.input[pos:l.currentPosition]
}

func (l *Lexer) readWord() string {
	pos := l.currentPosition
	for isLetter(l.currentChar) {
		if l.readChar() {
			break
		}
	}
	return l.input[pos:l.currentPosition]
}

func (l *Lexer) readString(end byte) (string, bool) {
	pos := l.currentPosition
	ended := false

	for l.currentChar != end {
		if l.readChar() {
			ended = true
			break
		}
	}

	return l.input[pos:l.currentPosition], ended
}

func isDigit(char byte) bool {
	return char >= '0' && char <= '9'
}

func isLetter(char byte) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == '_'
}

func isWhiteSpace(char byte) bool {
	if _, ok := WHITESPACES[char]; ok {
		return true
	}
	return false
}

func newToken(tokenType TokenType, literal byte) Token {
	return Token{Type: tokenType, Literal: string(literal)}
}

func newErrorAtPosition(input string, pos int) error {
	marker := []rune(strings.Repeat("-", len(input)))
	marker[pos] = '^'
	return fmt.Errorf("unterminated string literal at position <%d>\n<%s>\n<%s>",
		pos, input, string(marker))
}
