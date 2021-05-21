package vtp

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/sonirico/visigoth/internal/search"
)

var (
	BigEndianCompiler = &BytesCompiler{
		endian: binary.BigEndian,
	}
	BigEndianParser = &BytesParser{
		endian: binary.BigEndian,
	}
	VTPCompiler = NewVTPCompiler(BigEndianCompiler)
	VTPParser   = NewVTPParser(BigEndianParser)
)

func testHeadEquals(t *testing.T, one, other Message) bool {
	t.Helper()

	if one.Type() != other.Type() {
		t.Errorf("message type mismatch")
		return false
	}

	if one.ID() != other.ID() {
		t.Errorf("ID mismatch")
		return false
	}

	if one.Version() != other.Version() {
		t.Errorf("version missmatch")
		return false
	}

	return true
}

func testUnAliasRequestEquals(t *testing.T, one, other *UnAliasRequest) bool {
	t.Helper()

	if strings.Compare(one.Index.Value, other.Index.Value) != 0 {
		t.Errorf("unAliasRequest: '%s' is not equal to '%s'",
			one.Alias.Value, other.Alias.Value)
		return false
	}

	if strings.Compare(one.Alias.Value, other.Alias.Value) != 0 {
		t.Errorf("unAliasRequest: '%s' is not equal to '%s'",
			one.Alias.Value, other.Alias.Value)
		return false
	}

	return true
}

func testAliasRequestEquals(t *testing.T, one, other *AliasRequest) bool {
	t.Helper()

	if strings.Compare(one.Source.Value, other.Source.Value) != 0 {
		return false
	}

	if strings.Compare(one.Alias.Value, other.Alias.Value) != 0 {
		return false
	}

	return true
}

func testSearchRequestEquals(t *testing.T, one, other *SearchRequest) bool {
	t.Helper()

	if one.EngineType.Value != other.EngineType.Value {
		return false
	}

	if strings.Compare(one.Index.Value, other.Index.Value) != 0 {
		return false
	}

	if strings.Compare(one.Terms.Value, other.Terms.Value) != 0 {
		return false
	}

	return true
}

func testPutRequestEquals(t *testing.T, one, other *IndexRequest) bool {
	t.Helper()

	if one.Format.Value != other.Format.Value {
		return false
	}

	if strings.Compare(one.Index.Value, other.Index.Value) != 0 {
		return false
	}

	if strings.Compare(one.Doc.Value, other.Doc.Value) != 0 {
		return false
	}

	if strings.Compare(one.Text.Value, other.Text.Value) != 0 {
		return false
	}

	return true
}

func testListRequestEquals(t *testing.T, one, other *ListIndicesRequest) bool {
	return true
}

func testListResponseEquals(t *testing.T, one, other *ListIndicesResponse) bool {
	t.Helper()

	if len(one.Indices) != len(other.Indices) {
		return false
	}
	for i, oneIndex := range one.Indices {
		otherIndex := other.Indices[i]
		if 0 != strings.Compare(oneIndex.Value, otherIndex.Value) {
			return false
		}
	}
	return true
}

func testDropIndexRequestEquals(t *testing.T, one, other *DropIndexRequest) bool {
	t.Helper()

	if one.Index.Value != other.Index.Value {
		t.Errorf("index name mismatch")
		return false
	}

	return true
}

func testDropIndexResponseEquals(t *testing.T, one, other *DropIndexResponse) bool {
	t.Helper()

	if one.Ok.Value != other.Ok.Value {
		t.Errorf("ok value mismatch")
		return false
	}

	if one.Index.Value != other.Index.Value {
		t.Errorf("index name mismatch")
		return false
	}

	return true
}

func testStatusResponseEquals(t *testing.T, one, other *StatusResponse) bool {
	t.Helper()

	if one.Ok.Value != other.Ok.Value {
		t.Errorf("ok value mismatch")
		return false
	}

	return true
}

func testHitsSearchResponse(t *testing.T, one, other *HitsSearchResponse) bool {
	t.Helper()

	if one.Engine.Value != other.Engine.Value {
		t.Errorf("search response engine type mismatch. %d vs %d",
			one.Engine.Value, other.Engine.Value)
		return false
	}

	if len(one.Documents) != len(other.Documents) {
		t.Errorf("search response document length mismatch. %d vs %d",
			len(one.Documents), len(other.Documents))
		return false
	}

	for i, oneDoc := range one.Documents {
		otherDoc := other.Documents[i]
		if oneDoc.Hits.Value != otherDoc.Hits.Value {
			t.Errorf("document hits mismatch. %d vs %d",
				oneDoc.Hits.Value, otherDoc.Hits.Value)
			return false
		}
		if oneDoc.Document.Name.Value != otherDoc.Document.Name.Value {
			t.Errorf("document name value mismatch. %s vs %s",
				oneDoc.Document.Name.Value, otherDoc.Document.Name.Value)
			return false
		}
		if oneDoc.Document.Content.Value != otherDoc.Document.Content.Value {
			t.Errorf("document content value mismatch. %s vs %s",
				oneDoc.Document.Content.Value, otherDoc.Document.Content.Value)
			return false
		}
	}
	return true
}

