package vtp

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/sonirico/visigoth/internal/search"
)

type Parser interface {
	ParseUInt8(io.Reader) (uint8, error)
	ParseUInt32(io.Reader) (uint32, error)
	ParseUInt64(io.Reader) (uint64, error)
	ParseString(io.Reader) (string, error)
	ParseText(io.Reader) (string, error)
	ParseByteType(io.Reader) (*ByteType, error)
	ParseUInt32Type(io.Reader) (*UInt32Type, error)
	ParseUInt64Type(io.Reader) (*UInt64Type, error)
	ParseStringType(io.Reader) (*StringType, error)
	ParseTextType(io.Reader) (*StringType, error)
	ParseLongTextType(io.Reader) (*StringType, error)
}

type BytesParser struct {
	endian binary.ByteOrder
}

func NewParser(endian binary.ByteOrder) *BytesParser {
	return &BytesParser{endian}
}

func (p *BytesParser) ParseUInt8(src io.Reader) (uint8, error) {
	data := make([]byte, 1, 1)
	if _, err := io.ReadFull(src, data); err != nil {
		return 0, err
	}
	return data[0], nil
}

func (p *BytesParser) ParseUInt32(src io.Reader) (uint32, error) {
	data := make([]byte, 4, 4)
	if _, err := io.ReadFull(src, data); err != nil {
		return 0, err
	}
	return p.endian.Uint32(data), nil
}

func (p *BytesParser) ParseUInt64(src io.Reader) (uint64, error) {
	data := make([]byte, 8, 8)
	if _, err := io.ReadFull(src, data); err != nil {
		return 0, err
	}
	return p.endian.Uint64(data), nil
}

func (p *BytesParser) ParseString(src io.Reader) (string, error) {
	l, err := p.ParseUInt8(src)
	if err != nil {
		return "", err
	}
	data := make([]byte, l, l)
	if _, err := io.ReadFull(src, data); err != nil {
		return "", err
	}
	return string(data), nil
}

func (p *BytesParser) ParseText(src io.Reader) (string, error) {
	l, err := p.ParseUInt32(src)
	if err != nil {
		return "", err
	}
	data := make([]byte, l, l)
	if _, err := io.ReadFull(src, data); err != nil {
		return "", err
	}
	return string(data), nil
}

func (p *BytesParser) ParseLongText(src io.Reader) (string, error) {
	l, err := p.ParseUInt64(src)
	if err != nil {
		return "", err
	}
	data := make([]byte, l, l)
	if _, err := io.ReadFull(src, data); err != nil {
		return "", err
	}
	return string(data), nil
}

func (p *BytesParser) ParseByteType(src io.Reader) (*ByteType, error) {
	val, err := p.ParseUInt8(src)
	return &ByteType{Value: val}, err
}

func (p *BytesParser) ParseUInt32Type(src io.Reader) (*UInt32Type, error) {
	val, err := p.ParseUInt32(src)
	return &UInt32Type{Value: val}, err
}

func (p *BytesParser) ParseUInt64Type(src io.Reader) (*UInt64Type, error) {
	val, err := p.ParseUInt64(src)
	return &UInt64Type{Value: val}, err
}

func (p *BytesParser) ParseStringType(src io.Reader) (*StringType, error) {
	str, err := p.ParseString(src)
	return &StringType{Value: str}, err
}

func (p *BytesParser) ParseTextType(src io.Reader) (*StringType, error) {
	str, err := p.ParseText(src)
	return &StringType{Value: str}, err
}

func (p *BytesParser) ParseLongTextType(src io.Reader) (*StringType, error) {
	str, err := p.ParseLongText(src)
	return &StringType{Value: str}, err
}

func ParseListIndicesResponse(
	head *Head,
	src io.Reader,
	parser Parser,
) (*ListIndicesResponse, error) {
	totalRead, err := parser.ParseUInt32(src)
	if err != nil {
		return nil, err
	}
	res := &ListIndicesResponse{Head: head}
	for totalRead > 0 {
		index, err := parser.ParseStringType(src)
		if err != nil {
			return nil, err
		}
		res.Indices = append(res.Indices, index)
		totalRead--
	}
	return res, nil
}

