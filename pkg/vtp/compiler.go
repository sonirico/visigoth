package vtp

import (
	"encoding/binary"
	"io"
)

const (
	Version = 0
)

// TODO: Split interface
type Compiler interface {
	CompileUInt8(io.Writer, uint8) error
	CompileUInt32(io.Writer, uint32) error
	CompileUInt64(io.Writer, uint64) error
	CompileString(io.Writer, string) error
	CompileByteType(io.Writer, *ByteType) error
	CompileUInt32Type(io.Writer, *UInt32Type) error
	CompileUInt64Type(io.Writer, *UInt64Type) error
	CompileMessageType(io.Writer, MessageType) error
	CompileStringType(io.Writer, *StringType) error
	CompileVarcharType(io.Writer, *StringType) error
	CompileBlobType(io.Writer, *BlobType) error
}

type compiler struct {
	endian binary.ByteOrder
}

func NewCompiler(endian binary.ByteOrder) *compiler {
	return &compiler{endian}
}

func (c *compiler) CompileUInt8(w io.Writer, u8 uint8) error {
	if _, err := w.Write([]byte{u8}); err != nil {
		return err
	}
	return nil
}

func (c *compiler) CompileUInt32(w io.Writer, u32 uint32) error {
	data := make([]byte, 4)
	c.endian.PutUint32(data, u32)
	if _, err := w.Write(data); err != nil {
		return err
	}
	return nil
}

func (c *compiler) CompileUInt64(w io.Writer, u64 uint64) error {
	data := make([]byte, 8)
	c.endian.PutUint64(data, u64)
	if _, err := w.Write(data); err != nil {
		return err
	}
	return nil
}

func (c *compiler) CompileString(w io.Writer, s string) error {
	if _, err := w.Write([]byte(s)); err != nil {
		return err
	}
	return nil
}

func (c *compiler) CompileByteType(w io.Writer, bt *ByteType) error {
	return c.CompileUInt8(w, bt.Value)
}

func (c *compiler) CompileUInt32Type(w io.Writer, u32 *UInt32Type) error {
	return c.CompileUInt32(w, u32.Value)
}

func (c *compiler) CompileUInt64Type(w io.Writer, u64 *UInt64Type) error {
	return c.CompileUInt64(w, u64.Value)
}

func (c *compiler) CompileMessageType(w io.Writer, mt MessageType) error {
	return c.CompileUInt8(w, uint8(mt))
}

func (c *compiler) CompileStringType(w io.Writer, s *StringType) error {
	if _, err := w.Write([]byte(s.Value)); err != nil {
		return err
	}
	return nil
}

func (c *compiler) CompileVarcharType(w io.Writer, s *StringType) error {
	if err := c.CompileUInt32(w, uint32(len(s.Value))); err != nil {
		return err
	}

	if _, err := w.Write([]byte(s.Value)); err != nil {
		return err
	}
	return nil
}

func (c *compiler) CompileBlobType(w io.Writer, b *BlobType) error {
	if _, err := w.Write(b.Value); err != nil {
		return err
	}
	return nil
}

func (c *compiler) CompileHead(w io.Writer, m Message, comp Compiler) error {
	return compileHead(w, m, comp)
}

func compileHead(w io.Writer, m Message, comp Compiler) error {
	if err := comp.CompileUInt64(w, m.Id()); err != nil {
		return err
	}
	if err := comp.CompileUInt8(w, m.Version()); err != nil {
		return err
	}
	if err := comp.CompileMessageType(w, m.Type()); err != nil {
		return err
	}
	return nil
}

func compileAliasRequest(w io.Writer, alias *AliasRequest, comp Compiler) error {
	if err := comp.CompileUInt8(w, uint8(alias.Source.Len())); err != nil {
		return err
	}

	if err := comp.CompileStringType(w, alias.Source); err != nil {
		return err
	}

	if err := comp.CompileUInt8(w, uint8(alias.Alias.Len())); err != nil {
		return err
	}

	if err := comp.CompileStringType(w, alias.Alias); err != nil {
		return err
	}

	return nil
}

