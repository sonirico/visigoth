package search

import (
	"encoding/json"
	"log"

	"github.com/sonirico/visigoth/pkg/entities"
)

type (
	defaultSearchSchema struct {
		DocID string      `json:"_id"`
		Doc   interface{} `json:"_doc"`
	}

	hitsSearchRowSchema struct {
		DocID string                 `json:"_id"`
		Doc   map[string]interface{} `json:"_doc"`
		Hits  int                    `json:"hits"`
	}
)

type jsonSearchResultSerializer struct{}

func (j *jsonSearchResultSerializer) Serialize(item entities.Row) []byte {
	var res interface{}

	switch row := item.(type) {
	case hitsSearchResultRow:
		doc := make(map[string]interface{})
		_ = json.Unmarshal([]byte(row.Doc().Raw()), &doc)
		res = hitsSearchRowSchema{
			DocID: row.Doc().ID(),
			Doc:   doc,
			Hits:  row.Hits(),
		}
	default:
		doc := make(map[string]interface{})
		_ = json.Unmarshal([]byte(row.Doc().Raw()), &doc)
		res = defaultSearchSchema{
			DocID: item.Doc().ID(),
			Doc:   doc,
		}
	}

	raw, err := json.Marshal(res)
	if err != nil {
		log.Println("error:", err)
		return nil
	}
	return raw
}

var JSONHitsSearchResultSerializer = jsonSearchResultSerializer{}