func ParseIndexRequest(head *Head, src io.Reader, parser Parser) (*IndexRequest, error) {
	format, err := parser.ParseByteType(src)
	if err != nil {
		return nil, err
	}
	index, err := parser.ParseStringType(src)
	if err != nil {
		return nil, err
	}
	doc, err := parser.ParseTextType(src)
	if err != nil {
		return nil, err
	}
	txt, err := parser.ParseTextType(src)
	if err != nil {
		return nil, err
	}
	req := &IndexRequest{Head: head, Format: format, Index: index, Doc: doc, Text: txt}
	return req, nil
}

func ParseSearchRequest(head *Head, src io.Reader, parser Parser) (*SearchRequest, error) {
	engine, err := parser.ParseByteType(src)
	if err != nil {
		return nil, err
	}
	index, err := parser.ParseStringType(src)
	if err != nil {
		return nil, err
	}
	terms, err := parser.ParseTextType(src)
	if err != nil {
		return nil, err
	}
	req := &SearchRequest{Head: head, EngineType: engine, Index: index, Terms: terms}
	return req, nil
}

func ParseSearchResponse(head *Head, src io.Reader, parser Parser) (*HitsSearchResponse, error) {
	engine, err := parser.ParseByteType(src)
	if err != nil {
		return nil, err
	}
	switch search.EngineType(engine.Value) {
	case search.SmartsHits:
		fallthrough
	case search.Hits, search.Linear:
		return ParseHitsSearchResponse(head, src, parser)
	default:
		return nil, fmt.Errorf("unknown engine type %d", engine.Value)
	}
}

func ParseHitsSearchResponse(
	head *Head,
	src io.Reader,
	parser Parser,
) (*HitsSearchResponse, error) {
	count, err := parser.ParseUInt32(src)
	if err != nil {
		return nil, err
	}
	documents := make([]*HitsResponseRow, count)
	var i uint32
	for i < count {
		hits, err := parser.ParseUInt32Type(src)
		if err != nil {
			return nil, err
		}
		name, err := parser.ParseTextType(src)
		if err != nil {
			return nil, err
		}
		content, err := parser.ParseLongTextType(src)
		if err != nil {
			return nil, err
		}
		documents[i] = &HitsResponseRow{
			Hits: hits,
			Document: &DocumentView{
				Name:    name,
				Content: content,
			},
		}
		i++
	}
	req := &HitsSearchResponse{SearchResponse: &SearchResponse{
		Head:   head,
		Engine: &ByteType{Value: uint8(search.Hits)},
	}, Documents: documents}
	return req, nil
}

func ParseAliasMessage(head *Head, src io.Reader, parser Parser) (*AliasRequest, error) {
	source, err := parser.ParseStringType(src)
	if err != nil {
		return nil, err
	}
	alias, err := parser.ParseStringType(src)
	if err != nil {
		return nil, err
	}
	req := &AliasRequest{Head: head, Source: source, Alias: alias}
	return req, nil
}

func ParseUnAliasMessage(head *Head, src io.Reader, parser Parser) (*UnAliasRequest, error) {
	index, err := parser.ParseStringType(src)
	if err != nil {
		return nil, err
	}
	alias, err := parser.ParseStringType(src)
	if err != nil {
		return nil, err
	}
	return &UnAliasRequest{Head: head, Index: index, Alias: alias}, nil
}

func ParseListIndicesRequest(head *Head) (*ListIndicesRequest, error) {
	return &ListIndicesRequest{Head: head}, nil
}

func ParseStatusResponse(head *Head, src io.Reader, parser Parser) (*StatusResponse, error) {
	ok, err := parser.ParseByteType(src)
	if err != nil {
		return nil, err
	}
	return &StatusResponse{Head: head, Ok: ok}, nil
}

func ParseDropIndexRequest(head *Head, src io.Reader, parser Parser) (*DropIndexRequest, error) {
	index, err := parser.ParseStringType(src)
	if err != nil {
		return nil, err
	}
	return &DropIndexRequest{Head: head, Index: index}, nil
}

