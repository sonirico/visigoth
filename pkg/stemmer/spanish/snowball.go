package spanish

var vowels = map[rune]struct{}{
	'a': {},
	'e': {},
	'i': {},
	'o': {},
	'u': {},
	'á': {},
	'é': {},
	'í': {},
	'ó': {},
	'ú': {},
	'ü': {},
}

func isVowel(r rune) (ok bool) {
	_, ok = vowels[r]
	return
}

type runeSlot struct {
	Index   int
	Value   rune
	IsVowel bool
}

type region struct {
	data  []byte
	runes []*runeSlot
}

func r1r2rv(w []byte) (r1, r2, rv *region) {
	// UTF8 implicit!
	foundVowel := false
	foundNonVowel := false
	runes := make([]*runeSlot, 0)
	for index, runeValue := range string(w) {
		runes = append(runes, &runeSlot{Index: index, Value: runeValue, IsVowel: isVowel(runeValue)})
	}
	runesLength := len(runes)
	// R1
	for i, slot := range runes {
		if foundNonVowel {
			r1 = &region{data: w[slot.Index:], runes: runes[i:]}
			break
		}
		if foundVowel {
			if !slot.IsVowel {
				foundNonVowel = true
			}
		}
		if slot.IsVowel {
			foundVowel = true
		}
	}
	if r1 == nil {
		return
	}
	// R2
	foundVowel = false
	foundNonVowel = false
	for i, runeSlot := range r1.runes {
		if foundNonVowel {
			r2 = &region{data: w[runeSlot.Index:], runes: runes[i:]}
			break
		}
		if foundVowel {
			if !runeSlot.IsVowel {
				foundNonVowel = true
			}
		}
		if runeSlot.IsVowel {
			foundVowel = true
		}
	}
	// RV. If the second letter is a consonant, RV is the region after the next following vowel, or if the first
	// two letters are vowels, RV is the region after the next consonant, and otherwise (consonant-vowel case) RV
	// is the region after the third letter. But RV is the end of the word if these positions cannot be found.
	if len(runes) > 2 {
		foundNonVowel = false
		foundVowel = false
		if runes[0].IsVowel && runes[1].IsVowel {
			i := 2
			for ; i < runesLength; i++ {
				run := runes[i]
				if foundNonVowel {
					rv = &region{data: w[run.Index:], runes: runes[i:]}
					return
				}
				if !run.IsVowel {
					foundNonVowel = true
				}
			}
		}

		if !runes[1].IsVowel {
			i := 2
			for ; i < runesLength; i++ {
				run := runes[i]
				if foundVowel {
					rv = &region{data: w[run.Index:], runes: runes[i:]}
					return
				}
				if run.IsVowel {
					foundVowel = true
				}
			}
		}

		if runesLength > 3 && !runes[0].IsVowel && runes[1].IsVowel {
			rv = &region{data: w[runes[3].Index:], runes: runes[3:]}
			return
		}

	}
	return
}

type stemmer struct{}

func (s *stemmer) Stem(w []byte) []byte {
	return nil
}

func NewSnowball() *stemmer {
	return new(stemmer)
}
