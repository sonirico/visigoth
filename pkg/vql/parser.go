package vql

import (
	"fmt"
)

type Parser struct {
	lexer *Lexer

	errors []string

	currentToken Token
	peekToken    Token
}

func NewParser(lexer *Lexer) *Parser {
	parser := &Parser{
		lexer:  lexer,
		errors: []string{},
	}
	// Read to  so as to have initialised both currentToken and peekToken
	parser.nextToken()
	parser.nextToken()

	return parser
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) currentTokenIs(Type TokenType) bool {
	return p.currentToken.Type == Type
}

func (p *Parser) peekTokenIs(Type TokenType) bool {
	return p.peekToken.Type == Type
}

func (p *Parser) peekError(Type TokenType) {
	msg := fmt.Sprintf("Expected next  to be of type '%s'. Got '%s' -> %s",
		Type, p.peekToken.Type, p.peekToken.Literal)
	p.addError(msg)
}

func (p *Parser) addError(placeholder string, params ...interface{}) {
	p.errors = append(p.errors, fmt.Sprintf(placeholder, params...))
}

func (p *Parser) expectPeekToken(Type TokenType) bool {
	if p.peekTokenIs(Type) {
		p.nextToken()
		return true
	} else {
		p.peekError(Type)
		return false
	}
}

func (p *Parser) parseStatement() Statement {
	switch p.currentToken.Type {
	case SearchTokenType:
		return p.parseSearchStatement()
	case IndexTokenType:
		return p.parseIndexStatement()
	case UseTokenType:
		return p.parseUseStatement()
	case ShowTokenType:
		return p.parseShowStatement()
	case DropTokenType:
		return p.parseDropStatement()
	default:
		return nil
	}
}

func (p *Parser) parseShowStatement() Statement {
	st := &ShowStatement{
		Token: p.currentToken,
		Shown: nil,
	}
	if p.peekTokenIs(IdentifierTokenType) {
		p.nextToken()
		st.Shown = p.parseIdentifierExpression()
	} else {
		p.peekError(IdentifierTokenType)
		return nil
	}

	return st
}

func (p *Parser) parseDropStatement() Statement {
	st := &DropStatement{
		Token:  p.currentToken,
		Target: nil,
	}
	if p.peekTokenIs(IdentifierTokenType) {
		p.nextToken()
		st.Target = p.parseIdentifierExpression()
	} else if p.peekTokenIs(STRING) {
		p.nextToken()
		st.Target = p.parseStringLiteral()
	} else {
		p.peekError(IdentifierTokenType)
		return nil
	}

	return st
}

func (p *Parser) parseUseStatement() Statement {
	st := &UseStatement{
		Token: p.currentToken,
		Used:  nil,
	}
	if p.peekTokenIs(IdentifierTokenType) {
		p.nextToken()
		st.Used = p.parseIdentifierExpression()
	} else if p.peekTokenIs(STRING) {
		p.nextToken()
		st.Used = p.parseStringLiteral()
	} else {
		p.peekError(IdentifierTokenType)
		return nil
	}

	return st
}

func (p *Parser) parseIndexStatement() Statement {
	st := &IndexStatement{
		Token:   p.currentToken,
		Index:   &StringLiteral{Value: "", Token: *NewToken(STRING, "")},
		Aka:     &StringLiteral{Value: "", Token: *NewToken(STRING, "")},
		Format:  &StringLiteral{Value: "TEXT", Token: *NewToken(STRING, "")},
		Payload: nil,
	}

	if !p.expectPeekToken(STRING) {
		return nil
	}

	st.Payload = p.parseStringLiteral()

	if p.peekTokenIs(AkaTokenType) {
		p.nextToken()
		if !p.expectPeekToken(STRING) {
			return nil
		}
		st.Aka = p.parseIdentifierExpression()
	}

	if p.peekTokenIs(AsTokenType) {
		p.nextToken()
		if !p.expectPeekToken(IdentifierTokenType) {
			return nil
		}
		st.Format = p.parseIdentifierExpression()
	}

	if p.peekTokenIs(IntoTokenType) {
		p.nextToken()
		if p.peekTokenIs(STRING) {
			p.nextToken()
			st.Index = p.parseStringLiteral()
		} else if p.peekTokenIs(IdentifierTokenType) {
			p.nextToken()
			st.Index = p.parseIdentifierExpression()
		} else {
			p.addError("unexpected end of input. want literal or string, got %s", p.peekToken)
			return nil
		}
	}

	return st
}

func (p *Parser) parseSearchStatement() Statement {
	st := &SearchStatement{
		Token:   p.currentToken,
		Index:   nil, // Can be nil to allow clients to set it based on environment
		Payload: nil,
		Engine:  nil,
	}
	p.nextToken()
	if p.currentTokenIs(IdentifierTokenType) {
		st.Index = p.parseIdentifierExpression()
	} else if p.currentTokenIs(STRING) {
		st.Index = p.parseStringLiteral()
	}
	if !p.peekTokenIs(STRING) {
		// If there is no next string payload, the index is considered as such. Index will be the one stored
		// in the environment, if any
		st.Payload = st.Index
		st.Index = nil
	} else {
		p.nextToken()
		st.Payload = p.parseStringLiteral()
	}

	if p.peekTokenIs(UsingTokenType) {
		p.nextToken()

		if p.peekTokenIs(IdentifierTokenType) {
			p.nextToken()
			st.Engine = p.parseIdentifierExpression()
		}
	}

	return st
}

func (p *Parser) parseIdentifierExpression() Expression {
	return &Identifier{Token: p.currentToken, Value: p.currentToken.Literal}
}

func (p *Parser) parseStringLiteral() Expression {
	return &StringLiteral{Value: p.currentToken.Literal}
}

func (p *Parser) ParseQuery() *Query {
	program := &Query{Statements: []Statement{}}

	for p.currentToken.Type != EOF {
		statement := p.parseStatement()
		if statement != nil {
			program.Statements = append(program.Statements, statement)
		}
		p.nextToken()
	}

	return program
}
