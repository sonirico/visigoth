package spanish

import (
	"bytes"
)

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

type suffix struct {
	runes    []runeSlot
	bytesLen int
}

func newSuffix(literal string) suffix {
	runes := make([]runeSlot, 0)
	for i, runev := range literal {
		runes = append(runes, runeSlot{Index: i, Value: runev})
	}
	return suffix{runes: runes, bytesLen: len(literal)}
}

type runeSlot struct {
	Index   int
	Value   rune
	IsVowel bool
}

type region struct {
	data  []byte
	runes []*runeSlot // todo: consider remove pointer?
}

func newRegion(literal string) *region {
	runes := make([]*runeSlot, 0)
	for i, runev := range literal {
		runes = append(runes, &runeSlot{Index: i, Value: runev})
	}
	return &region{data: []byte(literal), runes: runes}
}

func (r *region) At(index int) rune {
	return r.runes[index].Value
}

func (r *region) Last() rune {
	return r.At(len(r.runes) - 1)
}

func (r *region) HasSuffix(s suffix) bool {
	runesl := len(r.runes)

	if len(s.runes) > runesl {
		return false
	}

	for slen := len(s.runes); slen > 0; slen-- {
		if r.runes[runesl-1].Value != s.runes[slen-1].Value {
			return false
		}
		runesl--
	}

	return true
}

func (r *region) RemoveSuffix(s suffix) bool {
	r.data = r.data[:s.bytesLen-1]
	r.runes = r.runes[:len(r.runes)-len(s.runes)]
	return true
}

func (r *region) RawString() string {
	return string(r.data)
}

