package search

import (
	"encoding/json"
	"github.com/sonirico/visigoth/pkg/entities"
	"log"
)

type hitsSearchRowSchema struct {
	DocId string                 `json:"_id"`
	Doc   map[string]interface{} `json:"_doc"`
	Hits  int                    `json:"hits"`
}

type jsonHitsSearchResultSerializer struct{}

func (j *jsonHitsSearchResultSerializer) Serialize(item entities.Row) []byte {
	row, ok := item.(HitsSearchRow)
	if !ok {
		log.Fatal("unexpected type cannot be serialized. want 'hitsSearchRow', have %V", row)
	}
	doc := make(map[string]interface{})
	_ = json.Unmarshal([]byte(row.Doc().Raw()), &doc)
	data := &hitsSearchRowSchema{
		DocId: row.Doc().Id(),
		Doc:   doc,
		Hits:  row.Hits(),
	}
	raw, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	return raw
}

var JsonHitsSearchResultSerializer = &jsonHitsSearchResultSerializer{}
