package visigoth

type MimeType byte

const (
	MimeText MimeType = iota + 1
	MimeJSON
)

type DocRequest struct {
	Name      string
	Content   string
	statement string
	MimeType  MimeType
}

func (d DocRequest) ID() string        { return d.Name }
func (d DocRequest) Raw() string       { return d.Content }
func (d DocRequest) Mime() MimeType    { return d.MimeType }
func (d DocRequest) Statement() string { return d.statement }

func NewDocRequest(name, content string) DocRequest {
	return DocRequest{
		Name:      name,
		Content:   content,
		MimeType:  MimeText,
		statement: content,
	}
}

func NewDocRequestWith(name, content, statement string) DocRequest {
	return DocRequest{
		Name:      name,
		Content:   content,
		MimeType:  MimeText,
		statement: statement,
	}
}

func NewDocRequestWithMime(name, content string, mime MimeType) DocRequest {
	return DocRequest{
		Name:     name,
		Content:  content,
		MimeType: mime,
	}
}