func (r *region) RuneString() string {
	var buf bytes.Buffer
	for _, slot := range r.runes {
		buf.WriteRune(slot.Value)
	}
	return buf.String()
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

/**
Step 0: Attached pronoun

    Search for the longest among the following suffixes

        me   se   sela   selo   selas   selos   la   le   lo   las   les   los   nos


    and delete it, if comes after one of

        (a) iéndo   ándo   ár   ér   ír
        (b) ando   iendo   ar   er   ir
        (c) yendo following u


    in RV. In the case of (c), yendo must lie in RV, but the preceding u can be outside it.

    In the case of (a), deletion is followed by removing the acute accent (for example, haciéndola -> haciendo).
*/

type step0Suffix struct {
	fullSuffix suffix
	suffix     suffix
}

// TODO: Raddix tree
// step0suffixes is a sorted list of suffixes by length. As longer suffixes have priority over
// lower, and lower can overlap with larger ones, longest ones are checked first.
var step0suffixesMain = []step0Suffix{
	{fullSuffix: newSuffix("uyendoselos"), suffix: newSuffix("selos")},
	{fullSuffix: newSuffix("uyendoselas"), suffix: newSuffix("selas")},
	{fullSuffix: newSuffix("iéndoselos"), suffix: newSuffix("selos")},
	{fullSuffix: newSuffix("iéndoselas"), suffix: newSuffix("selas")},
	{fullSuffix: newSuffix("iendoselos"), suffix: newSuffix("selos")},
	{fullSuffix: newSuffix("iendoselas"), suffix: newSuffix("selas")},
	{fullSuffix: newSuffix("uyendoselo"), suffix: newSuffix("selo")},
	{fullSuffix: newSuffix("uyendosela"), suffix: newSuffix("sela")},
	{fullSuffix: newSuffix("iéndoselo"), suffix: newSuffix("selo")},
	{fullSuffix: newSuffix("iéndosela"), suffix: newSuffix("sela")},
	{fullSuffix: newSuffix("ándoselos"), suffix: newSuffix("selos")},
	{fullSuffix: newSuffix("ándoselas"), suffix: newSuffix("selas")},
	{fullSuffix: newSuffix("iendoselo"), suffix: newSuffix("selo")},
	{fullSuffix: newSuffix("iendosela"), suffix: newSuffix("sela")},
	{fullSuffix: newSuffix("uyendoles"), suffix: newSuffix("les")},
	{fullSuffix: newSuffix("uyendolas"), suffix: newSuffix("las")},
	{fullSuffix: newSuffix("uyendolos"), suffix: newSuffix("los")},
	{fullSuffix: newSuffix("uyendonos"), suffix: newSuffix("nos")},
	{fullSuffix: newSuffix("iéndoles"), suffix: newSuffix("les")},
	{fullSuffix: newSuffix("iéndolas"), suffix: newSuffix("las")},
	{fullSuffix: newSuffix("iéndolos"), suffix: newSuffix("los")},
	{fullSuffix: newSuffix("iéndonos"), suffix: newSuffix("nos")},
	{fullSuffix: newSuffix("ándoselo"), suffix: newSuffix("selo")},
	{fullSuffix: newSuffix("ándosela"), suffix: newSuffix("sela")},
	{fullSuffix: newSuffix("andselos"), suffix: newSuffix("selos")},
	{fullSuffix: newSuffix("andselas"), suffix: newSuffix("selas")},
	{fullSuffix: newSuffix("iendoles"), suffix: newSuffix("les")},
	{fullSuffix: newSuffix("iendolas"), suffix: newSuffix("las")},
	{fullSuffix: newSuffix("iendolos"), suffix: newSuffix("los")},
	{fullSuffix: newSuffix("iendonos"), suffix: newSuffix("nos")},
	{fullSuffix: newSuffix("uyendome"), suffix: newSuffix("me")},
	{fullSuffix: newSuffix("uyendose"), suffix: newSuffix("se")},
	{fullSuffix: newSuffix("uyendola"), suffix: newSuffix("la")},
	{fullSuffix: newSuffix("uyendole"), suffix: newSuffix("le")},
	{fullSuffix: newSuffix("uyendolo"), suffix: newSuffix("lo")},
	{fullSuffix: newSuffix("iéndome"), suffix: newSuffix("me")},
	{fullSuffix: newSuffix("iéndose"), suffix: newSuffix("se")},
	{fullSuffix: newSuffix("iéndola"), suffix: newSuffix("la")},
	{fullSuffix: newSuffix("iéndole"), suffix: newSuffix("le")},
	{fullSuffix: newSuffix("iéndolo"), suffix: newSuffix("lo")},
	{fullSuffix: newSuffix("ándoles"), suffix: newSuffix("les")},
	{fullSuffix: newSuffix("ándolas"), suffix: newSuffix("las")},
	{fullSuffix: newSuffix("ándolos"), suffix: newSuffix("los")},
	{fullSuffix: newSuffix("ándonos"), suffix: newSuffix("nos")},
	{fullSuffix: newSuffix("árselos"), suffix: newSuffix("selos")},
	{fullSuffix: newSuffix("árselas"), suffix: newSuffix("selas")},
	{fullSuffix: newSuffix("érselos"), suffix: newSuffix("selos")},
	{fullSuffix: newSuffix("érselas"), suffix: newSuffix("selas")},
	{fullSuffix: newSuffix("írselos"), suffix: newSuffix("selos")},
	{fullSuffix: newSuffix("írselas"), suffix: newSuffix("selas")},
	{fullSuffix: newSuffix("andselo"), suffix: newSuffix("selo")},
	{fullSuffix: newSuffix("andsela"), suffix: newSuffix("sela")},
	{fullSuffix: newSuffix("iendome"), suffix: newSuffix("me")},
	{fullSuffix: newSuffix("iendose"), suffix: newSuffix("se")},
	{fullSuffix: newSuffix("iendola"), suffix: newSuffix("la")},
	{fullSuffix: newSuffix("iendole"), suffix: newSuffix("le")},
	{fullSuffix: newSuffix("iendolo"), suffix: newSuffix("lo")},
	{fullSuffix: newSuffix("arselos"), suffix: newSuffix("selos")},
	{fullSuffix: newSuffix("arselas"), suffix: newSuffix("selas")},
	{fullSuffix: newSuffix("erselos"), suffix: newSuffix("selos")},
	{fullSuffix: newSuffix("erselas"), suffix: newSuffix("selas")},
	{fullSuffix: newSuffix("irselos"), suffix: newSuffix("selos")},
	{fullSuffix: newSuffix("irselas"), suffix: newSuffix("selas")},
	{fullSuffix: newSuffix("ándome"), suffix: newSuffix("me")},
	{fullSuffix: newSuffix("ándose"), suffix: newSuffix("se")},
	{fullSuffix: newSuffix("ándola"), suffix: newSuffix("la")},
	{fullSuffix: newSuffix("ándole"), suffix: newSuffix("le")},
	{fullSuffix: newSuffix("ándolo"), suffix: newSuffix("lo")},
	{fullSuffix: newSuffix("árselo"), suffix: newSuffix("selo")},
	{fullSuffix: newSuffix("ársela"), suffix: newSuffix("sela")},
	{fullSuffix: newSuffix("érselo"), suffix: newSuffix("selo")},
	{fullSuffix: newSuffix("érsela"), suffix: newSuffix("sela")},
	{fullSuffix: newSuffix("írselo"), suffix: newSuffix("selo")},
	{fullSuffix: newSuffix("írsela"), suffix: newSuffix("sela")},
	{fullSuffix: newSuffix("andles"), suffix: newSuffix("les")},
	{fullSuffix: newSuffix("andlas"), suffix: newSuffix("las")},
	{fullSuffix: newSuffix("andlos"), suffix: newSuffix("los")},
	{fullSuffix: newSuffix("andnos"), suffix: newSuffix("nos")},
	{fullSuffix: newSuffix("arselo"), suffix: newSuffix("selo")},
	{fullSuffix: newSuffix("arsela"), suffix: newSuffix("sela")},
	{fullSuffix: newSuffix("erselo"), suffix: newSuffix("selo")},
	{fullSuffix: newSuffix("ersela"), suffix: newSuffix("sela")},
	{fullSuffix: newSuffix("irselo"), suffix: newSuffix("selo")},
	{fullSuffix: newSuffix("irsela"), suffix: newSuffix("sela")},
	{fullSuffix: newSuffix("árles"), suffix: newSuffix("les")},
	{fullSuffix: newSuffix("árlas"), suffix: newSuffix("las")},
	{fullSuffix: newSuffix("árlos"), suffix: newSuffix("los")},
	{fullSuffix: newSuffix("árnos"), suffix: newSuffix("nos")},
	{fullSuffix: newSuffix("érles"), suffix: newSuffix("les")},
	{fullSuffix: newSuffix("érlas"), suffix: newSuffix("las")},
	{fullSuffix: newSuffix("érlos"), suffix: newSuffix("los")},
	{fullSuffix: newSuffix("érnos"), suffix: newSuffix("nos")},
	{fullSuffix: newSuffix("írles"), suffix: newSuffix("les")},
	{fullSuffix: newSuffix("írlas"), suffix: newSuffix("las")},
	{fullSuffix: newSuffix("írlos"), suffix: newSuffix("los")},
	{fullSuffix: newSuffix("írnos"), suffix: newSuffix("nos")},
	{fullSuffix: newSuffix("andme"), suffix: newSuffix("me")},
	{fullSuffix: newSuffix("andse"), suffix: newSuffix("se")},
	{fullSuffix: newSuffix("andla"), suffix: newSuffix("la")},
	{fullSuffix: newSuffix("andle"), suffix: newSuffix("le")},
	{fullSuffix: newSuffix("andlo"), suffix: newSuffix("lo")},
	{fullSuffix: newSuffix("arles"), suffix: newSuffix("les")},
	{fullSuffix: newSuffix("arlas"), suffix: newSuffix("las")},
	{fullSuffix: newSuffix("arlos"), suffix: newSuffix("los")},
	{fullSuffix: newSuffix("arnos"), suffix: newSuffix("nos")},
	{fullSuffix: newSuffix("erles"), suffix: newSuffix("les")},
	{fullSuffix: newSuffix("erlas"), suffix: newSuffix("las")},
	{fullSuffix: newSuffix("erlos"), suffix: newSuffix("los")},
	{fullSuffix: newSuffix("ernos"), suffix: newSuffix("nos")},
	{fullSuffix: newSuffix("irles"), suffix: newSuffix("les")},
	{fullSuffix: newSuffix("irlas"), suffix: newSuffix("las")},
	{fullSuffix: newSuffix("irlos"), suffix: newSuffix("los")},
	{fullSuffix: newSuffix("irnos"), suffix: newSuffix("nos")},
	{fullSuffix: newSuffix("árme"), suffix: newSuffix("me")},
	{fullSuffix: newSuffix("árse"), suffix: newSuffix("se")},
	{fullSuffix: newSuffix("árla"), suffix: newSuffix("la")},
	{fullSuffix: newSuffix("árle"), suffix: newSuffix("le")},
	{fullSuffix: newSuffix("árlo"), suffix: newSuffix("lo")},
	{fullSuffix: newSuffix("érme"), suffix: newSuffix("me")},
	{fullSuffix: newSuffix("érse"), suffix: newSuffix("se")},
	{fullSuffix: newSuffix("érla"), suffix: newSuffix("la")},
	{fullSuffix: newSuffix("érle"), suffix: newSuffix("le")},
	{fullSuffix: newSuffix("érlo"), suffix: newSuffix("lo")},
	{fullSuffix: newSuffix("írme"), suffix: newSuffix("me")},
	{fullSuffix: newSuffix("írse"), suffix: newSuffix("se")},
	{fullSuffix: newSuffix("írla"), suffix: newSuffix("la")},
	{fullSuffix: newSuffix("írle"), suffix: newSuffix("le")},
	{fullSuffix: newSuffix("írlo"), suffix: newSuffix("lo")},
	{fullSuffix: newSuffix("arme"), suffix: newSuffix("me")},
	{fullSuffix: newSuffix("arse"), suffix: newSuffix("se")},
	{fullSuffix: newSuffix("arla"), suffix: newSuffix("la")},
	{fullSuffix: newSuffix("arle"), suffix: newSuffix("le")},
	{fullSuffix: newSuffix("arlo"), suffix: newSuffix("lo")},
	{fullSuffix: newSuffix("erme"), suffix: newSuffix("me")},
	{fullSuffix: newSuffix("erse"), suffix: newSuffix("se")},
	{fullSuffix: newSuffix("erla"), suffix: newSuffix("la")},
	{fullSuffix: newSuffix("erle"), suffix: newSuffix("le")},
	{fullSuffix: newSuffix("erlo"), suffix: newSuffix("lo")},
	{fullSuffix: newSuffix("irme"), suffix: newSuffix("me")},
	{fullSuffix: newSuffix("irse"), suffix: newSuffix("se")},
	{fullSuffix: newSuffix("irla"), suffix: newSuffix("la")},
	{fullSuffix: newSuffix("irle"), suffix: newSuffix("le")},
	{fullSuffix: newSuffix("irlo"), suffix: newSuffix("lo")},
}
var step0suffixesYendo = []step0Suffix{
	{fullSuffix: newSuffix("yendoselos"), suffix: newSuffix("selos")},
	{fullSuffix: newSuffix("yendoselas"), suffix: newSuffix("selas")},
	{fullSuffix: newSuffix("yendoselo"), suffix: newSuffix("selo")},
	{fullSuffix: newSuffix("yendosela"), suffix: newSuffix("sela")},
	{fullSuffix: newSuffix("yendoles"), suffix: newSuffix("les")},
	{fullSuffix: newSuffix("yendolas"), suffix: newSuffix("las")},
	{fullSuffix: newSuffix("yendolos"), suffix: newSuffix("los")},
	{fullSuffix: newSuffix("yendonos"), suffix: newSuffix("nos")},
	{fullSuffix: newSuffix("yendome"), suffix: newSuffix("me")},
	{fullSuffix: newSuffix("yendose"), suffix: newSuffix("se")},
	{fullSuffix: newSuffix("yendola"), suffix: newSuffix("la")},
	{fullSuffix: newSuffix("yendole"), suffix: newSuffix("le")},
	{fullSuffix: newSuffix("yendolo"), suffix: newSuffix("lo")},
}

func removeSuffixes(r1, r2, rv *region) bool {
	for _, suffix := range step0suffixesMain {
		if rv.HasSuffix(suffix.fullSuffix) {
			return rv.RemoveSuffix(suffix.suffix)
		}
	}
	for _, suffix := range step0suffixesYendo {
		if rv.HasSuffix(suffix.fullSuffix) {
			if r2 != nil && r2.Last() == 'u' {
				return rv.RemoveSuffix(suffix.suffix)
			}
			if r1 != nil && r1.Last() == 'u' {
				return rv.RemoveSuffix(suffix.suffix)
			}
		}
	}

	return false
}

func step0(r1, r2, rv *region) {
	removeSuffixes(r1, r2, rv)
	// removeAccents(r1, r2, rv)
}

/**
let ii = [
  'yendo',

]

let ss = [
  'selos', 'selas', 'selo',
  'sela',  'les',   'las',
  'los',   'nos',   'me',
  'se',    'la',    'le',
  'lo'
]

let c = []

for (const i of ii) {
	for (const s of ss) {
		c.push({total: i +s, prefix: i, suffix: s})
	}
}

c.sort((a, b) => b.total.length - a.total.length)

for (const cc of c) {
 console.log(`{fullSuffix: newSuffix("${cc.total}"), suffix: newSuffix("${cc.suffix}")},`);
}


*/
