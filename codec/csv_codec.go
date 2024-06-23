package codec

import (
	"encoding/csv"
	"fmt"
	"github.com/goccy/go-json"
	"strings"
)

func csvUnmarshal(input []byte, v interface{}) error {
	r := csv.NewReader(strings.NewReader(string(input)))

	// Read the first row for headers
	headers, err := r.Read()
	if err != nil {
		return fmt.Errorf("error reading CSV headers: %v", err)
	}

	var records [][]string

	// Append headers as the first record
	records = append(records, headers)

	// Read the remaining rows
	for {
		record, err := r.Read()
		if err != nil {
			break
		}
		records = append(records, record)
	}

	// Marshal the records to JSON and then unmarshal into the provided interface
	jsonData, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("error marshaling to JSON: %v", err)
	}

	if err := json.Unmarshal(jsonData, v); err != nil {
		return fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	return nil
}
