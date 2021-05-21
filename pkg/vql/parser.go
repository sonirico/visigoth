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
	peekToken, err := p.lexer.NextToken()
	if err != nil {
		p.addError(err.Error())
	}
	p.peekToken = peekToken
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
	}

	p.peekError(Type)
	return false
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
	case AliasTokenType:
		return p.parseAliasStatement()
	case UnAliasTokenType:
		return p.parseUnAliasStatement()
	default:
		return nil
	}
}

func (p *Parser) parseUnAliasStatement() Statement {
	unalias := &UnAliasStatement{
		Token: p.currentToken,
		Index: nil,
		Alias: nil,
	}
	unalias.Index = p.parseStringLiteralOrIdentifier(true)
	if unalias.Index == nil {
		return nil
	}
	if p.peekTokenIs(AsTokenType) {
		p.nextToken()
	}
	unalias.Alias = p.parseStringLiteralOrIdentifier(false)
	if unalias.Alias == nil {
		// If there is no second parameter, the first parameter
		// is considered to be the alias to be totally removed
		unalias.Alias = unalias.Index
		unalias.Index = nil
	}

	return unalias
}

func (p *Parser) parseAliasStatement() Statement {
	alias := &AliasStatement{
		Token: p.currentToken,
		Index: nil,
		Alias: nil,
	}
	alias.Index = p.parseStringLiteralOrIdentifier(true)
	if alias.Index == nil {
		return nil
	}
	if p.peekTokenIs(AsTokenType) {
		p.nextToken()
	}
	alias.Alias = p.parseStringLiteralOrIdentifier(true)
	if alias.Alias == nil {
		return nil
	}

	return alias
}

func (p *Parser) parseShowStatement() Statement {
	st := &ShowStatement{
		Token: p.currentToken,
		Shown: p.parseStringLiteralOrIdentifier(true),
	}

	if st.Shown == nil {
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

	if st.Used = p.parseStringLiteralOrIdentifier(true); st.Used == nil {
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
		if st.Index = p.parseStringLiteralOrIdentifier(true); st.Index == nil {
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

func (p *Parser) parseStringLiteralOrIdentifier(force bool) Expression {
	if p.peekTokenIs(STRING) {
		p.nextToken()
		return p.parseStringLiteral()
	} else if p.peekTokenIs(IdentifierTokenType) {
		p.nextToken()
		return p.parseIdentifierExpression()
	} else if p.peekToken.IsKeyword() {
		// literals that collide with keyword may pass as identifiers
		p.nextToken()
		return p.parseIdentifierExpression()
	} else {
		if force {
			p.addError("unexpected end of input. want literal or string, got %s", p.peekToken)
		}
		return nil
	}
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