func testListAliasesResponseEquals(t *testing.T, one, other *ListAliasesResponse) bool {
	if len(one.Aliases) != len(other.Aliases) {
		t.Errorf("list aliases response items length mismatch. %d vs %d",
			len(one.Aliases), len(other.Aliases))
		return false
	}
	for i, expectedAlias := range one.Aliases {
		actualAlias := other.Aliases[i]

		if expectedAlias.Alias.Value != actualAlias.Alias.Value {
			t.Errorf("list aliases response. alias mismatch. want '%s', have '%s'",
				expectedAlias.Alias.Value, actualAlias.Alias.Value)
			return false
		}

		if len(expectedAlias.Indices) != len(actualAlias.Indices) {
			t.Errorf("list aliases response items length mismatch. %d vs %d",
				len(one.Aliases), len(other.Aliases))
			return false
		}
	}
	return true
}

func testMessageEquals(t *testing.T, one, other Message) bool {
	t.Helper()

	if !testHeadEquals(t, one, other) {
		t.Errorf("headers mismatch. %s is not equals to %s",
			one, other)
		return false
	}

	switch one.Type() {
	case ListAliasesRes:
		resOne := one.(*ListAliasesResponse)
		resOther := one.(*ListAliasesResponse)
		return testListAliasesResponseEquals(t, resOne, resOther)
	case StatusRes:
		statusRes := one.(*StatusResponse)
		statusOtherRes := other.(*StatusResponse)
		return testStatusResponseEquals(t, statusRes, statusOtherRes)
	case AliasReq:
		aliasOne := one.(*AliasRequest)
		aliasOther := other.(*AliasRequest)
		return testAliasRequestEquals(t, aliasOne, aliasOther)
	case UnAliasReq:
		unAliasOne := one.(*UnAliasRequest)
		unAliasOther := other.(*UnAliasRequest)
		return testUnAliasRequestEquals(t, unAliasOne, unAliasOther)
	case SearchReq:
		searchOne := one.(*SearchRequest)
		searchOther := other.(*SearchRequest)
		return testSearchRequestEquals(t, searchOne, searchOther)
	case IndexReq:
		putOne := one.(*IndexRequest)
		putOther := other.(*IndexRequest)
		return testPutRequestEquals(t, putOne, putOther)
	case ListReq:
		listOne := one.(*ListIndicesRequest)
		listOther := other.(*ListIndicesRequest)
		return testListRequestEquals(t, listOne, listOther)
	case ListRes:
		listOneResp := one.(*ListIndicesResponse)
		listOtherResp := other.(*ListIndicesResponse)
		return testListResponseEquals(t, listOneResp, listOtherResp)
	case DropReq:
		dropOneReq := one.(*DropIndexRequest)
		dropOtherReq := other.(*DropIndexRequest)
		return testDropIndexRequestEquals(t, dropOneReq, dropOtherReq)
	case DropRes:
		dropOneRes := one.(*DropIndexResponse)
		dropOtherRes := other.(*DropIndexResponse)
		return testDropIndexResponseEquals(t, dropOneRes, dropOtherRes)
	case SearchRes:
		switch searchOneResponse := one.(type) {
		case *HitsSearchResponse:
			searchOtherResponse := other.(*HitsSearchResponse)
			return testHitsSearchResponse(t, searchOneResponse, searchOtherResponse)
		}
	}

	return false
}

