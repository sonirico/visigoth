package vtp

import (
	"bytes"
	"fmt"
)

type (
	MessageType uint8
)

const (
	StatusRes MessageType = iota + 1
	AliasReq
	IndexReq
	SearchReq
	SearchRes
	RenameReq
	DropReq
	DropRes
	ListReq
	ListRes
	UnAliasReq
)

var HeadLength = 8 + 1 + 1

type Message interface {
	String() string
	Id() uint64
	Version() uint8
	Type() MessageType
}

type Head struct {
	id          *UInt64Type
	version     *ByteType
	messageType *ByteType
}

func (h Head) Id() uint64 {
	return h.id.Value
}

func (h Head) Version() uint8 {
	return h.version.Value
}

func (h Head) Type() MessageType {
	return MessageType(h.messageType.Value)
}

func (h Head) String() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("id: %d\n", h.id))
	buf.WriteString(fmt.Sprintf("version: %d\n", h.version))
	buf.WriteString(fmt.Sprintf("type: %d\n", h.messageType))
	return buf.String()
}

type AliasRequest struct {
	*Head

	Source *StringType
	Alias  *StringType
}

type UnAliasRequest struct {
	*Head
	Alias *StringType
}

type SearchRequest struct {
	*Head
	EngineType *ByteType
	Index      *StringType
	Terms      *StringType
}

type IndexRequest struct {
	*Head
	Format *ByteType   // u8
	Index  *StringType // u8
	Doc    *StringType // u32
	Text   *StringType // u32
}

type ListIndicesRequest struct {
	*Head
}

type BlobRequest struct {
	*Head
	Algo *ByteType
	Blob *BlobType
}

type DropIndexRequest struct {
	*Head
	Index *StringType
}
