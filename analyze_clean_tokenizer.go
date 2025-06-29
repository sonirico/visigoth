package visigoth

import (
	"strings"
	"unicode"
)

type cleanFunc func(r rune) bool

type CleanTokenizer struct {
	fns []cleanFunc
}

func (c *CleanTokenizer) register(fn cleanFunc) {
	c.fns = append(c.fns, fn)
}

func (c *CleanTokenizer) Tokenize(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		for _, fn := range c.fns {
			if fn(r) {
				return false
			}
		}
		return true
	})
}

func NewCleanTokenizer(fns ...cleanFunc) CleanTokenizer {
	ct := CleanTokenizer{}
	for _, fn := range fns {
		ct.register(fn)
	}
	return ct
}

func NewKeepAlphanumericTokenizer() *CleanTokenizer {
	ct := new(CleanTokenizer)
	ct.register(func(r rune) bool {
		return unicode.IsNumber(r) || unicode.IsLetter(r)
	})
	return ct
}
