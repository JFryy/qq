package avro

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/goccy/go-json"
	"github.com/hamba/avro/v2/ocf"
)

type Codec struct{}

func (c *Codec) Unmarshal(data []byte, v any) error {
	dec, err := ocf.NewDecoder(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error creating avro decoder: %v", err)
	}

	var records []map[string]any
	for dec.HasNext() {
		var record map[string]any
		if err := dec.Decode(&record); err != nil {
			return fmt.Errorf("error decoding avro record: %v", err)
		}
		records = append(records, record)
	}
	if err := dec.Error(); err != nil {
		return fmt.Errorf("avro decode error: %v", err)
	}

	// JSON roundtrip normalises numeric types (same approach as parquet codec)
	jsonData, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("error marshaling to JSON: %v", err)
	}
	return json.Unmarshal(jsonData, v)
}

func (c *Codec) Marshal(v any) ([]byte, error) {
	// Normalise to []map[string]any via JSON roundtrip
	jsonData, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var records []map[string]any
	if err := json.Unmarshal(jsonData, &records); err != nil {
		// Try single record
		var record map[string]any
		if err2 := json.Unmarshal(jsonData, &record); err2 != nil {
			return nil, fmt.Errorf("avro output requires an array of objects or a single object")
		}
		records = []map[string]any{record}
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no data to write")
	}

	schemaStr, err := inferSchema(records[0])
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	enc, err := ocf.NewEncoder(schemaStr, &buf)
	if err != nil {
		return nil, fmt.Errorf("error creating avro encoder: %v", err)
	}

	for _, record := range records {
		if err := enc.Encode(stringifyComplex(record)); err != nil {
			return nil, fmt.Errorf("error encoding avro record: %v", err)
		}
	}
	if err := enc.Flush(); err != nil {
		return nil, fmt.Errorf("error flushing avro encoder: %v", err)
	}

	return buf.Bytes(), nil
}

// stringifyComplex returns a copy of the record with any complex (non-scalar)
// values JSON-serialised as strings, matching the schema produced by inferSchema.
func stringifyComplex(record map[string]any) map[string]any {
	out := make(map[string]any, len(record))
	for k, v := range record {
		switch v.(type) {
		case nil, bool, float64, int64, string:
			out[k] = v
		default:
			b, err := json.Marshal(v)
			if err != nil {
				out[k] = fmt.Sprintf("%v", v)
			} else {
				out[k] = string(b)
			}
		}
	}
	return out
}

// inferAvroFieldType returns an Avro type string for a Go value.
// Complex types (objects, arrays) are stringified.
func inferAvroFieldType(v any) string {
	switch v.(type) {
	case bool:
		return `"boolean"`
	case float64:
		return `"double"`
	case int64:
		return `"long"`
	case string:
		return `"string"`
	default:
		return `"string"`
	}
}

// inferSchema builds an Avro record schema from the first record's keys/types.
// All fields are nullable (["null", <type>]) with a default of null.
func inferSchema(sample map[string]any) (string, error) {
	var fieldNames []string
	for k := range sample {
		fieldNames = append(fieldNames, k)
	}
	sort.Strings(fieldNames)

	var fields []string
	for _, k := range fieldNames {
		avroType := inferAvroFieldType(sample[k])
		fields = append(fields, fmt.Sprintf(`{"name":%q,"type":["null",%s],"default":null}`, k, avroType))
	}

	return fmt.Sprintf(`{"type":"record","name":"Root","fields":[%s]}`, strings.Join(fields, ",")), nil
}