func ParseDropIndexResponse(head *Head, src io.Reader, parser Parser) (*DropIndexResponse, error) {
	ok, err := parser.ParseByteType(src)
	if err != nil {
		return nil, err
	}
	index, err := parser.ParseStringType(src)
	if err != nil {
		return nil, err
	}
	return &DropIndexResponse{Head: head, Index: index, Ok: ok}, nil
}

func ParseListAliasesRequest(head *Head) (*ListAliasesRequest, error) {
	return &ListAliasesRequest{Head: head}, nil
}

func ParseListAliasesResponse(
	head *Head,
	src io.Reader,
	parser ProtoParser,
) (*ListAliasesResponse, error) {
	aliasesCount, err := parser.ParseUInt32(src)
	if err != nil {
		return nil, err
	}
	aliases := make([]*ListAliasesResponseRow, aliasesCount, aliasesCount)
	for aliasesCount > 0 {
		alias, err := parser.ParseIndexName(src)
		if err != nil {
			return nil, err
		}
		indicesCount, err := parser.ParseUInt8(src)
		if err != nil {
			return nil, err
		}
		indices := make([]*StringType, indicesCount, indicesCount)
		for indicesCount > 0 {
			index, err := parser.ParseIndexName(src)
			if err != nil {
				return nil, err
			}
			indices[indicesCount-1] = index
			indicesCount--
		}
		row := &ListAliasesResponseRow{
			Alias:   alias,
			Indices: indices,
		}
		aliases[aliasesCount-1] = row
		aliasesCount--
	}
	return &ListAliasesResponse{Head: head, Aliases: aliases}, nil
}

func ParseHead(src io.Reader, parser Parser) (*Head, error) {
	head := new(Head)
	id, err := parser.ParseUInt64Type(src)
	if err != nil {
		return nil, err
	}
	head.id = id
	version, err := parser.ParseByteType(src)
	if err != nil {
		return nil, err
	}
	head.version = version
	mType, err := parser.ParseByteType(src)
	if err != nil {
		return nil, err
	}
	head.messageType = mType
	return head, nil
}

func ParseBody(src io.Reader, head *Head, parser ProtoParser) (Message, error) {
	switch head.Type() {
	case StatusRes:
		return ParseStatusResponse(head, src, parser)
	case ListReq:
		return ParseListIndicesRequest(head)
	case ListRes:
		return ParseListIndicesResponse(head, src, parser)
	case AliasReq:
		return ParseAliasMessage(head, src, parser)
	case UnAliasReq:
		return ParseUnAliasMessage(head, src, parser)
	case SearchReq:
		return ParseSearchRequest(head, src, parser)
	case SearchRes:
		return ParseSearchResponse(head, src, parser)
	case IndexReq:
		return ParseIndexRequest(head, src, parser)
	case DropReq:
		return ParseDropIndexRequest(head, src, parser)
	case DropRes:
		return ParseDropIndexResponse(head, src, parser)
	case ListAliasesReq:
		return ParseListAliasesRequest(head)
	case ListAliasesRes:
		return ParseListAliasesResponse(head, src, parser)
	default:
		return nil, nil
	}
}

func Parse(src io.Reader, parser ProtoParser) (Message, error) {
	head, err := ParseHead(src, parser)
	if err != nil {
		return nil, err
	}
	return ParseBody(src, head, parser)
}

func ParseStream(
	ctx context.Context,
	src io.Reader,
	parser ProtoParser,
	queue chan<- Message,
) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			message, err := Parse(src, parser)

			if err != nil {
				return err
			}

			queue <- message
		}
	}
}

type ProtoParser interface {
	Parser
	Parse(io.Reader) (Message, error)
	ParseIndexName(s io.Reader) (*StringType, error)
}

type vtpParser struct {
	Parser
}

func NewVTPParser(parser Parser) ProtoParser {
	return &vtpParser{Parser: parser}
}

func (p *vtpParser) ParseIndexName(src io.Reader) (*StringType, error) {
	le, err := p.ParseUInt8(src)
	if err != nil {
		return nil, err
	}
	data := make([]byte, le, le)
	if _, err := io.ReadFull(src, data); err != nil {
		return nil, err
	}
	return &StringType{Value: string(data)}, nil
}

func (p *vtpParser) Parse(src io.Reader) (Message, error) {
	return Parse(src, p)
}
