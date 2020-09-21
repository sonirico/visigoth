package loaders

import "io"

type TextLoader struct{}

func (t *TextLoader) Load(src io.Reader, dst io.Writer) error {
	_, err := io.Copy(dst, src)
	return err
}

func NewTextLoader() *TextLoader {
	return &TextLoader{}
}
