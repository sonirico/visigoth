package tokenizer

import (
	"unicode"
)

func isMarkNonSpacing(r rune) bool {
	return unicode.Is(unicode.Mn, r)
}