func compileSearchRequest(w io.Writer, req *SearchRequest, comp Compiler) error {
	if err := comp.CompileByteType(w, req.EngineType); err != nil {
		return err
	}

	if err := comp.CompileUInt8(w, uint8(req.Index.Len())); err != nil {
		return err
	}

	if err := comp.CompileStringType(w, req.Index); err != nil {
		return err
	}

	if err := comp.CompileUInt32(w, uint32(req.Terms.Len())); err != nil {
		return err
	}

	if err := comp.CompileStringType(w, req.Terms); err != nil {
		return err
	}

	return nil
}

func compileIndexRequest(w io.Writer, req *IndexRequest, comp Compiler) error {
	if err := comp.CompileByteType(w, req.Format); err != nil {
		return err
	}

	if err := comp.CompileUInt8(w, uint8(req.Index.Len())); err != nil {
		return err
	}

	if err := comp.CompileStringType(w, req.Index); err != nil {
		return err
	}

	if err := comp.CompileUInt32(w, uint32(req.Doc.Len())); err != nil {
		return err
	}

	if err := comp.CompileStringType(w, req.Doc); err != nil {
		return err
	}

	if err := comp.CompileUInt32(w, uint32(req.Text.Len())); err != nil {
		return err
	}

	if err := comp.CompileStringType(w, req.Text); err != nil {
		return err
	}

	return nil
}

func compileUnAliasRequest(w io.Writer, req *UnAliasRequest, comp Compiler) error {
	if err := comp.CompileUInt8(w, uint8(req.Index.Len())); err != nil {
		return err
	}
	if err := comp.CompileStringType(w, req.Index); err != nil {
		return err
	}
	if err := comp.CompileUInt8(w, uint8(req.Alias.Len())); err != nil {
		return err
	}
	if err := comp.CompileStringType(w, req.Alias); err != nil {
		return err
	}
	return nil
}

func compileListIndicesRequest(w io.Writer, req *ListIndicesRequest, comp Compiler) error {
	// Nothing more to write
	return nil
}

func compileListIndicesResponse(w io.Writer, res *ListIndicesResponse, comp Compiler) error {
	if err := comp.CompileUInt32(w, uint32(len(res.Indices))); err != nil {
		return err
	}
	for _, indice := range res.Indices {
		if err := comp.CompileUInt8(w, uint8(indice.Len())); err != nil {
			return err
		}

		if err := comp.CompileStringType(w, indice); err != nil {
			return err
		}
	}
	return nil
}

func compileBlobRequest(w io.Writer, req *BlobRequest, comp Compiler) error {
	if err := comp.CompileByteType(w, req.Algo); err != nil {
		return err
	}
	if err := comp.CompileUInt32(w, uint32(req.Blob.Len())); err != nil {
		return err
	}
	if err := comp.CompileBlobType(w, req.Blob); err != nil {
		return err
	}
	return nil
}

func compileSearchResponse(w io.Writer, req *SearchResponse, comp Compiler) error {
	if err := comp.CompileByteType(w, req.Engine); err != nil {
		return err
	}
	return nil
}

func compileHitsSearchResponse(w io.Writer, req *HitsSearchResponse, comp Compiler) error {
	if err := compileSearchResponse(w, req.SearchResponse, comp); err != nil {
		return err
	}
	if err := comp.CompileUInt32(w, uint32(len(req.Documents))); err != nil {
		return err
	}

	for _, doc := range req.Documents {
		if err := comp.CompileUInt32Type(w, doc.Hits); err != nil {
			return err
		}
		if err := comp.CompileUInt32(w, uint32(doc.Document.Name.Len())); err != nil {
			return err
		}
		if err := comp.CompileStringType(w, doc.Document.Name); err != nil {
			return err
		}
		if err := comp.CompileUInt64(w, uint64(doc.Document.Content.Len())); err != nil {
			return err
		}
		if err := comp.CompileStringType(w, doc.Document.Content); err != nil {
			return err
		}
	}
	return nil
}

