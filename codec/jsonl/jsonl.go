package jsonl

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"

	qqjson "github.com/JFryy/qq/codec/json"
	"github.com/JFryy/qq/codec/util"
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

	jsonCodec := &qqjson.Codec{}

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		var obj any
		var err error
		if util.PreserveKeyOrder {
			err = jsonCodec.Unmarshal([]byte(line), &obj)
		} else {
			err = json.Unmarshal([]byte(line), &obj)
		}
		if err != nil {
			return fmt.Errorf("error parsing JSON on line %d: %v", lineNum, err)
		}
		result = append(result, obj)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading JSONL: %v", err)
	}

	jsonData, err := jsonCodec.MarshalCompact(result)
	if err != nil {
		return err
	}

	if util.PreserveKeyOrder {
		return jsonCodec.Unmarshal(jsonData, v)
	}
	return json.Unmarshal(jsonData, v)
}

// Marshal converts data to JSONL format (one JSON object per line)
func (c *Codec) Marshal(v any) ([]byte, error) {
	var items []any
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice {
		items = make([]any, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			items[i] = rv.Index(i).Interface()
		}
	} else {
		items = []any{v}
	}

	jsonCodec := &qqjson.Codec{}

	var buf bytes.Buffer
	for i, item := range items {
		var lineData []byte
		var err error
		if util.PreserveKeyOrder {
			lineData, err = jsonCodec.MarshalCompact(item)
		} else {
			lineData, err = json.Marshal(item)
		}
		if err != nil {
			return nil, fmt.Errorf("error marshaling item %d: %v", i, err)
		}
		buf.Write(lineData)
		buf.WriteByte('\n')
	}

	return buf.Bytes(), nil
}
