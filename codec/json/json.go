package json

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"

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

// Unmarshal decodes JSON, keeping integers exact for generic destinations.
func (c *Codec) Unmarshal(data []byte, v any) error {
	return Unmarshal(data, v)
}

// Unmarshal decodes a single JSON value. Generic destinations receive int for
// integers, because decoding them as float64 corrupts values above 2^53.
func Unmarshal(data []byte, v any) error {
	ptr, ok := v.(*any)
	if !ok {
		return json.Unmarshal(data, v)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		if err != nil {
			return err
		}
		return fmt.Errorf("invalid character after top-level value")
	}
	normalized, err := NormalizeNumbers(value)
	if err != nil {
		return err
	}
	*ptr = normalized
	return nil
}

// NormalizeNumbers replaces json.Number values with int when they are whole
// numbers within int range, and float64 otherwise.
func NormalizeNumbers(value any) (any, error) {
	switch v := value.(type) {
	case json.Number:
		return normalizeNumber(v)
	case map[string]any:
		for key, item := range v {
			normalized, err := NormalizeNumbers(item)
			if err != nil {
				return nil, err
			}
			v[key] = normalized
		}
	case []any:
		for i, item := range v {
			normalized, err := NormalizeNumbers(item)
			if err != nil {
				return nil, err
			}
			v[i] = normalized
		}
	}
	return value, nil
}

func normalizeNumber(n json.Number) (any, error) {
	if i, err := n.Int64(); err == nil && i >= math.MinInt && i <= math.MaxInt {
		return int(i), nil
	}
	// Values that are not valid integers keep the previous float64 behaviour.
	return n.Float64()
}
