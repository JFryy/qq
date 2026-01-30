package yaml

import (
	"bytes"
	"io"

	"github.com/goccy/go-yaml"
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
	normalized := normalizeTypes(val)

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
func normalizeTypes(val any) any {
	switch v := val.(type) {
	case uint:
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		if v <= 9007199254740992 { // Max safe integer in JSON (2^53)
			return int(v)
		}
		return float64(v)
	case map[string]any:
		result := make(map[string]any, len(v))
		for key, value := range v {
			result[key] = normalizeTypes(value)
		}
		return result
	case []any:
		result := make([]any, len(v))
		for i, value := range v {
			result[i] = normalizeTypes(value)
		}
		return result
	default:
		return v
	}
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
				docBytes, err := yaml.Marshal(doc)
				if err != nil {
					return nil, err
				}
				buf.Write(docBytes)
			}
			return buf.Bytes(), nil
		}
	}

	// For everything else, use standard YAML marshaling
	return yaml.Marshal(v)
}