func compileStatusResponse(w io.Writer, res *StatusResponse, comp Compiler) error {
	if err := comp.CompileByteType(w, res.Ok); err != nil {
		return err
	}
	return nil
}

func compileDropIndexRequest(w io.Writer, req *DropIndexRequest, comp Compiler) error {
	if err := comp.CompileUInt8(w, uint8(req.Index.Len())); err != nil {
		return err
	}
	if err := comp.CompileStringType(w, req.Index); err != nil {
		return err
	}
	return nil
}

func compileDropIndexResponse(w io.Writer, res *DropIndexResponse, comp Compiler) error {
	if err := comp.CompileByteType(w, res.Ok); err != nil {
		return err
	}
	if err := comp.CompileUInt8(w, uint8(res.Index.Len())); err != nil {
		return err
	}
	if err := comp.CompileStringType(w, res.Index); err != nil {
		return err
	}
	return nil
}

func compileListAliasesRequest(w io.Writer, req *ListAliasesRequest, comp Compiler) error {
	return nil
}

func compileListAliasesResponse(w io.Writer, res *ListAliasesResponse, comp ProtoCompiler) error {
	if err := comp.CompileUInt32(w, uint32(len(res.Aliases))); err != nil {
		return err
	}
	for _, item := range res.Aliases {
		if err := comp.CompileIndexName(w, item.Alias); err != nil {
			return err
		}
		if err := comp.CompileUInt8(w, uint8(len(item.Indices))); err != nil {
			return err
		}
		for _, ind := range item.Indices {
			if err := comp.CompileIndexName(w, ind); err != nil {
				return err
			}
		}
	}
	return nil
}

func Compile(w io.Writer, msg Message, c ProtoCompiler) error {
	if err := compileHead(w, msg, c); err != nil {
		return err
	}
	switch val := msg.(type) {
	case *ListIndicesRequest:
		return compileListIndicesRequest(w, val, c)
	case *ListIndicesResponse:
		return compileListIndicesResponse(w, val, c)
	case *UnAliasRequest:
		return compileUnAliasRequest(w, val, c)
	case *AliasRequest:
		return compileAliasRequest(w, val, c)
	case *SearchRequest:
		return compileSearchRequest(w, val, c)
	case *IndexRequest:
		return compileIndexRequest(w, val, c)
	case *BlobRequest:
		return compileBlobRequest(w, val, c)
	case *HitsSearchResponse:
		return compileHitsSearchResponse(w, val, c)
	case *StatusResponse:
		return compileStatusResponse(w, val, c)
	case *DropIndexRequest:
		return compileDropIndexRequest(w, val, c)
	case *DropIndexResponse:
		return compileDropIndexResponse(w, val, c)
	case *ListAliasesRequest:
		return compileListAliasesRequest(w, val, c)
	case *ListAliasesResponse:
		return compileListAliasesResponse(w, val, c)
	default:
		return nil
	}
}

type ProtoCompiler interface {
	Compiler
	Compile(io.Writer, Message) error
	CompileIndexName(w io.Writer, name *StringType) error
}

type vtpCompiler struct {
	Compiler
}

func (v *vtpCompiler) CompileIndexName(w io.Writer, index *StringType) error {
	if err := v.CompileUInt8(w, byte(index.Len())); err != nil {
		return err
	}
	return v.CompileString(w, index.Value)
}

func (v *vtpCompiler) Compile(w io.Writer, msg Message) error {
	return Compile(w, msg, v)
}

func NewVTPCompiler(comp Compiler) *vtpCompiler {
	return &vtpCompiler{
		Compiler: comp,
	}
}
