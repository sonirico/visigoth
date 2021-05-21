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

//nolint:nakedret,gomnd
func R1R2RV(w []byte) (r1 []byte, r2 []byte, rv []byte) {
	// UTF8 implicit!
	foundVowel := false
	foundNonVowel := false
	type runeSlot struct {
		Index   int
		Value   rune
		IsVowel bool
	}
	runes := make([]runeSlot, 0)
	for index, runeValue := range string(w) {
		runes = append(runes, runeSlot{Index: index, Value: runeValue, IsVowel: isVowel(runeValue)})
	}
	runesLength := len(runes)
	// R1
	for _, slot := range runes {
		if foundNonVowel {
			r1 = w[slot.Index:]
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
	for index, runeValue := range string(r1) {
		if foundNonVowel {
			r2 = r1[index:]
			break
		}
		if foundVowel {
			if !isVowel(runeValue) {
				foundNonVowel = true
			}
		}
		if isVowel(runeValue) {
			foundVowel = true
		}
	}
	// RV. If the second letter is a consonant, RV is the region after the next following vowel, or if the first
	// two letters are vowels, RV is the region after the next consonant, and otherwise (consonant-vowel case) RV
	// is the region after the third letter. But RV is the end of the word if these positions cannot be found.
	if len(runes) < 3 {
		return
	}
	foundNonVowel = false
	foundVowel = false
	if runes[0].IsVowel && runes[1].IsVowel {
		i := 2
		for ; i < runesLength; i++ {
			run := runes[i]
			if foundNonVowel {
				rv = w[run.Index:]
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
				rv = w[run.Index:]
				return
			}
			if run.IsVowel {
				foundVowel = true
			}
		}
	}

	if runesLength > 3 && !runes[0].IsVowel && runes[1].IsVowel {
		rv = w[runes[3].Index:]
		return
	}

	return
}

type Stemmer struct{}

func (s *Stemmer) Stem(w []byte) []byte {
	return nil
}

func NewSnowball() *Stemmer {
	return new(Stemmer)
}
