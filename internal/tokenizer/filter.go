package tokenizer

type Filter interface {
	Filter(payload []byte) bool
}

type StopWordsTokenizerFilter struct {
	Tokenizer
	// Filled on startup, never changed from there on, therefore, mutex free.
	stopWords map[string]struct{}
}

func (st *StopWordsTokenizerFilter) TokenizeText(payload []byte) [][]byte {
	var res [][]byte
	for _, binaryWord := range st.Tokenizer.TokenizeText(payload) {
		if st.Filter(binaryWord) {
			res = append(res, binaryWord)
		}
	}
	return res
}

func (st *StopWordsTokenizerFilter) Filter(payload []byte) bool {
	_, ok := st.stopWords[string(payload)]
	return !ok
}

func NewStopWordsTokenizerFilter(stopWords []string, inner Tokenizer) Tokenizer {
	t := &StopWordsTokenizerFilter{
		Tokenizer: inner,
		stopWords: make(map[string]struct{}, len(stopWords)),
	}
	for _, w := range stopWords {
		if stopWordToken, ok := inner.TokenizeWord([]byte(w)); ok {
			if len(stopWordToken) > 0 {
				t.stopWords[string(stopWordToken)] = struct{}{}
			}
		}
	}
	return t
}
