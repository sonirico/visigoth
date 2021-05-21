package index

import (
	"bytes"
	"fmt"

	"github.com/sonirico/visigoth/pkg/entities"

	"github.com/sonirico/visigoth/internal/search"
)

//easyjson:json
type MemoryIndex struct {
	name string

	tokenizer tokenizer

	Docs          []entities.Doc   `json:"indexed"`
	InvertedIndex map[string][]int `json:"inverted"`
}

func (mi *MemoryIndex) Len() int {
	return len(mi.Docs)
}

func (mi *MemoryIndex) Indexed(key string) (data []int) {
	data, _ = mi.InvertedIndex[key]
	r := append(make([]int, 0, len(data)), data...)
	return r
}

func (mi *MemoryIndex) Document(index int) entities.Doc {
	return mi.Docs[index]
}

func (mi *MemoryIndex) String() string {
	var buf bytes.Buffer
	buf.WriteString("{\n")
	buf.WriteString("\tname=" + mi.name)
	for token, indexed := range mi.InvertedIndex {
		buf.WriteString(fmt.Sprintf("\n\t[token=%s,length=(%d)]=%v", token, len(indexed), indexed))
	}
	buf.WriteString("\n}")
	return buf.String()
}

func (mi *MemoryIndex) Put(payload indexable) Index {
	tokens := mi.tokenizer.Tokenize(payload.Statement())
	next := len(mi.Docs)
	newDoc := entities.NewDoc(payload.ID(), payload.Raw())
	mi.Docs = append(mi.Docs, newDoc)
tokenLoop:
	for _, tok := range tokens {
		indexedDocs := mi.InvertedIndex[tok]
		for _, docIndex := range indexedDocs {
			if docIndex == next {
				continue tokenLoop
			}
		}
		mi.InvertedIndex[tok] = append(indexedDocs, next)
	}
	return mi
}

func (mi *MemoryIndex) Search(payload string, engine search.Engine) entities.Iterator {
	return engine(mi.tokenizer.Tokenize(payload), mi)
}

func NewMemoryIndex(name string, tkr tokenizer) *MemoryIndex {
	return &MemoryIndex{
		name:          name,
		tokenizer:     tkr,
		Docs:          []entities.Doc{},
		InvertedIndex: make(map[string][]int),
	}
}

func NewMemoryIndexBuilder(tokenizer tokenizer) Builder {
	return func(name string) Index {
		return NewMemoryIndex(name, tokenizer)
	}
}
