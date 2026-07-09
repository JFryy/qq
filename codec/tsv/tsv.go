package tsv

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"reflect"
	"slices"
	"strings"

	qqjson "github.com/JFryy/qq/codec/json"
	"github.com/JFryy/qq/codec/util"
	"github.com/goccy/go-json"
)

type Codec struct{}

func (c *Codec) Marshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Comma = '\t' // Use tab as delimiter

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
	ptr := uintptr(reflect.ValueOf(firstElemValue).UnsafePointer())
	if util.PreserveKeyOrder {
		if list, ok := util.GetKeyOrder(ptr); ok && len(list) == len(firstElemValue) {
			headers = list
		}
	}

	if len(headers) == 0 {
		for key := range firstElemValue {
			headers = append(headers, key)
		}
		slices.Sort(headers)
	}

	if err := w.Write(headers); err != nil {
		return nil, fmt.Errorf("error writing TSV headers: %v", err)
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
			return nil, fmt.Errorf("error writing TSV record: %v", err)
		}
	}

	w.Flush()

	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("error flushing TSV writer: %v", err)
	}

	return buf.Bytes(), nil
}

func (c *Codec) Unmarshal(input []byte, v any) error {
	r := csv.NewReader(strings.NewReader(string(input)))
	r.Comma = '\t' // Use tab as delimiter
	r.LazyQuotes = true
	r.FieldsPerRecord = -1 // Allow variable number of fields

	headers, err := r.Read()
	if err != nil {
		return fmt.Errorf("error reading TSV headers: %v", err)
	}

	var records []any
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading TSV record: %v", err)
		}

		rowMap := make(map[string]any)
		for i, header := range headers {
			if i < len(record) {
				rowMap[header] = util.ParseValue(record[i])
			} else {
				rowMap[header] = ""
			}
		}
		if util.PreserveKeyOrder {
			ptr := uintptr(reflect.ValueOf(rowMap).UnsafePointer())
			util.SetKeyOrder(ptr, headers)
		}
		records = append(records, rowMap)
	}

	jsonCodec := &qqjson.Codec{}
	jsonData, err := jsonCodec.MarshalCompact(records)
	if err != nil {
		return fmt.Errorf("error marshaling to JSON: %v", err)
	}

	if util.PreserveKeyOrder {
		return jsonCodec.Unmarshal(jsonData, v)
	}
	return json.Unmarshal(jsonData, v)
}
