package tokenizer

import (
	"bytes"
	"testing"
)

func b(payload string) []byte {
	return []byte(payload)
}

func assertTokenResult(t *testing.T, actual, expected [][]byte) bool {
	t.Helper()

	if len(actual) != len(expected) {
		t.Errorf("unexpected token result length. want %d elements, have %d", len(expected), len(actual))
		return false
	}

	for index, eitem := range expected {
		aitem := actual[index]
		if 0 != bytes.Compare(aitem, eitem) {
			t.Errorf("unexpected token literal, want '%s', have '%s'", string(eitem), string(aitem))
			t.Errorf("unexpected token literal, want '%x', have '%x'", eitem, aitem)
			return false
		}
	}

	return true
}

func Test_SimpleTokenizer_Tokenize_ascii(t *testing.T) {
	tkr := NewSimpleTokenizer()

	payload := b(`This is a simple sentence`)
	expected := [][]byte{
		b("this"),
		b("is"),
		b("a"),
		b("simple"),
		b("sentence"),
	}
	actual := tkr.TokenizeText(payload)

	if !assertTokenResult(t, actual, expected) {
		t.FailNow()
	}
}

func Test_SimpleTokenizer_Tokenize_utf8(t *testing.T) {
	tkr := NewSimpleTokenizer()

	payload := b(`El ñu corría por la (SABANA)`)
	expected := [][]byte{
		b("el"),
		b("nu"),
		b("corria"),
		b("por"),
		b("la"),
		b("sabana"),
	}
	actual := tkr.TokenizeText(payload)

	if !assertTokenResult(t, actual, expected) {
		t.FailNow()
	}
}

func Test_StopWordsTokenizer_Tokenize_utf8(t *testing.T) {
	stopWords := []string{"el", "la", "por"}
	tkr := NewStopWordsTokenizerFilter(stopWords, NewSimpleTokenizer())

	payload := b(`El ñu corría por la SABANA`)
	expected := [][]byte{
		b("nu"),
		b("corria"),
		b("sabana"),
	}
	actual := tkr.TokenizeText(payload)

	if !assertTokenResult(t, actual, expected) {
		t.FailNow()
	}
}
