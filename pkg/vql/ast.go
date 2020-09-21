package vql

import (
	"bytes"
)

type Node interface {
	Literal() string

	String() string
}

type Statement interface {
	Node

	statementNode()
}

type Expression interface {
	Node

	expressionNode()
}

// IdentifierTokenType expression
type Identifier struct {
	Token Token
	Value string
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) Literal() string {
	return i.Value
}
func (i *Identifier) String() string {
	return i.Value
}

// STRING literal
type StringLiteral struct {
	Token Token

	Value string
}

func (sl *StringLiteral) expressionNode() {}
func (sl *StringLiteral) Literal() string {
	return sl.Value
}
func (sl *StringLiteral) String() string {
	return sl.Value
}

// USE statement
type UseStatement struct {
	Token Token

	Used Expression
}

func (i *UseStatement) statementNode() {}
func (i *UseStatement) Literal() string {
	return i.Token.Literal
}
func (i *UseStatement) String() string {
	return i.Used.String()
}

// SHOW statement
type ShowStatement struct {
	Token Token

	Shown Expression
}

func (i *ShowStatement) statementNode() {}
func (i *ShowStatement) Literal() string {
	return i.Token.Literal
}
func (i *ShowStatement) String() string {
	return i.Shown.String()
}

// SearchTokenType statement
type SearchStatement struct {
	Token   Token
	Index   Expression
	Payload Expression
	Engine  Expression
}

func (i *SearchStatement) statementNode() {}
func (i *SearchStatement) Literal() string {
	return i.Token.Literal
}
func (i *SearchStatement) String() string {
	// todo
	return i.Payload.Literal()
}

// INDEX statement
type IndexStatement struct {
	Token   Token
	Index   Expression
	Payload Expression
}

func (i *IndexStatement) statementNode() {}
func (i *IndexStatement) Literal() string {
	return i.Token.Literal
}
func (i *IndexStatement) String() string {
	// todo
	return i.Payload.Literal()
}

// DROP statement
type DropStatement struct {
	Token  Token
	Target Expression
}

func (i *DropStatement) statementNode() {}
func (i *DropStatement) Literal() string {
	return i.Token.Literal
}
func (i *DropStatement) String() string {
	// todo
	return i.Target.Literal()
}

// QUERY statement
type Query struct {
	Statements []Statement
}

func (q *Query) Literal() string {
	if len(q.Statements) > 0 {
		return q.Statements[0].Literal()
	} else {
		return ""
	}
}
func (q *Query) String() string {
	var buffer bytes.Buffer

	for _, stmt := range q.Statements {
		buffer.WriteString(stmt.String())
	}

	return buffer.String()
}
