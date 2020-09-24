package loaders

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"strconv"
)

const EOL = '\n'

type LoadedDoc struct {
	doc  []byte
	text []byte
}

type Loader interface {
	Load(reader io.Reader, writer io.Writer) error
}

type JsonRowLoader struct {
	withKeys bool
}

func NewJSONRowLoader(withKeys bool) *JsonRowLoader {
	return &JsonRowLoader{withKeys: withKeys}
}

func (j *JsonRowLoader) Load(src io.Reader, dst io.Writer) error {
	buf := bufio.NewReader(src)
	for {
		by, err := buf.ReadBytes(EOL)
		if err != nil {
			log.Fatal(err)
		}
		original := bytes.TrimSpace(by)
		if len(original) < 1 {
			return nil
		}
		object := make(map[string]interface{})
		if err := json.Unmarshal(by, &object); err != nil {
			log.Printf("invalid row, skipping")
			continue
		}
		text := bytes.TrimSpace(compact(object, j.withKeys))
		if len(text) > 0 {
			if _, err := dst.Write(append(original, text...)); err != nil {
				return err
			}
		}
		// index.Put(api.NewDocRequest("imported", compact(object)))
	}
}

func compact(data interface{}, withKeys bool) []byte {
	switch val := data.(type) {
	case map[string]interface{}:
		return compactMap(val, withKeys)
	case []interface{}:
		return compactList(val, withKeys)
	case float64:
		return []byte(strconv.FormatFloat(val, 'f', 0, 64))
	case string:
		return []byte(val)
	default:
		return []byte("")
	}
}

func compactList(listLike []interface{}, withKeys bool) []byte {
	buf := bytes.NewBuffer(nil)
	for _, v := range listLike {
		buf.Write(compact(v, withKeys))
		buf.WriteString(" ")
	}
	return buf.Bytes()
}

func compactMap(mapLike map[string]interface{}, withKeys bool) []byte {
	buf := bytes.NewBuffer(nil)
	for k, v := range mapLike {
		if withKeys {
			buf.WriteString(k)
		}
		buf.Write(compact(v, withKeys))
		buf.WriteString(" ")
	}
	return buf.Bytes()
}
