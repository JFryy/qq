package codec

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/goccy/go-json"
	"io"
	"strings"
)

func detectDelimiter(input []byte) rune {
	lines := bytes.Split(input, []byte("\n"))
	if len(lines) < 2 {
		return ','
	}

	delimiters := []rune{',', ';', '\t', '|', ' '}
	var maxDelimiter rune
	maxCount := 0

	for _, delimiter := range delimiters {
		count := strings.Count(string(lines[0]), string(delimiter))
		if count > maxCount {
			maxCount = count
			maxDelimiter = delimiter
		}
	}

	if maxCount == 0 {
		return ','
	}

	return maxDelimiter
}

func csvUnmarshal(input []byte, v interface{}) error {
	delimiter := detectDelimiter(input)
	r := csv.NewReader(strings.NewReader(string(input)))
	r.Comma = delimiter
	r.TrimLeadingSpace = true
	headers, err := r.Read()
	if err != nil {
		return fmt.Errorf("error reading CSV headers: %v", err)
	}

	var records []map[string]interface{}
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading CSV record: %v", err)
		}

		rowMap := make(map[string]interface{})
		for i, header := range headers {
			rowMap[header] = parseValue(record[i])
		}
		records = append(records, rowMap)
	}

	jsonData, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("error marshaling to JSON: %v", err)
	}

	if err := json.Unmarshal(jsonData, v); err != nil {
		return fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	return nil
}
