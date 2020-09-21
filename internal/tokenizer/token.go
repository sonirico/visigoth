package tokenizer

import (
	"bytes"
	"golang.org/x/text/runes"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type Tokenizer interface {
	Tokenize([]byte) [][]byte
	TokenizeSingle([]byte) ([]byte, bool)
}

type Transformer interface {
	Transform(payload []byte) ([]byte, error)
}

type lowercaseTransformer struct {
	Tf transform.Transformer
}

type IsMarkNonSpacingChecker struct{}

func (c *IsMarkNonSpacingChecker) Contains(r rune) bool {
	return isMarkNonSpacing(r)
}

var runeRemover = &IsMarkNonSpacingChecker{}

func newLowercaseTransformer() *lowercaseTransformer {
	tf := &lowercaseTransformer{}
	tf.Tf = transform.Chain(norm.NFD, runes.Remove(runeRemover), norm.NFC)
	return tf
}

func (t *lowercaseTransformer) Transform(payload []byte) ([]byte, error) {
	final := make([]byte, len(payload))
	nDst, _, err := t.Tf.Transform(final, bytes.ToLowerSpecial(unicode.CaseRanges, payload), true)
	if err != nil {
		return nil, err
	}
	final = final[:nDst]
	return final, nil
}

type SimpleTokenizer struct {
	Tf Transformer
}

func NewSimpleTokenizer() *SimpleTokenizer {
	st := &SimpleTokenizer{}
	st.Tf = newLowercaseTransformer()
	return st
}

func (s *SimpleTokenizer) Tokenize(payload []byte) [][]byte {
	var res [][]byte
	for _, bword := range bytes.Fields(payload) {
		bword, ok := s.TokenizeSingle(bword)
		if ok {
			res = append(res, bword)
		}
	}
	return res
}

func (s *SimpleTokenizer) TokenizeSingle(payload []byte) ([]byte, bool) {
	bword := bytes.Trim(payload, ".,-~?!\"'`;:()<>[]{}\\|/=_+*&^%$#@")
	if len(bword) > 0 {
		word, err := s.Tf.Transform(bword)
		if err != nil {
			panic(err)
		}
		return word, true
	}
	return nil, false
}

type StopWordsTokenizer struct {
	T         *SimpleTokenizer
	stopWords map[string]bool
}

func NewStopWordsTokenizer(stopWords []string) *StopWordsTokenizer {
	base := NewSimpleTokenizer()
	t := &StopWordsTokenizer{
		T:         base,
		stopWords: make(map[string]bool, len(stopWords)),
	}
	for _, w := range stopWords {
		stopWordToken, ok := base.TokenizeSingle([]byte(w))
		if ok {
			t.stopWords[string(stopWordToken)] = true
		}
	}
	return t
}

func (st *StopWordsTokenizer) Tokenize(payload []byte) [][]byte {
	var res [][]byte
	for _, bword := range st.T.Tokenize(payload) {
		bword, ok := st.TokenizeSingle(bword)
		if ok {
			res = append(res, bword)
		}
	}
	return res
}

func (st *StopWordsTokenizer) TokenizeSingle(payload []byte) ([]byte, bool) {
	sword := string(payload)
	if ok := st.stopWords[sword]; ok {
		return nil, false
	}
	return payload, true
}

var (
	SpanishTokenizer = NewStopWordsTokenizer(SpanishStopWords)
)
