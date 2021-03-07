package analyze

import "strings"

type Tokenizer interface {
	Tokenize(text string) []string
}

type Filter interface {
	Filter(tokens []string) []string
}

type LowerCaseFilter struct{}

func (l LowerCaseFilter) Filter(tokens []string) []string {
	r := make([]string, len(tokens), len(tokens))
	for i, token := range tokens {
		r[i] = strings.ToLower(token)
	}
	return r
}

func NewLowerCaseTokenizer() LowerCaseFilter {
	return LowerCaseFilter{}
}
