package loaders

import (
	"encoding/json"
	"io"
)

type JSONLoader struct {
	withKeys bool
}

func (j *JSONLoader) Load(src io.Reader, dst io.Writer) error {
	dec := json.NewDecoder(src)
	data := make(map[string]interface{})
	if err := dec.Decode(&data); err != nil {
		return err
	}
	if _, err := dst.Write(compact(data, j.withKeys)); err != nil {
		return err
	}
	return nil
}

func NewJSONLoader(withKeys bool) *JSONLoader {
	return &JSONLoader{withKeys: withKeys}
}
