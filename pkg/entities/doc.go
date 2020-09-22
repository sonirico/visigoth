package entities

type Doc struct {
	Name    string `json:"id"`
	Content string `json:"raw"`
}

func NewDoc(name, content string) Doc {
	return Doc{Name: name, Content: content}
}

func (d Doc) Id() string {
	return d.Name
}

func (d Doc) Raw() string {
	return d.Content
}
