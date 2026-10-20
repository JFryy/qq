package yaml

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"go.yaml.in/yaml/v4"
)

type Codec struct{}

// Unmarshal handles both single and multi-document YAML.
// For multi-document YAML (separated by ---), it returns an array of documents.
// For single-document YAML, it returns the document as-is.
func (c Codec) Unmarshal(data []byte, v any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	// Try to decode the first document
	var firstDoc any
	err := decoder.Decode(&firstDoc)
	if err != nil {
		return err
	}

	// Try to decode a second document to check if this is multi-document YAML
	var secondDoc any
	err = decoder.Decode(&secondDoc)
	if err == io.EOF {
		// Only one document, return it directly
		return setInterface(v, firstDoc)
	}
	if err != nil {
		return err
	}

	// Multiple documents exist, collect them all into an array
	docs := []any{firstDoc, secondDoc}

	// Continue reading remaining documents
	for {
		var doc any
		err := decoder.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		docs = append(docs, doc)
	}

	return setInterface(v, docs)
}

// setInterface sets the value of v to val
func setInterface(v any, val any) error {
	// Normalize types to be compatible with gojq/JSON
	normalized, err := normalizeTypes(val)
	if err != nil {
		return err
	}

	// v is a pointer to any, so we need to set it properly
	switch ptr := v.(type) {
	case *any:
		*ptr = normalized
		return nil
	default:
		// If it's a specific type, unmarshal val into it
		b, err := yaml.Marshal(normalized)
		if err != nil {
			return err
		}
		return yaml.Unmarshal(b, v)
	}
}

// normalizeTypes converts YAML-specific types (uint, uint64, etc.) to types
// compatible with JSON and gojq (int, float64, etc.)
func normalizeTypes(val any) (any, error) {
	switch v := val.(type) {
	case time.Time:
		return timestampString(v), nil
	case uint:
		return int(v), nil
	case uint8:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint64:
		if v <= 9007199254740992 { // Max safe integer in JSON (2^53)
			return int(v), nil
		}
		return float64(v), nil
	case map[string]any:
		result := make(map[string]any, len(v))
		for key, value := range v {
			normalized, err := normalizeTypes(value)
			if err != nil {
				return nil, err
			}
			result[key] = normalized
		}
		return result, nil
	case map[any]any:
		// A mapping with a non-string key (e.g. `200: what`) decodes to
		// map[any]any, which neither JSON nor gojq accept, so stringify the
		// keys. Two keys can stringify to the same value (`1` and `1.0`,
		// `true` and "true"); report that instead of keeping one at random.
		result := make(map[string]any, len(v))
		for key, value := range v {
			name := mapKey(key)
			if _, exists := result[name]; exists {
				return nil, fmt.Errorf("yaml: mapping has more than one key that converts to %q", name)
			}
			normalized, err := normalizeTypes(value)
			if err != nil {
				return nil, err
			}
			result[name] = normalized
		}
		return result, nil
	case []any:
		result := make([]any, len(v))
		for i, value := range v {
			normalized, err := normalizeTypes(value)
			if err != nil {
				return nil, err
			}
			result[i] = normalized
		}
		return result, nil
	default:
		return v, nil
	}
}

// mapKey returns the string form of a YAML mapping key.
func mapKey(key any) string {
	switch k := key.(type) {
	case string:
		return k
	case nil:
		return "null"
	case time.Time:
		return timestampString(k)
	default:
		return fmt.Sprint(k)
	}
}

// timestampString formats a YAML timestamp: a date alone (decoded as midnight
// UTC) as 2006-01-02, like yq, and anything else as RFC 3339.
func timestampString(t time.Time) string {
	if t.Location() == time.UTC && t.Equal(t.Truncate(24*time.Hour)) {
		return t.Format(time.DateOnly)
	}
	return t.Format(time.RFC3339Nano)
}

// Marshal handles both single values and arrays.
// For arrays of maps/objects, it outputs multi-document YAML (with --- separators).
// For simple arrays or single values, it uses standard YAML marshaling.
func (c Codec) Marshal(v any) ([]byte, error) {
	// Check if this is a slice of objects that should be output as multi-document YAML
	if slice, ok := v.([]any); ok && len(slice) > 0 {
		// Check if all elements are maps (objects)
		allMaps := true
		for _, item := range slice {
			if _, isMap := item.(map[string]any); !isMap {
				allMaps = false
				break
			}
		}

		// If all items are maps, output as multi-document YAML
		if allMaps {
			var buf bytes.Buffer
			for i, doc := range slice {
				// Add document separator
				if i > 0 {
					buf.WriteString("\n")
				}
				buf.WriteString("---\n")

				// Marshal the document
				docBytes, err := marshalIndent(doc)
				if err != nil {
					return nil, err
				}
				buf.Write(docBytes)
			}
			return buf.Bytes(), nil
		}
	}

	// For everything else, use standard YAML marshaling
	return marshalIndent(v)
}

func marshalIndent(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
