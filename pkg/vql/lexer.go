package vql

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

func (l *Lexer) readChar() {
	if l.nextPosition >= l.inputLength {
		l.currentChar = 0
	} else {
		l.currentChar = l.input[l.nextPosition]
	}

	l.currentPosition = l.nextPosition
	l.nextPosition += 1
}

func (l *Lexer) peekChar() byte {
	if l.nextPosition >= l.inputLength {
		return 0
	}
	return l.input[l.nextPosition]
}

func (l *Lexer) NextToken() Token {
	var tok Token

	l.consumeWhitespace()

	switch l.currentChar {
	// Delimiters
	case ',':
		tok = newToken(COMMA, l.currentChar)
	case '{':
		tok = newToken(LBRACE, l.currentChar)
		break
	case '}':
		tok = newToken(RBRACE, l.currentChar)
		break
	case '(':
		tok = newToken(LPAREN, l.currentChar)
		break
	case ')':
		tok = newToken(RPAREN, l.currentChar)
		break
	case ';':
		tok = newToken(SEMICOLON, l.currentChar)
		break
	case '[':
		tok = newToken(LBRACKET, l.currentChar)
		break
	case ']':
		tok = newToken(RBRACKET, l.currentChar)
		break
	case ':':
		tok = newToken(COLON, l.currentChar)
		break
	case '"', '\'':
		prevChar := l.currentChar
		l.readChar()
		tok.Type = STRING
		tok.Literal = l.readString(prevChar)
		break
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
			break
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
			break
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.currentChar
			l.readChar()
			tok.Type = NOT_EQ
			tok.Literal = string(ch) + string(l.currentChar)
		} else {
			tok = newToken(BANG, l.currentChar)
		}
		break
	case '/':
		tok = newToken(SLASH, l.currentChar)
		break
	case '+':
		tok = newToken(PLUS, l.currentChar)
		break
	case '-':
		tok = newToken(MINUS, l.currentChar)
		break
	case '*':
		tok = newToken(ASTERISK, l.currentChar)
		break
	case '%':
		tok = newToken(PERCENT, l.currentChar)
		break
	case '^':
		tok = newToken(POWER, l.currentChar)
		break
	case '=':
		{
			ch := l.currentChar
			switch l.peekChar() {
			case '=':
				l.readChar()
				tok.Type = EQ
				tok.Literal = string(ch) + string(l.currentChar)
				break
			default:
				tok = newToken(ASSIGNMENT, l.currentChar)
			}
		}
		break
	case 0:
		tok.Literal = ""
		tok.Type = EOF
		break
	default:
		if isDigit(l.currentChar) {
			return Token{Type: INT, Literal: l.readNumber()}
		} else if isLetter(l.currentChar) {
			literal := l.readWord()
			ttype := LookupKeyword(literal)
			return Token{Type: ttype, Literal: literal}
		}
		tok = newToken(ILLEGAL, l.currentChar)
	}
	l.readChar()
	return tok
}

func (l *Lexer) consumeWhitespace() {
	for isWhiteSpace(l.currentChar) {
		l.readChar()
	}
}

func (l *Lexer) readNumber() string {
	pos := l.currentPosition
	for isDigit(l.currentChar) {
		l.readChar()
	}
	return l.input[pos:l.currentPosition]
}

func (l *Lexer) readWord() string {
	pos := l.currentPosition
	for isLetter(l.currentChar) {
		l.readChar()
	}
	return l.input[pos:l.currentPosition]
}

func (l *Lexer) readString(end byte) string {
	pos := l.currentPosition

	for l.currentChar != end {
		l.readChar()
	}

	return l.input[pos:l.currentPosition]
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
