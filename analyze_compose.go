package visigoth

type TokenizationPipeline struct {
	tokenizer Tokenizer
	filters   []Filter
}

func (p *TokenizationPipeline) Tokenize(text string) []string {
	var res = p.tokenizer.Tokenize(text)
	for _, filter := range p.filters {
		res = filter.Filter(res)
	}
	return res
}

func NewTokenizationPipeline(t Tokenizer, f ...Filter) *TokenizationPipeline {
	return &TokenizationPipeline{
		tokenizer: t,
		filters:   f,
	}
}
