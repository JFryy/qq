package json

import (
	"bytes"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/JFryy/qq/codec/keyorder"
	"github.com/JFryy/qq/codec/util"
	"github.com/goccy/go-json"
)

type Codec struct{}

func (c *Codec) Marshal(v any) ([]byte, error) {
	if keyorder.Enabled() {
		return marshalOrdered(v, "  ", false)
	}
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
	if err := recordKeyOrder(data, []any{normalized}); err != nil {
		return err
	}
	*ptr = normalized
	return nil
}

// UnmarshalAll decodes a stream of whitespace-separated JSON values, keeping
// integers exact like Unmarshal.
func UnmarshalAll(data []byte) ([]any, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var values []any
	for {
		var value any
		if err := decoder.Decode(&value); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		normalized, err := NormalizeNumbers(value)
		if err != nil {
			return nil, err
		}
		values = append(values, normalized)
	}
	if err := recordKeyOrder(data, values); err != nil {
		return nil, err
	}
	return values, nil
}

// MarshalCompact encodes v on one line with HTML escaping, like json.Marshal,
// writing keys in recorded order when key order preservation is enabled.
func MarshalCompact(v any) ([]byte, error) {
	if keyorder.Enabled() {
		return marshalOrdered(v, "", true)
	}
	return json.Marshal(v)
}

// recordKeyOrder scans data, which has already been decoded into values, and
// records the key order of each decoded object. It does nothing unless key
// order recording is on.
func recordKeyOrder(data []byte, values []any) error {
	if !keyorder.Recording() {
		return nil
	}
	decoder := stdjson.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	for _, value := range values {
		if err := scanKeyOrder(decoder, value); err != nil {
			return fmt.Errorf("recording key order: %w", err)
		}
	}
	return nil
}

// scanKeyOrder reads one JSON value from decoder and records the key order of
// the objects in it on the matching maps in value. Where a key is repeated the
// decoded map holds the last value, which is also the last one recorded.
func scanKeyOrder(decoder *stdjson.Decoder, value any) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(stdjson.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		object, _ := value.(map[string]any)
		var keys []string
		for decoder.More() {
			token, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := token.(string)
			if !ok {
				return fmt.Errorf("expected object key, got %v", token)
			}
			keys = append(keys, key)
			if err := scanKeyOrder(decoder, object[key]); err != nil {
				return err
			}
		}
		keyorder.Record(object, keys)
	case '[':
		array, _ := value.([]any)
		for i := 0; decoder.More(); i++ {
			var item any
			if i < len(array) {
				item = array[i]
			}
			if err := scanKeyOrder(decoder, item); err != nil {
				return err
			}
		}
	}
	_, err = decoder.Token() // closing delimiter
	return err
}

// marshalOrdered encodes v like the go-json encoder with the given indent and
// HTML escaping, but writes object keys in recorded order.
func marshalOrdered(v any, indent string, escapeHTML bool) ([]byte, error) {
	var buf bytes.Buffer
	if err := writeOrdered(&buf, v, indent, "", escapeHTML); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeOrdered(buf *bytes.Buffer, v any, indent, prefix string, escapeHTML bool) error {
	inner := prefix + indent
	switch val := v.(type) {
	case map[string]any:
		if val == nil {
			break
		}
		if len(val) == 0 {
			buf.WriteString("{}")
			return nil
		}
		buf.WriteByte('{')
		for i, key := range keyorder.Keys(val) {
			if i > 0 {
				buf.WriteByte(',')
			}
			writeNewline(buf, indent, inner)
			if err := writeValue(buf, key, indent, inner, escapeHTML); err != nil {
				return err
			}
			buf.WriteByte(':')
			if indent != "" {
				buf.WriteByte(' ')
			}
			if err := writeOrdered(buf, val[key], indent, inner, escapeHTML); err != nil {
				return err
			}
		}
		writeNewline(buf, indent, prefix)
		buf.WriteByte('}')
		return nil
	case []any:
		if val == nil {
			break
		}
		if len(val) == 0 {
			buf.WriteString("[]")
			return nil
		}
		buf.WriteByte('[')
		for i, item := range val {
			if i > 0 {
				buf.WriteByte(',')
			}
			writeNewline(buf, indent, inner)
			if err := writeOrdered(buf, item, indent, inner, escapeHTML); err != nil {
				return err
			}
		}
		writeNewline(buf, indent, prefix)
		buf.WriteByte(']')
		return nil
	}
	return writeValue(buf, v, indent, prefix, escapeHTML)
}

func writeNewline(buf *bytes.Buffer, indent, prefix string) {
	if indent != "" {
		buf.WriteByte('\n')
		buf.WriteString(prefix)
	}
}

// writeValue encodes any value that is not an ordered container with go-json.
func writeValue(buf *bytes.Buffer, v any, indent, prefix string, escapeHTML bool) error {
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(escapeHTML)
	if indent != "" {
		encoder.SetIndent(prefix, indent)
	}
	if err := encoder.Encode(v); err != nil {
		return err
	}
	buf.Write(bytes.TrimSuffix(out.Bytes(), []byte("\n")))
	return nil
}

// NormalizeNumbers replaces json.Number values with int when they are whole
// numbers within int range, *big.Int for larger integer literals, and float64
// otherwise.
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
	// gojq keeps integers of any size exact, as jq does.
	if i, ok := util.ParseInteger(n.String()); ok {
		return util.NormalizeBigInt(i), nil
	}
	// Values that are not valid integers keep the previous float64 behaviour.
	return n.Float64()
}
