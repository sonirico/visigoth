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

type commandEvaluator struct {
	counter *atomicCounter
	env     *environment
}

func newCommandEvaluator(env *environment) *commandEvaluator {
	return &commandEvaluator{
		counter: &atomicCounter{},
		env:     env,
	}
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
		if eval := c.evalStatement(stmt); eval != nil {
			msgs = append(msgs, eval)
		}
	}
	return msgs
}

func (c *commandEvaluator) evalStatement(node vql.Node) vtp.Message {
	switch s := node.(type) {
	case *vql.SearchStatement:
		return c.evalSearchStatement(s)
	case *vql.ShowStatement:
		return c.evalShowStatement(s)
	case *vql.UseStatement:
		return c.evalUseStatement(s)
	case *vql.IndexStatement:
		return c.evalIndexStatement(s)
	case *vql.AliasStatement:
		return c.evalAliasStatement(s)
	case *vql.UnAliasStatement:
		return c.evalUnAliasStatement(s)
	default:
		return nil
	}
}

func (c *commandEvaluator) evalUnAliasStatement(node *vql.UnAliasStatement) vtp.Message {
	var indexName string
	if node.Index != nil {
		indexName = node.Index.Literal()
	}
	return vtp.NewUnAliasRequest(c.nextID(), Version, indexName, node.Alias.Literal())
}

func (c *commandEvaluator) evalAliasStatement(node *vql.AliasStatement) vtp.Message {
	return vtp.NewAliasRequest(c.nextID(), Version, node.Index.Literal(), node.Alias.Literal())
}

func (c *commandEvaluator) evalIndexStatement(node *vql.IndexStatement) vtp.Message {
	format := entities.MimeText
	if strings.ToLower(node.Format.Literal()) == "json" {
		format = entities.MimeJSON
	}
	return vtp.NewIndexRequest(
		c.nextID(),
		Version,
		node.Index.Literal(),
		node.Aka.Literal(),
		node.Payload.Literal(),
		format,
	)
}

func (c *commandEvaluator) evalUseStatement(node *vql.UseStatement) vtp.Message {
	currentIndex := node.Used.Literal()
	c.env.Index = &currentIndex
	return nil
}

func (c *commandEvaluator) evalShowStatement(q *vql.ShowStatement) vtp.Message {
	switch q.Shown.Literal() {
	case "aliases":
		return vtp.NewListAliasesRequest(c.nextID(), Version)
	case "indices", "indexes":
		fallthrough
	default:
		return vtp.NewListIndicesRequest(c.nextID(), Version)
	}
}

func (c *commandEvaluator) evalSearchStatement(node *vql.SearchStatement) vtp.Message {
	var index string
	var engine search.EngineType
	if node.Index == nil {
		if c.env.Index == nil {
			log.Println("SEARCH command requires a payload")
			return nil
		}
		index = *c.env.Index
	} else {
		index = node.Index.Literal()
	}
	if node.Engine == nil {
		engine = search.Hits
	} else {
		switch strings.ToLower(node.Engine.Literal()) {
		case "hits":
			engine = search.Hits
		case "smart_hits":
			engine = search.SmartsHits
		case "noop":
			engine = search.NoopZero
		case "noop_all":
			engine = search.NoopAll
		case "linear":
			engine = search.Linear
		default:
			engine = search.Hits
		}
	}
	return vtp.NewSearchRequest(
		c.nextID(),
		Version,
		uint8(engine),
		index,
		node.Payload.Literal(),
	)
}

func (c *commandEvaluator) nextID() uint64 {
	return c.counter.Inc()
}
