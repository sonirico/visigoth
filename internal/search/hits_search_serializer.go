package search

import (
	"encoding/json"
	"github.com/sonirico/visigoth/pkg/entities"
	"log"
)

type hitsSearchRowSchema struct {
	Doc  entities.Doc `json:"doc"`
	Hits int          `json:"hits"`
}

type jsonHitsSearchResultSerializer struct{}

func (j *jsonHitsSearchResultSerializer) Serialize(item entities.Row) []byte {
	row, ok := item.(HitsSearchRow)
	if !ok {
		log.Fatal("unexpected type cannot be serialized. want 'hitsSearchRow', have %V", row)
	}
	data := &hitsSearchRowSchema{
		Doc:  row.Doc(),
		Hits: row.Hits(),
	}
	raw, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	return raw
}

var JsonHitsSearchResultSerializer = &jsonHitsSearchResultSerializer{}
