package visigoth

import "github.com/sonirico/vago/slices"

type tokenizer interface {
	Tokenize(text string) []string
}

type Index interface {
	Put(payload DocRequest) Index
	Search(terms string, engine Engine) slices.Slice[SearchResult]
}

type Builder func(name string) Index
