package visigoth

//go:generate easyjson

//easyjson:json
type SearchResult struct {
	Document Doc `json:"doc"`
	Hits     int `json:"hits"`
}

func (r SearchResult) GetDoc() Doc {
	return r.Document
}

func (r SearchResult) GetHits() int {
	return r.Hits
}

// Implementación de la interfaz Row
func (r SearchResult) Doc() Doc {
	return r.Document
}

type SearchResults []SearchResult

// sort.Interface implementation
func (r SearchResults) Len() int {
	return len(r)
}

func (r SearchResults) Less(i, j int) bool {
	return r[i].Hits > r[j].Hits
}

func (r SearchResults) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
