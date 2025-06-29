package visigoth

import snowballSpanish "github.com/kljensen/snowball/spanish"

type SpanishStemmerFilter struct {
	removeStopWords bool
}

func (s SpanishStemmerFilter) Filter(tokens []string) []string {
	r := make([]string, 0, len(tokens))
	for _, token := range tokens {
		r = append(r, snowballSpanish.Stem(token, s.removeStopWords))
	}
	return r
}

func NewSpanishStemmer(removeStopWords bool) SpanishStemmerFilter {
	return SpanishStemmerFilter{removeStopWords: removeStopWords}
}
