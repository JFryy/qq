package yaml

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/JFryy/qq/codec/keyorder"
	"go.yaml.in/yaml/v4"
)

type Codec struct{}

// Unmarshal handles both single and multi-document YAML.
// For multi-document YAML (separated by ---), it returns an array of documents.
// For single-document YAML, it returns the document as-is.
func (c Codec) Unmarshal(data []byte, v any) error {
	if keyorder.Recording() {
		return unmarshalRecordingOrder(data, v)
	}
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

// unmarshalRecordingOrder decodes like Unmarshal, through yaml.Node so that
// the key order of each mapping can be recorded on the decoded maps.
func unmarshalRecordingOrder(data []byte, v any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var docs []any
	var nodes []*yaml.Node
	for {
		var node yaml.Node
		err := decoder.Decode(&node)
		if err == io.EOF && len(docs) > 0 {
			break
		}
		if err != nil {
			return err
		}
		var doc any
		if err := node.Decode(&doc); err != nil {
			return err
		}
		docs = append(docs, doc)
		nodes = append(nodes, &node)
	}

	var value any = docs
	if len(docs) == 1 {
		value = docs[0]
	}
	normalized, err := normalizeTypes(value)
	if err != nil {
		return err
	}

	if len(nodes) == 1 {
		err = recordKeyOrder(nodes[0], normalized)
	} else {
		for i, doc := range normalized.([]any) {
			if err = recordKeyOrder(nodes[i], doc); err != nil {
				break
			}
		}
	}
	if err != nil {
		return fmt.Errorf("yaml: recording key order: %w", err)
	}
	return assign(v, normalized)
}

// recordKeyOrder records the key order of each mapping in node on the matching
// map in value, the normalized result of decoding node.
func recordKeyOrder(node *yaml.Node, value any) error {
	node = resolveNode(node)
	switch node.Kind {
	case yaml.MappingNode:
		m, ok := value.(map[string]any)
		if !ok {
			return nil
		}
		keys, values, err := mappingEntries(node)
		if err != nil {
			return err
		}
		keyorder.Record(m, keys)
		for i, key := range keys {
			if err := recordKeyOrder(values[i], m[key]); err != nil {
				return err
			}
		}
	case yaml.SequenceNode:
		items, ok := value.([]any)
		if !ok || len(items) != len(node.Content) {
			return nil
		}
		for i, item := range items {
			if err := recordKeyOrder(node.Content[i], item); err != nil {
				return err
			}
		}
	}
	return nil
}

// resolveNode follows document and alias nodes to the node holding the value.
func resolveNode(node *yaml.Node) *yaml.Node {
	for {
		switch {
		case node.Kind == yaml.DocumentNode && len(node.Content) == 1:
			node = node.Content[0]
		case node.Kind == yaml.AliasNode && node.Alias != nil:
			node = node.Alias
		default:
			return node
		}
	}
}

// mappingEntries returns the keys of a mapping in order of first appearance,
// with merge keys (<<) expanded in place, and the node holding each key's
// value. As in YAML merge semantics, an explicit key overrides merged ones and
// an earlier merged mapping overrides a later one.
func mappingEntries(node *yaml.Node) ([]string, []*yaml.Node, error) {
	var keys []string
	var values []*yaml.Node
	index := make(map[string]int)
	explicit := make(map[string]bool)
	add := func(key string, value *yaml.Node, isExplicit bool) {
		if i, ok := index[key]; ok {
			if isExplicit && !explicit[key] {
				values[i] = value
				explicit[key] = true
			}
			return
		}
		index[key] = len(keys)
		explicit[key] = isExplicit
		keys = append(keys, key)
		values = append(values, value)
	}

	for i := 0; i+1 < len(node.Content); i += 2 {
		keyNode, valueNode := resolveNode(node.Content[i]), node.Content[i+1]
		if keyNode.Kind == yaml.ScalarNode && keyNode.ShortTag() == "!!merge" {
			merged := resolveNode(valueNode)
			sources := []*yaml.Node{merged}
			if merged.Kind == yaml.SequenceNode {
				sources = merged.Content
			}
			for _, source := range sources {
				source = resolveNode(source)
				if source.Kind != yaml.MappingNode {
					continue
				}
				mergedKeys, mergedValues, err := mappingEntries(source)
				if err != nil {
					return nil, nil, err
				}
				for j, key := range mergedKeys {
					add(key, mergedValues[j], false)
				}
			}
			continue
		}
		key, err := nodeKey(keyNode)
		if err != nil {
			return nil, nil, err
		}
		add(key, valueNode, true)
	}
	return keys, values, nil
}

// nodeKey returns the map key a mapping key node becomes after normalizeTypes.
func nodeKey(node *yaml.Node) (string, error) {
	if node.Kind == yaml.ScalarNode && node.ShortTag() == "!!str" {
		return node.Value, nil
	}
	var key any
	if err := node.Decode(&key); err != nil {
		return "", err
	}
	return mapKey(key), nil
}

// setInterface sets the value of v to val
func setInterface(v any, val any) error {
	// Normalize types to be compatible with gojq/JSON
	normalized, err := normalizeTypes(val)
	if err != nil {
		return err
	}
	return assign(v, normalized)
}

// assign sets v, a pointer, to the normalized value.
func assign(v any, normalized any) error {
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
	if keyorder.Enabled() {
		node, err := orderedNode(v)
		if err != nil {
			return nil, err
		}
		v = node
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// orderedNode converts v to a yaml.Node whose mappings list keys in recorded
// order. Values other than maps and arrays are encoded as the encoder would.
func orderedNode(v any) (*yaml.Node, error) {
	switch val := v.(type) {
	case map[string]any:
		if val == nil {
			break
		}
		node := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		for _, key := range keyorder.Keys(val) {
			keyNode := &yaml.Node{}
			if err := keyNode.Encode(key); err != nil {
				return nil, err
			}
			valueNode, err := orderedNode(val[key])
			if err != nil {
				return nil, err
			}
			node.Content = append(node.Content, keyNode, valueNode)
		}
		return node, nil
	case []any:
		if val == nil {
			break
		}
		node := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		for _, item := range val {
			itemNode, err := orderedNode(item)
			if err != nil {
				return nil, err
			}
			node.Content = append(node.Content, itemNode)
		}
		return node, nil
	}
	node := &yaml.Node{}
	if err := node.Encode(v); err != nil {
		return nil, err
	}
	return node, nil
}
