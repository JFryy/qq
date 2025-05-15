package csv

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/JFryy/qq/codec/util"
	"github.com/goccy/go-json"
	"io"
	"reflect"
	"slices"
	"strings"
)

type Codec struct{}

func (c *Codec) detectDelimiter(input []byte) rune {
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

func (c *Codec) Marshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return nil, errors.New("input data must be a slice")
	}

	if rv.Len() == 0 {
		return nil, errors.New("no data to write")
	}

	firstElem := rv.Index(0).Interface()
	firstElemValue, ok := firstElem.(map[string]any)
	if !ok {
		return nil, errors.New("slice elements must be of type map[string]any")
	}

	var headers []string
	for key := range firstElemValue {
		headers = append(headers, key)
	}
	slices.Sort(headers)

	if err := w.Write(headers); err != nil {
		return nil, fmt.Errorf("error writing CSV headers: %v", err)
	}

	for i := 0; i < rv.Len(); i++ {
		recordMap := rv.Index(i).Interface().(map[string]any)
		row := make([]string, len(headers))
		for j, header := range headers {
			if value, ok := recordMap[header]; ok {
				row[j] = fmt.Sprintf("%v", value)
			} else {
				row[j] = ""
			}
		}
		if err := w.Write(row); err != nil {
			return nil, fmt.Errorf("error writing CSV record: %v", err)
		}
	}

	w.Flush()

	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("error flushing CSV writer: %v", err)
	}

	return buf.Bytes(), nil
}

func (c *Codec) Unmarshal(input []byte, v any) error {
	delimiter := c.detectDelimiter(input)
	r := csv.NewReader(strings.NewReader(string(input)))
	r.Comma = delimiter
	r.TrimLeadingSpace = true
	headers, err := r.Read()
	if err != nil {
		return fmt.Errorf("error reading CSV headers: %v", err)
	}

	var records []map[string]any
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading CSV record: %v", err)
		}

		rowMap := make(map[string]any)
		for i, header := range headers {
			rowMap[header] = util.ParseValue(record[i])
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
