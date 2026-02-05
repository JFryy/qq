package jsonl

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			return fmt.Errorf("error parsing JSON on line %d: %v", lineNum, err)
		}
		result = append(result, obj)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading JSONL: %v", err)
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
	// Convert to JSON first to normalize the data
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var items []any
	if err := json.Unmarshal(data, &items); err != nil {
		// If it's not an array, wrap it in an array
		var singleItem any
		if err := json.Unmarshal(data, &singleItem); err != nil {
			return nil, err
		}
		items = []any{singleItem}
	}

	var buf bytes.Buffer
	for i, item := range items {
		lineData, err := json.Marshal(item)
		if err != nil {
			return nil, fmt.Errorf("error marshaling item %d: %v", i, err)
		}
		buf.Write(lineData)
		buf.WriteByte('\n')
	}

	return buf.Bytes(), nil
}
