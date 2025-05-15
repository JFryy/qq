package json

import (
	"bytes"
	"github.com/goccy/go-json"
)

type Codec struct{}

func (c *Codec) Marshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(v)
	if err != nil {
		return nil, err
	}
	encodedBytes := bytes.TrimSpace(buf.Bytes())
	return encodedBytes, nil
}
