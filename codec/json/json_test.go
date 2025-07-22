package json

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestJSONMarshalBasicTypes(t *testing.T) {
	testData := map[string]any{
		"string":  "hello",
		"number":  42,
		"float":   3.14,
		"boolean": true,
		"null":    nil,
	}

	codec := &Codec{}

	// Test marshaling
	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal JSON data: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Marshaled data is empty")
	}

	// Verify it's valid JSON
	var result map[string]any
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Marshaled data is not valid JSON: %v", err)
	}

	// Verify values
	if result["string"] != "hello" {
		t.Errorf("Expected string 'hello', got %v", result["string"])
	}
	if result["number"].(float64) != 42 {
		t.Errorf("Expected number 42, got %v", result["number"])
	}
	if result["boolean"] != true {
		t.Errorf("Expected boolean true, got %v", result["boolean"])
	}
	if result["null"] != nil {
		t.Errorf("Expected null nil, got %v", result["null"])
	}
}

func TestJSONMarshalArray(t *testing.T) {
	testData := []map[string]any{
		{"id": 1, "name": "Alice"},
		{"id": 2, "name": "Bob"},
	}

	codec := &Codec{}

	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal JSON array: %v", err)
	}

	// Verify it's valid JSON array
	var result []map[string]any
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Marshaled array is not valid JSON: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(result))
	}

	if result[0]["name"] != "Alice" {
		t.Errorf("Expected first name 'Alice', got %v", result[0]["name"])
	}
}

func TestJSONMarshalNested(t *testing.T) {
	testData := map[string]any{
		"user": map[string]any{
			"name": "Alice",
			"metadata": map[string]any{
				"department": "Engineering",
				"level":      "Senior",
			},
		},
		"tags": []string{"golang", "json"},
	}

	codec := &Codec{}

	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal nested JSON: %v", err)
	}

	// Verify it's valid JSON
	var result map[string]any
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Marshaled nested data is not valid JSON: %v", err)
	}

	// Navigate nested structure
	user := result["user"].(map[string]any)
	metadata := user["metadata"].(map[string]any)

	if metadata["department"] != "Engineering" {
		t.Errorf("Expected department 'Engineering', got %v", metadata["department"])
	}

	tags := result["tags"].([]any)
	if len(tags) != 2 || tags[0] != "golang" {
		t.Errorf("Expected tags array with 'golang', got %v", tags)
	}
}

func TestJSONMarshalSpecialCharacters(t *testing.T) {
	testData := map[string]any{
		"quotes":    "He said \"Hello\"",
		"backslash": "Path\\to\\file",
		"unicode":   "Hello 世界 🌍",
		"newline":   "Line 1\nLine 2",
		"tab":       "Col1\tCol2",
	}

	codec := &Codec{}

	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal JSON with special chars: %v", err)
	}

	// Verify it's valid JSON
	var result map[string]any
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Marshaled data with special chars is not valid JSON: %v", err)
	}

	// Verify special characters are preserved
	if result["quotes"] != "He said \"Hello\"" {
		t.Errorf("Quotes not preserved: %v", result["quotes"])
	}
	if result["unicode"] != "Hello 世界 🌍" {
		t.Errorf("Unicode not preserved: %v", result["unicode"])
	}
}

func TestJSONMarshalFormatting(t *testing.T) {
	testData := map[string]any{
		"key1": "value1",
		"key2": "value2",
	}

	codec := &Codec{}

	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	jsonStr := string(data)

	// Should be pretty-printed (indented)
	if !strings.Contains(jsonStr, "  ") {
		t.Error("JSON should be indented but appears to be compact")
	}

	// Should not have trailing newline (trimmed)
	if strings.HasSuffix(jsonStr, "\n") {
		t.Error("JSON should not have trailing newline")
	}

	// Should not escape HTML (SetEscapeHTML(false))
	testDataWithHTML := map[string]any{
		"html": "<div>test</div>",
	}

	dataHTML, err := codec.Marshal(testDataWithHTML)
	if err != nil {
		t.Fatalf("Failed to marshal JSON with HTML: %v", err)
	}

	if !strings.Contains(string(dataHTML), "<div>") {
		t.Error("HTML should not be escaped")
	}
}

func TestJSONMarshalEmptyValues(t *testing.T) {
	testCases := []struct {
		name string
		data any
	}{
		{"empty map", map[string]any{}},
		{"empty array", []any{}},
		{"empty string", ""},
		{"zero number", 0},
		{"false boolean", false},
	}

	codec := &Codec{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := codec.Marshal(tc.data)
			if err != nil {
				t.Fatalf("Failed to marshal %s: %v", tc.name, err)
			}

			// Should be valid JSON
			var result any
			err = json.Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Marshaled %s is not valid JSON: %v", tc.name, err)
			}
		})
	}
}
