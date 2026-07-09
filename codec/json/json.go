package json

import (
	"bytes"
	"fmt"
	"math/big"
	"reflect"
	"sort"

	"github.com/JFryy/qq/codec/util"
	"github.com/goccy/go-json"
)

type Codec struct{}

func (c *Codec) Unmarshal(data []byte, v any) error {
	if !util.PreserveKeyOrder {
		return json.Unmarshal(data, v)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	val, err := parseJSONOrdered(dec)
	if err != nil {
		return err
	}
	return setInterface(v, val)
}

func (c *Codec) Marshal(v any) ([]byte, error) {
	if !util.PreserveKeyOrder {
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
	return marshalJSONOrdered(v, "  ", "")
}

func setInterface(v any, val any) error {
	switch ptr := v.(type) {
	case *any:
		*ptr = val
		return nil
	default:
		// Fallback to standard json Unmarshal for specific concrete types
		b, err := json.Marshal(val)
		if err != nil {
			return err
		}
		return json.Unmarshal(b, v)
	}
}

func parseJSONOrdered(dec *json.Decoder) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			m := make(map[string]any)
			var keyList []string
			for dec.More() {
				keyTok, err := dec.Token()
				if err != nil {
					return nil, err
				}
				key, ok := keyTok.(string)
				if !ok {
					return nil, fmt.Errorf("expected string key, got %T", keyTok)
				}
				val, err := parseJSONOrdered(dec)
				if err != nil {
					return nil, err
				}
				m[key] = val
				keyList = append(keyList, key)
			}
			if _, err := dec.Token(); err != nil { // consume '}'
				return nil, err
			}
			ptr := uintptr(reflect.ValueOf(m).UnsafePointer())
			util.SetKeyOrder(ptr, keyList)
			return m, nil
		case '[':
			var arr []any
			for dec.More() {
				val, err := parseJSONOrdered(dec)
				if err != nil {
					return nil, err
				}
				arr = append(arr, val)
			}
			if _, err := dec.Token(); err != nil { // consume ']'
				return nil, err
			}
			return arr, nil
		default:
			return nil, fmt.Errorf("unexpected delimiter %v", t)
		}
	default:
		return t, nil
	}
}

func marshalJSONOrdered(v any, indent string, currentIndent string) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice {
		if rv.Len() == 0 {
			return []byte("[]"), nil
		}
		var buf bytes.Buffer
		buf.WriteByte('[')
		nextIndent := currentIndent + indent
		for i := 0; i < rv.Len(); i++ {
			if i > 0 {
				buf.WriteByte(',')
			}
			if indent != "" {
				buf.WriteByte('\n')
				buf.WriteString(nextIndent)
			}
			elemBytes, err := marshalJSONOrdered(rv.Index(i).Interface(), indent, nextIndent)
			if err != nil {
				return nil, err
			}
			buf.Write(elemBytes)
		}
		if indent != "" {
			buf.WriteByte('\n')
			buf.WriteString(currentIndent)
		}
		buf.WriteByte(']')
		return buf.Bytes(), nil
	}

	switch val := v.(type) {
	case bool:
		if val {
			return []byte("true"), nil
		}
		return []byte("false"), nil

	case string, int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8, float64, float32, json.Number, *big.Int:
		// Delegate primitive marshaling to go-json
		return json.Marshal(val)

	case map[string]any:
		if len(val) == 0 {
			return []byte("{}"), nil
		}
		var buf bytes.Buffer
		buf.WriteByte('{')

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
			// Double check all keys in keys are actually in map (just in case)
			for _, k := range keys {
				if _, has := val[k]; !has {
					// Fallback to sorted keys
					keys = make([]string, 0, len(val))
					for k2 := range val {
						keys = append(keys, k2)
					}
					sort.Strings(keys)
					break
				}
			}
		}

		nextIndent := currentIndent + indent
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			if indent != "" {
				buf.WriteByte('\n')
				buf.WriteString(nextIndent)
			}
			// Write key
			keyBytes, err := json.Marshal(k)
			if err != nil {
				return nil, err
			}
			buf.Write(keyBytes)
			if indent != "" {
				buf.WriteString(": ")
			} else {
				buf.WriteByte(':')
			}
			// Write value
			valBytes, err := marshalJSONOrdered(val[k], indent, nextIndent)
			if err != nil {
				return nil, err
			}
			buf.Write(valBytes)
		}
		if indent != "" {
			buf.WriteByte('\n')
			buf.WriteString(currentIndent)
		}
		buf.WriteByte('}')
		return buf.Bytes(), nil

	default:
		// Fallback for any other type
		return json.Marshal(val)
	}
}
