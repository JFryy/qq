package codec

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

func TestGetEncodingType(t *testing.T) {
	tests := []struct {
		input    string
		expected EncodingType
	}{
		{"json", JSON},
		{"yaml", YAML},
		{"yml", YAML}, // yml maps to YAML
		{"toml", TOML},
		{"hcl", HCL},
		{"tf", HCL}, // tf maps to HCL
		{"csv", CSV},
		{"xml", XML},
		{"ini", INI},
		{"gron", GRON},
		{"cbor", CBOR},
		{"avro", AVRO},
		//		{"html", HTML},
	}

	for _, tt := range tests {
		result, err := GetEncodingType(tt.input)
		if err != nil {
			t.Errorf("unexpected error for type %s: %v", tt.input, err)
		} else if result != tt.expected {
			t.Errorf("expected %v, got %v", tt.expected, result)
		}
	}

	unsupportedResult, err := GetEncodingType("unsupported")
	if err == nil {
		t.Errorf("expected error for unsupported type, got result: %v", unsupportedResult)
	}
}

func TestMarshal(t *testing.T) {
	data := map[string]any{"key": "value"}
	tests := []struct {
		encodingType EncodingType
	}{
		{JSON}, {YAML}, {TOML}, {HCL}, {CSV}, {XML}, {INI}, {GRON}, {HTML},
	}

	for _, tt := range tests {
		// wrap in an interface for things like CSV that require the basic test data be a []map[string]any
		var currentData any
		currentData = data
		if tt.encodingType == CSV {
			currentData = []any{data}
		}

		_, err := Marshal(currentData, tt.encodingType)
		if err != nil {
			t.Errorf("marshal failed for %v: %v", tt.encodingType, err)
		}
	}
}

func TestUnmarshal(t *testing.T) {
	jsonData := `{"key": "value"}`
	xmlData := `<root><key>value</key></root>`
	yamlData := "key: value"
	tomlData := "key = \"value\""
	gronData := `key = "value";`
	tfData := `key = "value"`
	// note: html and csv tests are not yet functional
	//	htmlData := `<html><body><key>value</key></body></html>`
	//	csvData := "key1,key2\nvalue1,value2\nvalue3,value4"

	tests := []struct {
		input        []byte
		encodingType EncodingType
		expected     any
	}{
		{[]byte(jsonData), JSON, map[string]any{"key": "value"}},
		{[]byte(xmlData), XML, map[string]any{"root": map[string]any{"key": "value"}}},
		{[]byte(yamlData), YAML, map[string]any{"key": "value"}},
		{[]byte(tomlData), TOML, map[string]any{"key": "value"}},
		{[]byte(gronData), GRON, map[string]any{"key": "value"}},
		{[]byte(tfData), HCL, map[string]any{"key": "value"}},
		//		{[]byte(htmlData), HTML, map[string]any{"html": map[string]any{"body": map[string]any{"key": "value"}}}},
		//		{[]byte(csvData), CSV, []map[string]any{
		//			{"key1": "value1", "key2": "value2"},
		//			{"key1": "value3", "key2": "value4"},
		//		}},
	}

	for _, tt := range tests {
		var data any
		err := Unmarshal(tt.input, tt.encodingType, &data)
		if err != nil {
			t.Errorf("unmarshal failed for %v: %v", tt.encodingType, err)
		}

		expectedJSON, _ := json.Marshal(tt.expected)
		actualJSON, _ := json.Marshal(data)

		if !reflect.DeepEqual(data, tt.expected) {
			fmt.Printf("expected: %s\n", string(expectedJSON))
			fmt.Printf("got: %s\n", string(actualJSON))
			t.Errorf("%s: expected %v, got %v", tt.encodingType, tt.expected, data)
		}
	}
}

func TestPrettyFormatRawOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		fileType EncodingType
		expected string
	}{
		{
			name:     "string with tab escape sequence",
			input:    `"string_1\tsecond string"`,
			fileType: JSON,
			expected: "string_1\tsecond string", // actual tab character
		},
		{
			name:     "string with newline escape sequence",
			input:    `"line1\nline2"`,
			fileType: JSON,
			expected: "line1\nline2", // actual newline
		},
		{
			name:     "string with backslash",
			input:    `"path\\to\\file"`,
			fileType: JSON,
			expected: "path\\to\\file", // actual backslashes
		},
		{
			name:     "string with quotes",
			input:    `"say \"hello\""`,
			fileType: JSON,
			expected: `say "hello"`, // actual quotes
		},
		{
			name:     "simple string",
			input:    `"hello"`,
			fileType: JSON,
			expected: "hello",
		},
		{
			name:     "number should stay unchanged",
			input:    `42`,
			fileType: JSON,
			expected: "42",
		},
		{
			name:     "boolean should stay unchanged",
			input:    `true`,
			fileType: JSON,
			expected: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := PrettyFormat(tt.input, tt.fileType, true, true)
			if err != nil {
				t.Fatalf("PrettyFormat failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPrettyFormatRawOutputMapAndArray(t *testing.T) {
	// Maps and arrays should not be stripped of quotes in raw mode
	tests := []struct {
		name     string
		input    string
		fileType EncodingType
	}{
		{
			name:     "object",
			input:    `{"key": "value"}`,
			fileType: JSON,
		},
		{
			name:     "array",
			input:    `["a", "b", "c"]`,
			fileType: JSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := PrettyFormat(tt.input, tt.fileType, true, true)
			if err != nil {
				t.Fatalf("PrettyFormat failed: %v", err)
			}
			// Result should be unchanged (not stripped)
			if result != tt.input {
				t.Errorf("expected %q to remain unchanged, got %q", tt.input, result)
			}
		})
	}
}

func TestPrettyFormatMonochrome(t *testing.T) {
	input := `{"key": "value", "number": 42}`

	result, err := PrettyFormat(input, JSON, false, true)
	if err != nil {
		t.Fatalf("PrettyFormat failed: %v", err)
	}

	// Monochrome output should not contain ANSI escape codes
	if strings.Contains(result, "\033[") || strings.Contains(result, "\x1b[") {
		t.Errorf("Monochrome output should not contain ANSI escape codes, got: %q", result)
	}
}

func TestPrettyFormatWithColors(t *testing.T) {
	input := `{"key": "value", "number": 42}`

	result, err := PrettyFormat(input, JSON, false, false)
	if err != nil {
		t.Fatalf("PrettyFormat failed: %v", err)
	}

	// Since we're not in a TTY during tests, colors should be disabled
	// and output should be plain
	if strings.Contains(result, "\033[") || strings.Contains(result, "\x1b[") {
		// If colors are present, that's actually OK - it means TTY detection
		// thinks we're in a terminal
		t.Logf("Colors detected in output (TTY might be detected)")
	}

	// Result should contain the input data
	if !strings.Contains(result, "key") || !strings.Contains(result, "value") {
		t.Errorf("Output should contain the input data")
	}
}

func TestCBORRoundTrip(t *testing.T) {
	original := map[string]any{"name": "alice", "active": true, "score": float64(42)}

	encoded, err := Marshal(original, CBOR)
	if err != nil {
		t.Fatalf("cbor marshal failed: %v", err)
	}

	var decoded any
	if err := Unmarshal(encoded, CBOR, &decoded); err != nil {
		t.Fatalf("cbor unmarshal failed: %v", err)
	}

	encodedJSON, _ := json.Marshal(original)
	decodedJSON, _ := json.Marshal(decoded)
	if string(encodedJSON) != string(decodedJSON) {
		t.Errorf("cbor round-trip mismatch:\n  want: %s\n   got: %s", encodedJSON, decodedJSON)
	}
}

func TestAvroRoundTrip(t *testing.T) {
	original := []any{
		map[string]any{"name": "alice", "score": float64(1)},
		map[string]any{"name": "bob", "score": float64(2)},
	}

	encoded, err := Marshal(original, AVRO)
	if err != nil {
		t.Fatalf("avro marshal failed: %v", err)
	}

	var decoded any
	if err := Unmarshal(encoded, AVRO, &decoded); err != nil {
		t.Fatalf("avro unmarshal failed: %v", err)
	}

	encodedJSON, _ := json.Marshal(original)
	decodedJSON, _ := json.Marshal(decoded)
	if string(encodedJSON) != string(decodedJSON) {
		t.Errorf("avro round-trip mismatch:\n  want: %s\n   got: %s", encodedJSON, decodedJSON)
	}
}

func TestIsBinaryFormat(t *testing.T) {
	tests := []struct {
		format   EncodingType
		expected bool
	}{
		{PARQUET, true},
		{MSGPACK, true},
		{CBOR, true},
		{AVRO, true},
		{JSON, false},
		{YAML, false},
		{XML, false},
		{CSV, false},
	}

	for _, tt := range tests {
		t.Run(tt.format.String(), func(t *testing.T) {
			result := IsBinaryFormat(tt.format)
			if result != tt.expected {
				t.Errorf("IsBinaryFormat(%v) = %v, expected %v", tt.format, result, tt.expected)
			}
		})
	}
}

func TestKeyOrderPreservation(t *testing.T) {
	// Test default behavior (PreserveKeyOrder = false) -> alphabetical sorting
	SetPreserveKeyOrder(false)

	jsonInput := `{"z":1,"a":2,"m":3}`
	var valJSON any
	if err := Unmarshal([]byte(jsonInput), JSON, &valJSON); err != nil {
		t.Fatalf("JSON Unmarshal failed: %v", err)
	}

	marshaledJSON, err := Marshal(valJSON, JSON)
	if err != nil {
		t.Fatalf("JSON Marshal failed: %v", err)
	}
	expectedSortedJSON := `{
  "a": 2,
  "m": 3,
  "z": 1
}`
	if string(marshaledJSON) != expectedSortedJSON {
		t.Errorf("Default JSON should sort keys alphabetically:\n  got: %s\n  want: %s", string(marshaledJSON), expectedSortedJSON)
	}

	// Test flag enabled behavior (PreserveKeyOrder = true) -> preserve order
	SetPreserveKeyOrder(true)
	defer SetPreserveKeyOrder(false) // Reset for other tests

	var valJSONPreserved any
	if err := Unmarshal([]byte(jsonInput), JSON, &valJSONPreserved); err != nil {
		t.Fatalf("JSON Unmarshal failed: %v", err)
	}

	marshaledJSONPreserved, err := Marshal(valJSONPreserved, JSON)
	if err != nil {
		t.Fatalf("JSON Marshal failed: %v", err)
	}
	expectedPreservedJSON := `{
  "z": 1,
  "a": 2,
  "m": 3
}`
	if string(marshaledJSONPreserved) != expectedPreservedJSON {
		t.Errorf("JSON key order not preserved with flag:\n  got: %s\n  want: %s", string(marshaledJSONPreserved), expectedPreservedJSON)
	}

	// Test YAML key order preservation
	yamlInput := `z: 1
a: 2
m: 3
`
	var valYAML any
	if err := Unmarshal([]byte(yamlInput), YAML, &valYAML); err != nil {
		t.Fatalf("YAML Unmarshal failed: %v", err)
	}

	marshaledYAML, err := Marshal(valYAML, YAML)
	if err != nil {
		t.Fatalf("YAML Marshal failed: %v", err)
	}
	expectedPreservedYAML := `z: 1
a: 2
m: 3
`
	if string(marshaledYAML) != expectedPreservedYAML {
		t.Errorf("YAML key order not preserved with flag:\n  got: %s\n  want: %s", string(marshaledYAML), expectedPreservedYAML)
	}

	// Test CSV key/column order preservation
	csvInput := "z,a,m\n1,2,3\n"
	var valCSV any
	if err := Unmarshal([]byte(csvInput), CSV, &valCSV); err != nil {
		t.Fatalf("CSV Unmarshal failed: %v", err)
	}

	marshaledCSV, err := Marshal(valCSV, CSV)
	if err != nil {
		t.Fatalf("CSV Marshal failed: %v", err)
	}
	expectedPreservedCSV := "z,a,m\n1,2,3\n"
	if string(marshaledCSV) != expectedPreservedCSV {
		t.Errorf("CSV column order not preserved with flag:\n  got: %s\n  want: %s", string(marshaledCSV), expectedPreservedCSV)
	}

	// Test TSV key/column order preservation
	tsvInput := "z\ta\tm\n1\t2\t3\n"
	var valTSV any
	if err := Unmarshal([]byte(tsvInput), TSV, &valTSV); err != nil {
		t.Fatalf("TSV Unmarshal failed: %v", err)
	}

	marshaledTSV, err := Marshal(valTSV, TSV)
	if err != nil {
		t.Fatalf("TSV Marshal failed: %v", err)
	}
	expectedPreservedTSV := "z\ta\tm\n1\t2\t3\n"
	if string(marshaledTSV) != expectedPreservedTSV {
		t.Errorf("TSV column order not preserved with flag:\n  got: %s\n  want: %s", string(marshaledTSV), expectedPreservedTSV)
	}

	// Test JSON to CSV column preservation
	jsonSliceInput := `[{"z":1,"a":2,"m":3}]`
	var valJSONSlice any
	if err := Unmarshal([]byte(jsonSliceInput), JSON, &valJSONSlice); err != nil {
		t.Fatalf("JSON Slice Unmarshal failed: %v", err)
	}
	marshaledJSONtoCSV, err := Marshal(valJSONSlice, CSV)
	if err != nil {
		t.Fatalf("JSON to CSV Marshal failed: %v", err)
	}
	expectedJSONtoCSV := "z,a,m\n1,2,3\n"
	if string(marshaledJSONtoCSV) != expectedJSONtoCSV {
		t.Errorf("JSON to CSV column order not preserved:\n  got: %s\n  want: %s", string(marshaledJSONtoCSV), expectedJSONtoCSV)
	}

	// Test CSV to JSON key preservation
	var valCSVtoJSON any
	if err := Unmarshal([]byte(expectedJSONtoCSV), CSV, &valCSVtoJSON); err != nil {
		t.Fatalf("CSV Unmarshal failed: %v", err)
	}
	marshaledCSVtoJSON, err := Marshal(valCSVtoJSON, JSON)
	if err != nil {
		t.Fatalf("CSV to JSON Marshal failed: %v", err)
	}
	expectedCSVtoJSON := `[
  {
    "z": 1,
    "a": 2,
    "m": 3
  }
]`
	if string(marshaledCSVtoJSON) != expectedCSVtoJSON {
		t.Errorf("CSV to JSON key order not preserved:\n  got: %s\n  want: %s", string(marshaledCSVtoJSON), expectedCSVtoJSON)
	}

	// Test TSV to CSV preservation
	var valTSVtoCSV any
	if err := Unmarshal([]byte(tsvInput), TSV, &valTSVtoCSV); err != nil {
		t.Fatalf("TSV Unmarshal failed: %v", err)
	}
	marshaledTSVtoCSV, err := Marshal(valTSVtoCSV, CSV)
	if err != nil {
		t.Fatalf("TSV to CSV Marshal failed: %v", err)
	}
	expectedTSVtoCSV := "z,a,m\n1,2,3\n"
	if string(marshaledTSVtoCSV) != expectedTSVtoCSV {
		t.Errorf("TSV to CSV column order not preserved:\n  got: %s\n  want: %s", string(marshaledTSVtoCSV), expectedTSVtoCSV)
	}

	// Test CSV to TSV preservation
	var valCSVtoTSV any
	if err := Unmarshal([]byte(csvInput), CSV, &valCSVtoTSV); err != nil {
		t.Fatalf("CSV Unmarshal failed: %v", err)
	}
	marshaledCSVtoTSV, err := Marshal(valCSVtoTSV, TSV)
	if err != nil {
		t.Fatalf("CSV to TSV Marshal failed: %v", err)
	}
	expectedCSVtoTSV := "z\ta\tm\n1\t2\t3\n"
	if string(marshaledCSVtoTSV) != expectedCSVtoTSV {
		t.Errorf("CSV to TSV column order not preserved:\n  got: %s\n  want: %s", string(marshaledCSVtoTSV), expectedCSVtoTSV)
	}

	// Test XML to XML preservation
	xmlInput := `<root>
  <z>1</z>
  <a>2</a>
  <m>
    <y>10</y>
    <x>20</x>
  </m>
</root>`
	var valXML any
	if err := Unmarshal([]byte(xmlInput), XML, &valXML); err != nil {
		t.Fatalf("XML Unmarshal failed: %v", err)
	}
	marshaledXML, err := Marshal(valXML, XML)
	if err != nil {
		t.Fatalf("XML Marshal failed: %v", err)
	}
	if string(marshaledXML) != xmlInput {
		t.Errorf("XML to XML order not preserved:\n  got: %s\n  want: %s", string(marshaledXML), xmlInput)
	}
}
