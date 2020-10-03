package client

import (
	"github.com/sonirico/visigoth/pkg/entities"
	"log"
	"strings"
	"sync/atomic"

	"github.com/sonirico/visigoth/internal/search"

	"github.com/sonirico/visigoth/pkg/vql"

	"github.com/sonirico/visigoth/pkg/vtp"
)

type atomicCounter struct {
	Value int64
}

func (a *atomicCounter) Inc() uint64 {
	atomic.AddInt64(&a.Value, 1)
	return uint64(a.Value)
}

type Evaluator interface {
	Eval(string) []vtp.Message
}

type evalFunc func(vql.Node) vtp.Message

type commandEvaluator struct {
	counter       *atomicCounter
	evalFunctions map[vql.TokenType]evalFunc
	env           *environment
}

func newCommandEvaluator(env *environment) *commandEvaluator {
	cp := &commandEvaluator{
		counter:       &atomicCounter{},
		evalFunctions: make(map[vql.TokenType]evalFunc),
		env:           env,
	}
	cp.evalFunctions[vql.SearchTokenType] = cp.evalSearchStatement
	cp.evalFunctions[vql.ShowTokenType] = cp.evalShowIndicesStatement
	cp.evalFunctions[vql.UseTokenType] = cp.evalUseStatement
	cp.evalFunctions[vql.IndexTokenType] = cp.evalIndexStatement
	return cp
}

func (c *commandEvaluator) Eval(raw string) []vtp.Message {
	var msgs []vtp.Message
	lexer := vql.NewLexer(raw)
	parser := vql.NewParser(lexer)
	query := parser.ParseQuery()
	if errors := parser.Errors(); len(errors) > 0 {
		log.Println(errors)
		return msgs
	}

	for _, stmt := range query.Statements {
		eval, ok := c.evalFunctions[vql.TokenType(stmt.Literal())]
		var msg vtp.Message
		if ok {
			msg = eval(stmt)
		}

		if msg != nil {
			msgs = append(msgs, msg)
		}
	}
	return msgs
}

func (c *commandEvaluator) evalIndexStatement(node vql.Node) vtp.Message {
	q, _ := node.(*vql.IndexStatement)
	format := entities.MimeText
	if strings.ToLower(q.Format.Literal()) == "json" {
		format = entities.MimeJSON
	}
	return vtp.NewIndexRequest(c.counter.Inc(), Version, q.Index.Literal(), q.Aka.Literal(), q.Payload.Literal(), format)
}

func (c *commandEvaluator) evalUseStatement(node vql.Node) vtp.Message {
	q, _ := node.(*vql.UseStatement)
	currentIndex := q.Used.Literal()
	c.env.Index = &currentIndex
	return nil
}

func (c *commandEvaluator) evalShowIndicesStatement(q vql.Node) vtp.Message {
	return vtp.NewListIndicesRequest(c.counter.Inc(), Version)
}

func (c *commandEvaluator) evalSearchStatement(node vql.Node) vtp.Message {
	q, _ := node.(*vql.SearchStatement)
	var index string
	var engine search.EngineType
	if q.Index == nil {
		if c.env.Index == nil {
			log.Println("SEARCH command requires a payload")
			return nil
		}
		index = *c.env.Index
	} else {
		index = q.Index.Literal()
	}
	if q.Engine == nil {
		engine = search.Hits
	} else {
		switch strings.ToLower(q.Engine.Literal()) {
		case "hits":
			engine = search.Hits
		case "smart_hits":
			engine = search.SmartsHits
		case "noop":
			engine = search.NoopZero
		case "noop_all":
			engine = search.NoopAll
		}
	}
	return vtp.NewSearchRequest(
		c.counter.Inc(),
		Version,
		uint8(engine),
		index,
		q.Payload.Literal(),
	)
}
