package vtp

import "errors"

var (
	UnknownMessageType     = errors.New("unknown message type")
	responseMessageTypeMap = map[MessageType]MessageType{
		ListReq:    ListRes,
		SearchReq:  SearchRes,
		IndexReq:   StatusRes,
		AliasReq:   StatusRes,
		UnAliasReq: StatusRes,
		DropReq:    DropRes,
	}
)

func LookupResponseMessageType(req MessageType) (MessageType, error) {
	msgType, ok := responseMessageTypeMap[req]
	if !ok {
		return 0, UnknownMessageType
	}
	return msgType, nil
}

type ListIndicesResponse struct {
	*Head
	Indices []*StringType
}

type DocumentView struct {
	Name    *StringType
	Content *StringType
}

type SearchResponse struct {
	*Head
	Engine    *ByteType
	Documents []*StringType
}

type HitsResponseRow struct {
	Document *DocumentView
	Hits     *UInt32Type
}

type HitsSearchResponse struct {
	*SearchResponse

	Documents []*HitsResponseRow
}

type StatusResponse struct {
	*Head

	Ok *ByteType
}

type DropIndexResponse struct {
	*Head

	Ok    *ByteType
	Index *StringType
}

func (d *DropIndexResponse) IsOk() bool { return d.Ok.Value == 1 }

type ListAliasesResponseRow struct {
	Alias   *StringType
	Indices []*StringType
}

type ListAliasesResponse struct {
	*Head

	Aliases []*ListAliasesResponseRow
}

func NewHeadResponse(req Message) *Head {
	msgType, err := LookupResponseMessageType(req.Type())
	if err != nil {
		panic(err)
	}
	return &Head{
		id:          &UInt64Type{Value: req.Id()},
		version:     &ByteType{Value: req.Version()}, // Sure?
		messageType: &ByteType{Value: uint8(msgType)},
	}
}
