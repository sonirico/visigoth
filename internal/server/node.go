package server

import (
	"bytes"
	"log"
	"strings"

	"github.com/sonirico/visigoth/internal/loaders"
	"github.com/sonirico/visigoth/pkg/entities"

	"github.com/sonirico/visigoth/internal/repos"
	"github.com/sonirico/visigoth/internal/search"
	"github.com/sonirico/visigoth/pkg/vtp"
)

type tracer func(message vtp.Message)

func simpleTracer(message vtp.Message) {
	log.Println(vtp.MessageToString(message))
}

type NodeConfig struct {
	tracer tracer
}

// Node is the internal server that dispatches requests in vtp format
type Node interface {
	Run(in chan vtp.Message, out chan vtp.Message, cfg *NodeConfig)
}

var (
	jsonLoader = loaders.NewJSONLoader(false)
	textLoader = loaders.NewTextLoader()
)

type node struct {
	repo repos.IndexRepo
}

func NewNode(repo repos.IndexRepo) Node {
	return &node{repo: repo}
}

func (n *node) Run(in chan vtp.Message, out chan vtp.Message, cfg *NodeConfig) {
	var tracer tracer
	if cfg != nil && cfg.tracer != nil {
		tracer = cfg.tracer
	} else {
		tracer = simpleTracer
	}

	for req := range in {
		tracer(req)
		res := n.dispatch(req)
		out <- res
	}
}

func (n *node) dispatch(req vtp.Message) vtp.Message {
	switch req.Type() {
	case vtp.ListAliasesReq:
		return n.handleListAliasesRequest(req)
	case vtp.ListReq:
		return n.handleListIndicesRequest(req)
	case vtp.SearchReq:
		return n.handleSearchRequest(req)
	case vtp.IndexReq:
		return n.handleIndexRequest(req)
	case vtp.UnAliasReq:
		return n.handleUnAliasRequest(req)
	case vtp.AliasReq:
		return n.handleAliasRequest(req)
	case vtp.DropReq:
		return n.handleDropRequest(req)
	default:
		return req
	}
}

func (n *node) handleAliasRequest(msg vtp.Message) *vtp.StatusResponse {
	req := msg.(*vtp.AliasRequest)
	ok := n.repo.Alias(req.Alias.Value, req.Source.Value)
	return vtp.NewStatusResponse(msg.Id(), msg.Version(), ok)
}

func (n *node) handleUnAliasRequest(msg vtp.Message) *vtp.StatusResponse {
	req := msg.(*vtp.UnAliasRequest)
	ok := n.repo.UnAlias(req.Alias.Value, req.Index.Value)
	return vtp.NewStatusResponse(msg.Id(), msg.Version(), ok)
}

func (n *node) handleDropRequest(msg vtp.Message) *vtp.DropIndexResponse {
	req := msg.(*vtp.DropIndexRequest)
	ok := n.repo.Drop(req.Index.Value)
	return vtp.NewDropIndexResponse(msg.Id(), msg.Version(), ok, req.Index.Value)
}

func (n *node) handleIndexRequest(msg vtp.Message) *vtp.StatusResponse {
	req := msg.(*vtp.IndexRequest)
	var statement string
	in := bytes.NewBufferString(req.Text.Value)
	out := bytes.NewBuffer(nil)
	switch entities.MimeType(req.Format.Value) {
	case entities.MimeJSON:
		if err := jsonLoader.Load(in, out); err != nil {
			log.Println(err)
			return nil
		}
	case entities.MimeText:
		if err := textLoader.Load(in, out); err != nil {
			log.Println(err)
			return nil
		}
	}
	statement = out.String()
	doc := entities.NewDocRequestWith(req.Doc.Value, req.Text.Value, statement)
	n.repo.Put(req.Index.Value, doc)
	return &vtp.StatusResponse{Head: vtp.NewHeadResponse(req), Ok: &vtp.ByteType{Value: 1}}
}

func (n *node) handleListIndicesRequest(msg vtp.Message) *vtp.ListIndicesResponse {
	indices := n.repo.List()
	typedIndices := make([]*vtp.StringType, len(indices))
	for i, index := range indices {
		typedIndices[i] = &vtp.StringType{
			Value: index,
		}
	}
	return &vtp.ListIndicesResponse{
		Head:    vtp.NewHeadResponse(msg),
		Indices: typedIndices,
	}
}

func (n *node) handleSearchRequest(msg vtp.Message) vtp.Message {
	req, _ := msg.(*vtp.SearchRequest)
	var engine search.Engine

	index := strings.TrimSpace(req.Index.Value)
	terms := strings.TrimSpace(req.Terms.Value)

	switch search.EngineType(req.EngineType.Value) {
	case search.Hits:
		engine = search.HitsSearchEngine
	default:
		return vtp.NewStatusResponse(req.Id(), req.Version(), false)
	}

	results, err := n.repo.Search(index, terms, engine)

	if err != nil {
		log.Println(err)
		return vtp.NewStatusResponse(req.Id(), req.Version(), false)
	}

	res := &vtp.HitsSearchResponse{
		SearchResponse: &vtp.SearchResponse{
			Head:   vtp.NewHeadResponse(msg),
			Engine: req.EngineType.Clone(),
		},
	}

	for {
		row, done := results.Next()

		if row != nil {
			switch mrow := row.(type) {
			case search.HitsSearchRow:
				doc := &vtp.HitsResponseRow{
					Document: &vtp.DocumentView{
						Name:    &vtp.StringType{Value: mrow.Doc().Id()},
						Content: &vtp.StringType{Value: mrow.Doc().Raw()},
					},
					Hits: &vtp.UInt32Type{Value: uint32(mrow.Hits())},
				}
				res.Documents = append(res.Documents, doc)
			}
		}

		if done {
			break
		}
	}

	return res
}

func (n *node) handleListAliasesRequest(msg vtp.Message) *vtp.ListAliasesResponse {
	aliases := n.repo.ListAliases()
	vtpAliases := make([]*vtp.ListAliasesResponseRow, len(aliases.Aliases), len(aliases.Aliases))
	for i, alias := range aliases.Aliases {
		vtpIndices := make([]*vtp.StringType, len(alias.Indices))
		for j, index := range alias.Indices {
			vtpIndices[j] = &vtp.StringType{Value: index}
		}

		vtpAliases[i] = &vtp.ListAliasesResponseRow{
			Alias:   &vtp.StringType{Value: alias.Alias},
			Indices: vtpIndices,
		}
	}

	return &vtp.ListAliasesResponse{
		Head:    vtp.NewHeadResponse(msg),
		Aliases: vtpAliases,
	}
}
