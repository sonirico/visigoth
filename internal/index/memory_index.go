package index

import (
	"github.com/sonirico/visigoth/pkg/entities"
	"sync"

	"github.com/sonirico/visigoth/internal/search"

	visigoth "github.com/sonirico/visigoth/internal/tokenizer"
)

type Indexable interface {
	ID() string
	Raw() string // The original content of the payload
	Mime() entities.MimeType
	Statement() string
}

type Index interface {
	Put(Doc Indexable) Index
	Search(terms string, engine search.Engine) entities.Iterator
}

type MemoryIndex struct {
	L         sync.RWMutex
	Name      string
	Tokenizer visigoth.Tokenizer
	indexed   []entities.Doc
	inverted  map[string][]int
}

func NewMemoryIndex(name string, tkr visigoth.Tokenizer) *MemoryIndex {
	return &MemoryIndex{
		Name:      name,
		Tokenizer: tkr,
		indexed:   []entities.Doc{},
		inverted:  make(map[string][]int),
	}
}

func (mi *MemoryIndex) Len() int {
	return len(mi.indexed)
}

func (mi *MemoryIndex) Indexed(key string) (data []int) {
	data, _ = mi.inverted[key]
	return
}

func (mi *MemoryIndex) Document(index int) entities.Doc {
	Doc := mi.indexed[index]
	return Doc
}

func (mi *MemoryIndex) String() string {
	return mi.Name
}

func (mi *MemoryIndex) Put(payload Indexable) Index {
	mi.L.Lock()
	defer mi.L.Unlock()
	tokens := mi.Tokenizer.Tokenize([]byte(payload.Statement()))
	next := len(mi.indexed)
	newDoc := entities.NewDoc(payload.ID(), payload.Raw())
	mi.indexed = append(mi.indexed, newDoc)
tokenLoop:
	for _, tok := range tokens {
		tokStr := string(tok)
		indexedDocs := mi.inverted[tokStr]
		for _, docIndex := range indexedDocs {
			if docIndex == next {
				continue tokenLoop
			}
		}
		mi.inverted[tokStr] = append(indexedDocs, next)
	}

	return mi
}

func (mi *MemoryIndex) Search(payload string, engine search.Engine) entities.Iterator {
	mi.L.RLock()
	defer mi.L.RUnlock()
	tokens := mi.Tokenizer.Tokenize([]byte(payload))
	return engine(tokens, mi)
}