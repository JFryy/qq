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

func TestIsBinaryFormat(t *testing.T) {
	tests := []struct {
		format   EncodingType
		expected bool
	}{
		{PARQUET, true},
		{MSGPACK, true},
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
