package codec

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/goccy/go-json"
)

// StreamParser parses input in streaming mode, emitting path-value pairs via a channel
// For JSON: matches jq's --stream behavior exactly
// For other formats: converts each record/document to path-value pairs
func StreamParser(reader io.Reader, inputType EncodingType) (<-chan any, <-chan error) {
	dataChan := make(chan any, 100) // Buffer for performance
	errChan := make(chan error, 1)

	go func() {
		defer close(dataChan)
		defer close(errChan)

		var err error
		switch inputType {
		case JSON:
			err = streamJSON(reader, dataChan)
		case JSONL:
			err = streamJSONL(reader, dataChan)
		case YAML:
			err = streamYAML(reader, dataChan)
		case LINE, TXT:
			err = streamLines(reader, dataChan)
		case CSV:
			err = streamCSV(reader, dataChan)
		case TSV:
			err = streamTSV(reader, dataChan)
		default:
			// For unsupported formats, read all and convert to stream
			data, readErr := io.ReadAll(reader)
			if readErr != nil {
				errChan <- readErr
				return
			}
			var value any
			if unmarshalErr := Unmarshal(data, inputType, &value); unmarshalErr != nil {
				errChan <- unmarshalErr
				return
			}
			streamSlice := convertToStream(value, []any{})
			for _, item := range streamSlice {
				dataChan <- item
			}
			return
		}

		if err != nil {
			errChan <- err
		}
	}()

	return dataChan, errChan
}

// Legacy function that collects all stream elements (for backward compatibility)
func StreamParserCollect(reader io.Reader, inputType EncodingType) ([]any, error) {
	dataChan, errChan := StreamParser(reader, inputType)

	var result []any
	for item := range dataChan {
		result = append(result, item)
	}

	select {
	case err := <-errChan:
		return result, err
	default:
		return result, nil
	}
}

// streamJSON parses JSON in streaming mode, emitting to channel
func streamJSON(reader io.Reader, dataChan chan<- any) error {
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()

	// Parse and emit to channel
	err := parseStreamToChannel(decoder, []any{}, dataChan)
	if err != nil && err != io.EOF {
		return err
	}

	return nil
}

// streamJSONL parses JSONL (newline-delimited JSON) in streaming mode
func streamJSONL(reader io.Reader, dataChan chan<- any) error {
	scanner := bufio.NewScanner(reader)
	index := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var obj any
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			return fmt.Errorf("error parsing JSON on line %d: %v", index+1, err)
		}

		// Convert and emit each object immediately
		objStream := convertToStream(obj, []any{index})
		for _, item := range objStream {
			dataChan <- item
		}
		index++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading JSONL: %v", err)
	}

	return nil
}

