package yaml

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"

	"github.com/JFryy/qq/codec/util"
	"go.yaml.in/yaml/v4"
)

type Codec struct{}

// Unmarshal handles both single and multi-document YAML.
// For multi-document YAML (separated by ---), it returns an array of documents.
// For single-document YAML, it returns the document as-is.
func (c Codec) Unmarshal(data []byte, v any) error {
	if !util.PreserveKeyOrder {
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

	decoder := yaml.NewDecoder(bytes.NewReader(data))

	// Try to decode the first document as a node
	var firstNode yaml.Node
	err := decoder.Decode(&firstNode)
	if err != nil {
		return err
	}
	firstDoc, err := parseYAMLNode(&firstNode)
	if err != nil {
		return err
	}

	// Try to decode a second document to check if this is multi-document YAML
	var secondNode yaml.Node
	err = decoder.Decode(&secondNode)
	if err == io.EOF {
		// Only one document, return it directly
		return setInterface(v, firstDoc)
	}
	if err != nil {
		return err
	}
	secondDoc, err := parseYAMLNode(&secondNode)
	if err != nil {
		return err
	}

	// Multiple documents exist, collect them all into an array
	docs := []any{firstDoc, secondDoc}

	// Continue reading remaining documents
	for {
		var node yaml.Node
		err := decoder.Decode(&node)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		doc, err := parseYAMLNode(&node)
		if err != nil {
			return err
		}
		docs = append(docs, doc)
	}

	return setInterface(v, docs)
}

func parseYAMLNode(n *yaml.Node) (any, error) {
	switch n.Kind {
	case yaml.DocumentNode:
		if len(n.Content) == 0 {
			return nil, nil
		}
		return parseYAMLNode(n.Content[0])
	case yaml.MappingNode:
		m := make(map[string]any)
		var keyList []string
		for i := 0; i < len(n.Content); i += 2 {
			kNode := n.Content[i]
			vNode := n.Content[i+1]
			key := kNode.Value
			val, err := parseYAMLNode(vNode)
			if err != nil {
				return nil, err
			}
			m[key] = val
			keyList = append(keyList, key)
		}
		ptr := uintptr(reflect.ValueOf(m).UnsafePointer())
		util.SetKeyOrder(ptr, keyList)
		return m, nil
	case yaml.SequenceNode:
		arr := make([]any, len(n.Content))
		for i, item := range n.Content {
			val, err := parseYAMLNode(item)
			if err != nil {
				return nil, err
			}
			arr[i] = val
		}
		return arr, nil
	case yaml.ScalarNode:
		var val any
		if err := n.Decode(&val); err != nil {
			return nil, err
		}
		return val, nil
	default:
		return nil, fmt.Errorf("unknown node kind %d", n.Kind)
	}
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
		if util.PreserveKeyOrder {
			ptrOld := uintptr(reflect.ValueOf(v).UnsafePointer())
			if keyList, ok := util.GetKeyOrder(ptrOld); ok {
				ptrNew := uintptr(reflect.ValueOf(result).UnsafePointer())
				util.SetKeyOrder(ptrNew, keyList)
			}
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
	if !util.PreserveKeyOrder {
		var buf bytes.Buffer
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		if err := enc.Encode(v); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	node, err := toYAMLNode(v)
	if err != nil {
		return nil, err
	}

	var docNode *yaml.Node
	if node.Kind == yaml.DocumentNode {
		docNode = node
	} else {
		docNode = &yaml.Node{
			Kind:    yaml.DocumentNode,
			Content: []*yaml.Node{node},
		}
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(docNode); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func toYAMLNode(v any) (*yaml.Node, error) {
	if v == nil {
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!null",
			Value: "null",
		}, nil
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice {
		content := make([]*yaml.Node, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			node, err := toYAMLNode(rv.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			content[i] = node
		}
		return &yaml.Node{
			Kind:    yaml.SequenceNode,
			Tag:     "!!seq",
			Content: content,
		}, nil
	}

	switch val := v.(type) {
	case json.Number:
		if _, err := val.Int64(); err == nil {
			return &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!int",
				Value: val.String(),
			}, nil
		}
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!float",
			Value: val.String(),
		}, nil
	case map[string]any:
		// Determine key order
		ptr := uintptr(reflect.ValueOf(val).UnsafePointer())
		keys, ok := util.GetKeyOrder(ptr)
		if !ok || len(keys) != len(val) {
			keys = make([]string, 0, len(val))
			for k := range val {
				keys = append(keys, k)
			}
			sort.Strings(keys)
		} else {
			// Double check keys
			for _, k := range keys {
				if _, has := val[k]; !has {
					keys = make([]string, 0, len(val))
					for k2 := range val {
						keys = append(keys, k2)
					}
					sort.Strings(keys)
					break
				}
			}
		}

		content := make([]*yaml.Node, 0, len(keys)*2)
		for _, k := range keys {
			kNode := &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: k,
			}
			vNode, err := toYAMLNode(val[k])
			if err != nil {
				return nil, err
			}
			content = append(content, kNode, vNode)
		}
		return &yaml.Node{
			Kind:    yaml.MappingNode,
			Tag:     "!!map",
			Content: content,
		}, nil

	default:
		// Fallback: marshal/unmarshal using yaml.v4
		b, err := yaml.Marshal(val)
		if err != nil {
			return nil, err
		}
		var node yaml.Node
		if err := yaml.Unmarshal(b, &node); err != nil {
			return nil, err
		}
		if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
			return node.Content[0], nil
		}
		return &node, nil
	}
}
