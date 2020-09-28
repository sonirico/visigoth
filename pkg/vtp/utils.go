package vtp

import (
	"fmt"
	"github.com/sonirico/visigoth/pkg/entities"
)

var (
	messageTypeName = map[MessageType]string{
		AliasReq:   "ALIAS",
		UnAliasReq: "UNALIAS",
		ListReq:    "LIST",
		ListRes:    "LIST",
		IndexReq:   "INDEX",
		SearchReq:  "SEARCH",
		SearchRes:  "SEARCH",
		DropReq:    "DROP",
		DropRes:    "DROP",
		RenameReq:  "RENAME",
	}
)

func MessageToString(m Message) string {
	name, _ := messageTypeName[m.Type()]
	return fmt.Sprintf("request{type=%s,id=%d,version=%d}", name, m.Id(), m.Version())
}

func MessageTypeToByte(m MessageType) *ByteType {
	return &ByteType{Value: uint8(m)}
}

func NewListIndicesRequest(id uint64, version uint8) *ListIndicesRequest {
	return &ListIndicesRequest{
		&Head{
			id:          &UInt64Type{id},
			version:     &ByteType{version},
			messageType: MessageTypeToByte(ListReq),
		},
	}
}

func NewIndexRequest(id uint64, version uint8, index, name, payload string, format entities.MimeType) *IndexRequest {
	return &IndexRequest{
		Head: &Head{
			id:          &UInt64Type{id},
			version:     &ByteType{version},
			messageType: MessageTypeToByte(IndexReq),
		},
		Format: &ByteType{Value: uint8(format)},
		Index:  &StringType{Value: index},
		Doc:    &StringType{Value: name},
		Text:   &StringType{Value: payload},
	}
}

func NewSearchRequest(id uint64, version, engine uint8, index, terms string) *SearchRequest {
	return &SearchRequest{
		Head: &Head{
			id:          &UInt64Type{id},
			version:     &ByteType{version},
			messageType: MessageTypeToByte(SearchReq),
		},
		EngineType: &ByteType{engine},
		Index:      &StringType{index},
		Terms:      &StringType{terms},
	}
}

func NewAliasRequest(id uint64, version uint8, index, alias string) *AliasRequest {
	return &AliasRequest{
		Head: &Head{
			id:          &UInt64Type{id},
			version:     &ByteType{version},
			messageType: MessageTypeToByte(AliasReq),
		},
		Source: &StringType{index},
		Alias:  &StringType{alias},
	}
}

func NewUnAliasRequest(id uint64, version uint8, alias string) *UnAliasRequest {
	return &UnAliasRequest{
		Head: &Head{
			id:          &UInt64Type{id},
			version:     &ByteType{version},
			messageType: MessageTypeToByte(UnAliasReq),
		},
		Alias: &StringType{alias},
	}
}

func NewDropIndexRequest(id uint64, version uint8, index string) *DropIndexRequest {
	return &DropIndexRequest{
		Head: &Head{
			id:          &UInt64Type{id},
			version:     &ByteType{version},
			messageType: MessageTypeToByte(DropReq),
		},
		Index: &StringType{Value: index},
	}
}

func NewDropIndexResponse(id uint64, version uint8, ok bool, index string) *DropIndexResponse {
	var okVal uint8 = 0
	if ok {
		okVal = 1
	}
	return &DropIndexResponse{
		Head: &Head{
			id:          &UInt64Type{id},
			version:     &ByteType{version},
			messageType: MessageTypeToByte(DropRes),
		},
		Ok:    &ByteType{Value: okVal},
		Index: &StringType{Value: index},
	}
}

func NewStatusResponse(id uint64, version uint8, ok bool) *StatusResponse {
	var okVal uint8 = 0
	if ok {
		okVal = 1
	}
	return &StatusResponse{
		Head: &Head{
			id:          &UInt64Type{id},
			version:     &ByteType{version},
			messageType: MessageTypeToByte(StatusRes),
		},
		Ok: &ByteType{Value: okVal},
	}
}