func TestCompile_BigEndian(t *testing.T) {
	tests := []struct {
		name         string
		message      Message
		expectedCode []byte
		compiler     ProtoCompiler
		parser       ProtoParser
	}{
		{
			name: "compile and parse list aliases response",
			message: &ListAliasesResponse{
				Head: &Head{
					id:          &UInt64Type{1},
					version:     &ByteType{2},
					messageType: MessageTypeToByte(ListAliasesRes),
				},
				Aliases: []*ListAliasesResponseRow{
					{Alias: &StringType{Value: "alias"}, Indices: []*StringType{
						{Value: "idx"},
						{Value: "idy"},
					}},
					{Alias: &StringType{Value: "alias2"}, Indices: []*StringType{
						{Value: "idz"},
					}},
				},
			},
			expectedCode: []byte{
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1,
				0x2,
				byte(ListAliasesRes),
				0x0, 0x0, 0x0, 0x2, // 2 rows
				0x5, // "alias" length
				byte('a'),
				byte('l'),
				byte('i'),
				byte('a'),
				byte('s'),
				0x2, // index count
				0x3, // "idx" length
				byte('i'),
				byte('d'),
				byte('x'),
				0x3, // "idy" length
				byte('i'),
				byte('d'),
				byte('y'),
				0x6, // "alias2" length
				byte('a'),
				byte('l'),
				byte('i'),
				byte('a'),
				byte('s'),
				byte('2'),
				0x1, // index count
				0x3, // "idz" length
				byte('i'),
				byte('d'),
				byte('z'),
			},
			compiler: VTPCompiler,
			parser:   VTPParser,
		},
		{
			name: "compile and parse status response",
			message: &StatusResponse{
				Head: &Head{
					id:          &UInt64Type{1},
					version:     &ByteType{2},
					messageType: MessageTypeToByte(StatusRes),
				},
				Ok: &ByteType{Value: 1},
			},
			expectedCode: []byte{
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1,
				0x2,
				byte(StatusRes),
				0x1,
			},
			compiler: VTPCompiler,
			parser:   VTPParser,
		},
		{
			name: "compile and parse drop index response",
			message: &DropIndexResponse{
				Head: &Head{
					id:          &UInt64Type{1},
					version:     &ByteType{2},
					messageType: MessageTypeToByte(DropRes),
				},
				Ok:    &ByteType{Value: 1},
				Index: &StringType{Value: "courses"},
			},
			expectedCode: []byte{
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1,
				0x2,
				byte(DropRes),
				0x1,
				0x7, byte('c'), byte('o'), byte('u'), byte('r'), byte('s'), byte('e'), byte('s'),
			},
			compiler: VTPCompiler,
			parser:   VTPParser,
		},
		{
			name: "compile and parse drop index request",
			message: &DropIndexRequest{
				Head: &Head{
					id:          &UInt64Type{1},
					version:     &ByteType{2},
					messageType: MessageTypeToByte(DropReq),
				},
				Index: &StringType{Value: "courses"},
			},
			expectedCode: []byte{
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1,
				0x2,
				byte(DropReq),
				0x7, byte('c'), byte('o'), byte('u'), byte('r'), byte('s'), byte('e'), byte('s'),
			},
			compiler: VTPCompiler,
			parser:   VTPParser,
		},
		{
			name: "compile and parse search hits response",
			message: &HitsSearchResponse{
				SearchResponse: &SearchResponse{
					Head: &Head{
						id:          &UInt64Type{1},
						version:     &ByteType{2},
						messageType: MessageTypeToByte(SearchRes),
					},
					Engine: &ByteType{Value: uint8(search.Hits)},
				},
				Documents: []*HitsResponseRow{
					{
						Document: &DocumentView{
							Name:    &StringType{Value: "name"},
							Content: &StringType{Value: "content"},
						},
						Hits: &UInt32Type{2},
					},
				},
			},
			expectedCode: []byte{
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1,
				0x2,
				byte(SearchRes),
				0x2,                 // engine
				0x0, 0x0, 0x0, 0x01, // result count
				// row 1
				0x0, 0x0, 0x0, 0x02, // hits
				0x0, 0x0, 0x0, 0x04, // document name length
				byte('n'), byte('a'), byte('m'), byte('e'), // document name
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x07, // document content length
				byte('c'), byte('o'), byte('n'), byte('t'), byte('e'), byte('n'), byte('t'), // document content
				// end row 1
			},
			compiler: VTPCompiler,
			parser:   VTPParser,
		},
		{
			name: "compile and parse list indices response",
			message: &ListIndicesResponse{
				Head: &Head{
					id:          &UInt64Type{1},
					version:     &ByteType{2},
					messageType: MessageTypeToByte(ListRes),
				},
				Indices: []*StringType{
					{"courses"},
				},
			},
			expectedCode: []byte{
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1,
				0x2,
				byte(ListRes),
				0x0, 0x0, 0x0, 0x01, // string length
				0x7, byte('c'), byte('o'), byte('u'), byte('r'), byte('s'), byte('e'), byte('s'),
			},
			compiler: VTPCompiler,
			parser:   VTPParser,
		},
		{
			name: "compile and parse list indices request",
			message: &ListIndicesRequest{
				Head: &Head{
					id:          &UInt64Type{1},
					version:     &ByteType{2},
					messageType: MessageTypeToByte(ListReq),
				},
			},
			expectedCode: []byte{
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1,
				0x2,
				byte(ListReq),
			},
			compiler: VTPCompiler,
			parser:   VTPParser,
		},
		{
			name: "compile and parse unalias request",
			message: &UnAliasRequest{
				Head: &Head{
					id:          &UInt64Type{0},
					version:     &ByteType{0},
					messageType: MessageTypeToByte(UnAliasReq),
				},
				Index: &StringType{"index"},
				Alias: &StringType{"aliases"},
			},
			expectedCode: []byte{
				0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
				0x0,
				byte(UnAliasReq),
				0x5, byte('i'), byte('n'), byte('d'), byte('e'), byte('x'),
				0x7, byte('a'), byte('l'), byte('i'), byte('a'), byte('s'), byte('e'), byte('s'),
			},
			compiler: VTPCompiler,
			parser:   VTPParser,
		},
		{
			name: "compile and parse index request",
			message: &IndexRequest{
				Head: &Head{
					id:          &UInt64Type{(1<<64 - 1) - 3},
					version:     &ByteType{2},
					messageType: MessageTypeToByte(IndexReq),
				},
				Format: &ByteType{Value: 1},
				Index:  &StringType{"verbos"},
				Doc:    &StringType{"hinco"},
				Text:   &StringType{"accion de hincar"},
			},
			expectedCode: []byte{
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFC, // ID
				0x2,            // Version
				byte(IndexReq), // MessageType
				0x01,
				0x06, byte('v'), byte('e'), byte('r'), byte('b'), byte('o'), byte('s'),
				0x00, 0x00, 0x00, 0x05, byte('h'), byte('i'), byte('n'), byte('c'), byte('o'),
				0x00, 0x00, 0x00, 0x10,
				byte('a'), byte('c'), byte('c'), byte('i'), byte('o'), byte('n'),
				byte(' '), byte('d'), byte('e'), byte(' '),
				byte('h'), byte('i'), byte('n'), byte('c'), byte('a'), byte('r'),
			},
			compiler: VTPCompiler,
			parser:   VTPParser,
		},
		{
			name: "compile and parse search request",
			message: &SearchRequest{
				Head: &Head{
					id:          &UInt64Type{(1<<64 - 1) - 2},
					version:     &ByteType{1},
					messageType: MessageTypeToByte(SearchReq),
				},
				EngineType: &ByteType{uint8(search.NoopAll)},
				Index:      &StringType{Value: "index"},
				Terms:      &StringType{Value: "hope"},
			},
			expectedCode: []byte{
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFD, // ID
				0x1,             // Version
				byte(SearchReq), // MessageType
				byte(search.NoopAll),
				0x05, byte('i'), byte('n'), byte('d'), byte('e'), byte('x'),
				0x00, 0x00, 0x00, 0x04,
				byte('h'), byte('o'), byte('p'), byte('e'),
			},
			compiler: VTPCompiler,
			parser:   VTPParser,
		},
		{
			name: "compile and parse alias request",
			message: &AliasRequest{
				Head: &Head{
					id:          &UInt64Type{(1<<64 - 1) - 1},
					version:     &ByteType{1},
					messageType: MessageTypeToByte(AliasReq),
				},
				Source: &StringType{"index_name"},
				Alias:  &StringType{"alias"},
			},
			expectedCode: []byte{
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE, // ID
				0x1,            // Version
				byte(AliasReq), // MessageType
				0x0A, byte('i'), byte('n'), byte('d'), byte('e'), byte('x'), byte('_'), byte('n'), byte('a'), byte('m'), byte('e'),
				0x05, byte('a'), byte('l'), byte('i'), byte('a'), byte('s'),
			},
			compiler: VTPCompiler,
			parser:   VTPParser,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Test serialization
			buf := new(bytes.Buffer)
			if err := test.compiler.Compile(buf, test.message); err != nil {
				t.Fatalf("error when parsing: %s", err.Error())
			}

			if !bytes.Equal(test.expectedCode, buf.Bytes()) {
				t.Fatalf("unexpected encoded value. \nwant: %x\nhave: %x",
					test.expectedCode, buf.Bytes())
			}

			// Test deserialization
			message, err := test.parser.Parse(buf)
			if err != nil {
				t.Fatalf("unexpected error %s", err)
			}

			if !testMessageEquals(t, test.message, message) {
				t.Errorf("unexpected parsed result. want %T(%+v), have %T(%+v)",
					test.message, test.message, message, message)
			}

			// Another round-trip to check that parsed compiles against to test sampling
			buf.Reset()
			if err := test.compiler.Compile(buf, test.message); err != nil {
				t.Errorf("unexpected re-encoded value. \nwant: %x\nhave: %x",
					test.expectedCode, buf.Bytes())
			}
		})
	}
}
