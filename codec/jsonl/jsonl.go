package jsonl

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"

	qqjson "github.com/JFryy/qq/codec/json"
	"github.com/JFryy/qq/codec/keyorder"
	"github.com/goccy/go-json"
)

// Codec handles JSON Lines (newline-delimited JSON) format
type Codec struct{}

// Unmarshal parses JSONL data into a slice of objects
func (c *Codec) Unmarshal(data []byte, v any) error {
	if v == nil {
		return errors.New("v cannot be nil")
	}

	var result []any
	scanner := bufio.NewScanner(bytes.NewReader(data))

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		var obj any
		if err := qqjson.Unmarshal([]byte(line), &obj); err != nil {
			return fmt.Errorf("error parsing JSON on line %d: %v", lineNum, err)
		}
		result = append(result, obj)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading JSONL: %v", err)
	}

	// A JSON round trip would turn integers back into float64.
	if ptr, ok := v.(*any); ok {
		if result == nil {
			*ptr = nil
		} else {
			*ptr = result
		}
		return nil
	}

	// Marshal and unmarshal through JSON to convert to target type
	jsonData, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, v)
}

// Marshal converts data to JSONL format (one JSON object per line)
func (c *Codec) Marshal(v any) ([]byte, error) {
	items, err := lineItems(v)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	for i, item := range items {
		lineData, err := qqjson.MarshalCompact(item)
		if err != nil {
			return nil, fmt.Errorf("error marshaling item %d: %v", i, err)
		}
		buf.Write(lineData)
		buf.WriteByte('\n')
	}

	return buf.Bytes(), nil
}

// lineItems returns the values to write one per line: the elements of an
// array, or v itself.
func lineItems(v any) ([]any, error) {
	// A JSON round trip would lose recorded key order, and query results are
	// already plain JSON values.
	if keyorder.Enabled() {
		switch val := v.(type) {
		case []any:
			return val, nil
		case map[string]any:
			return []any{val}, nil
		}
	}

	// Convert to JSON first to normalize the data
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var decoded any
	if err := qqjson.Unmarshal(data, &decoded); err != nil {
		return nil, err
	}
	items, ok := decoded.([]any)
	if !ok {
		// A typed nil slice marshals to JSON null, which would otherwise be
		// treated as a single null item. Treat nil slices as empty arrays so
		// -k (key-preserved path returns slices directly) and the default path
		// produce the same number of JSONL records.
		if decoded == nil && v != nil {
			if rv := reflect.ValueOf(v); rv.Kind() == reflect.Slice && rv.IsNil() {
				return nil, nil
			}
		}
		items = []any{decoded}
	}
	return items, nil
}
