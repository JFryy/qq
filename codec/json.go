package codec

import (
	"bytes"
	"github.com/goccy/go-json"
)

func jsonMarshal(v interface{}) ([]byte, error) {
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