// streamYAML parses YAML (including multi-document) in streaming mode
// Actually streams by parsing documents one at a time and emitting immediately
func streamYAML(reader io.Reader, dataChan chan<- any) error {
	scanner := bufio.NewScanner(reader)

	// Increase buffer size for larger lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var currentDoc strings.Builder
	docIndex := 0

	for scanner.Scan() {
		line := scanner.Text()

		// Check if this is a document separator
		if strings.TrimSpace(line) == "---" {
			// Parse and emit the previous document if we have one
			if currentDoc.Len() > 0 {
				var doc any
				if err := Unmarshal([]byte(currentDoc.String()), YAML, &doc); err != nil {
					return fmt.Errorf("error parsing YAML document %d: %v", docIndex, err)
				}

				// Convert and emit document immediately
				docStream := convertToStream(doc, []any{docIndex})
				for _, item := range docStream {
					dataChan <- item
				}
				docIndex++

				// Reset for next document
				currentDoc.Reset()
			}
			continue
		}

		// Accumulate lines for current document
		currentDoc.WriteString(line)
		currentDoc.WriteString("\n")
	}

	// Parse and emit the last document
	if currentDoc.Len() > 0 {
		var doc any
		if err := Unmarshal([]byte(currentDoc.String()), YAML, &doc); err != nil {
			return fmt.Errorf("error parsing YAML document %d: %v", docIndex, err)
		}

		docStream := convertToStream(doc, []any{docIndex})
		for _, item := range docStream {
			dataChan <- item
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading YAML: %v", err)
	}

	return nil
}

// streamLines parses line-delimited text in streaming mode
func streamLines(reader io.Reader, dataChan chan<- any) error {
	scanner := bufio.NewScanner(reader)
	index := 0

	for scanner.Scan() {
		line := scanner.Text()
		// Emit [index, line] for each line immediately
		dataChan <- []any{[]any{index}, line}
		index++
	}

	// Emit closing marker with last index
	if index > 0 {
		dataChan <- []any{[]any{index - 1}}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading lines: %v", err)
	}

	return nil
}

// streamCSV parses CSV in streaming mode
// Reads header once, then streams each row immediately
func streamCSV(reader io.Reader, dataChan chan<- any) error {
	scanner := bufio.NewScanner(reader)

	// Read header row
	if !scanner.Scan() {
		return fmt.Errorf("empty CSV file")
	}

	headerLine := scanner.Text()
	headers := strings.Split(headerLine, ",")

	// Trim spaces from headers
	for i := range headers {
		headers[i] = strings.TrimSpace(headers[i])
	}

	rowIndex := 0

	// Stream each data row immediately
	for scanner.Scan() {
		line := scanner.Text()
		values := strings.Split(line, ",")

		// Create object from row
		row := make(map[string]any)
		for i, header := range headers {
			if i < len(values) {
				row[header] = strings.TrimSpace(values[i])
			}
		}

		// Convert and emit row immediately
		rowStream := convertToStream(row, []any{rowIndex})
		for _, item := range rowStream {
			dataChan <- item
		}
		rowIndex++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading CSV: %v", err)
	}

	return nil
}

// streamTSV parses TSV in streaming mode
// Reads header once, then streams each row immediately
func streamTSV(reader io.Reader, dataChan chan<- any) error {
	scanner := bufio.NewScanner(reader)

	// Read header row
	if !scanner.Scan() {
		return fmt.Errorf("empty TSV file")
	}

	headerLine := scanner.Text()
	headers := strings.Split(headerLine, "\t")

	// Trim spaces from headers
	for i := range headers {
		headers[i] = strings.TrimSpace(headers[i])
	}

	rowIndex := 0

	// Stream each data row immediately
	for scanner.Scan() {
		line := scanner.Text()
		values := strings.Split(line, "\t")

		// Create object from row
		row := make(map[string]any)
		for i, header := range headers {
			if i < len(values) {
				row[header] = strings.TrimSpace(values[i])
			}
		}

		// Convert and emit row immediately
		rowStream := convertToStream(row, []any{rowIndex})
		for _, item := range rowStream {
			dataChan <- item
		}
		rowIndex++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading TSV: %v", err)
	}

	return nil
}

// convertToStream converts any value to streaming format (path-value pairs)
func convertToStream(value any, basePath []any) []any {
	var result []any

	switch v := value.(type) {
	case map[string]any:
		var lastKey string
		hasKeys := false
		for key, val := range v {
			lastKey = key
			hasKeys = true
			path := append(append([]any{}, basePath...), key)
			result = append(result, convertToStream(val, path)...)
		}
		// Emit closing marker
		if hasKeys {
			closePath := append(append([]any{}, basePath...), lastKey)
			result = append(result, []any{closePath})
		}

	case []any:
		lastIndex := -1
		for i, val := range v {
			lastIndex = i
			path := append(append([]any{}, basePath...), i)
			result = append(result, convertToStream(val, path)...)
		}
		// Emit closing marker
		if lastIndex >= 0 {
			closePath := append(append([]any{}, basePath...), lastIndex)
			result = append(result, []any{closePath})
		}

	default:
		// Leaf value - emit [path, value]
		pathCopy := make([]any, len(basePath))
		copy(pathCopy, basePath)
		result = append(result, []any{pathCopy, v})
	}

	return result
}

// parseStreamToChannel emits stream elements to a channel as they're generated
func parseStreamToChannel(decoder *json.Decoder, path []any, dataChan chan<- any) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}

	switch t := token.(type) {
	case json.Delim:
		switch t {
		case '[':
			// Start of array
			index := 0
			lastIndex := -1
			for decoder.More() {
				newPath := append(append([]any{}, path...), index)
				if err := parseStreamToChannel(decoder, newPath, dataChan); err != nil {
					return err
				}
				lastIndex = index
				index++
			}
			// Consume the closing ']'
			if _, err := decoder.Token(); err != nil {
				return err
			}
			// Emit closing marker
			if lastIndex >= 0 {
				closePath := append(append([]any{}, path...), lastIndex)
				dataChan <- []any{closePath}
			}
			return nil

		case '{':
			// Start of object
			var lastKey string
			hasKeys := false
			for decoder.More() {
				// Get the key
				keyToken, err := decoder.Token()
				if err != nil {
					return err
				}
				key, ok := keyToken.(string)
				if !ok {
					return fmt.Errorf("expected string key, got %T", keyToken)
				}
				lastKey = key
				hasKeys = true
				newPath := append(append([]any{}, path...), key)
				if err := parseStreamToChannel(decoder, newPath, dataChan); err != nil {
					return err
				}
			}
			// Consume the closing '}'
			if _, err := decoder.Token(); err != nil {
				return err
			}
			// Emit closing marker
			if hasKeys {
				closePath := append(append([]any{}, path...), lastKey)
				dataChan <- []any{closePath}
			}
			return nil

		default:
			return fmt.Errorf("unexpected delimiter: %v", t)
		}

	default:
		// Leaf value - emit [path, value]
		pathCopy := make([]any, len(path))
		copy(pathCopy, path)

		// Convert json.Number to appropriate type
		if num, ok := t.(json.Number); ok {
			if intVal, err := num.Int64(); err == nil {
				t = float64(intVal)
			} else if floatVal, err := num.Float64(); err == nil {
				t = floatVal
			}
		}

		dataChan <- []any{pathCopy, t}
		return nil
	}
}
