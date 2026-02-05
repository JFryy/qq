package base64

import (
	"encoding/base64"
	"errors"

	"github.com/goccy/go-json"
)

// Codec handles base64 encoding/decoding
type Codec struct{}

// Unmarshal decodes base64 data and parses as JSON
func (c *Codec) Unmarshal(data []byte, v interface{}) error {
	if v == nil {
		return errors.New("v cannot be nil")
	}

	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return err
	}

	// Parse decoded data as JSON
	return json.Unmarshal(decoded, v)
}

// Marshal converts data to JSON and encodes as base64
func (c *Codec) Marshal(v interface{}) ([]byte, error) {
	// Marshal to JSON first
	jsonData, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(jsonData)
	return []byte(encoded), nil
}
