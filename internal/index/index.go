package index

import (
	"github.com/sonirico/visigoth/internal/search"
	"github.com/sonirico/visigoth/pkg/entities"
)

type tokenizer interface {
	Tokenize(text string) []string
}

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

type IndexBuilder func(name string) Index
