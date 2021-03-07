package spanish

import "github.com/sonirico/visigoth/internal/tokenizer"

type stemmer struct {
	cleanupTokenizer tokenizer.SimpleTokenizer
}

func (s *stemmer) Stem(word []byte) ([]byte, bool) {
	// TODO: Accents remove and lowercase steps are done prior to step0. Should be done as the final step of step0.
	cleaned, ok := s.cleanupTokenizer.TokenizeWord(word)
	if !ok {
		return word, false
	}
	r1, r2, rv := r1r2rv(cleaned)
	step0(r1, r2, rv)
	//step1(r1, r2, rv)
	return nil, false
}

func NewSnowball() stemmer {
	return stemmer{}
}
