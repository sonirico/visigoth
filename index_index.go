package visigoth

import "github.com/sonirico/vago/slices"

type tokenizer interface {
	Tokenize(text string) []string
}

type indexable interface {
	ID() string
	Raw() string // The original content of the payload
	Mime() MimeType
	Statement() string
}

type Index interface {
	Put(Doc indexable) Index
	Search(terms string, engine Engine) slices.Slice[SearchResult]
}

type Builder func(name string) Index
