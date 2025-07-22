package xml

import (
	"strings"
	"testing"
)

func TestBasicXMLMarshalUnmarshal(t *testing.T) {
	testData := map[string]any{
		"name":   "Alice",
		"age":    30,
		"active": true,
		"score":  95.5,
	}

	codec := &Codec{}

	// Test marshaling
	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal XML data: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Marshaled data is empty")
	}

	// Check that it contains XML structure
	xmlStr := string(data)
	if !strings.Contains(xmlStr, "<root>") || !strings.Contains(xmlStr, "</root>") {
		t.Error("XML output missing root element")
	}

	// Test unmarshaling
	var result map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal XML data: %v", err)
	}

	// Verify basic fields - XML converts everything to strings
	if result["name"] != "Alice" {
		t.Errorf("Expected name 'Alice', got %v", result["name"])
	}
	if result["age"] != "30" {
		t.Errorf("Expected age '30', got %v", result["age"])
	}
	if result["active"] != "true" {
		t.Errorf("Expected active 'true', got %v", result["active"])
	}
	if result["score"] != "95.5" {
		t.Errorf("Expected score '95.5', got %v", result["score"])
	}
}

func TestXMLArrayMarshalUnmarshal(t *testing.T) {
	testData := []map[string]any{
		{
			"id":   1,
			"name": "Alice",
		},
		{
			"id":   2,
			"name": "Bob",
		},
	}

	codec := &Codec{}

	// Test marshaling
	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal XML array: %v", err)
	}

	// Check XML structure
	xmlStr := string(data)
	if !strings.Contains(xmlStr, "<root>") {
		t.Error("XML array output missing root element")
	}

	// Test unmarshaling
	var result []map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal XML array: %v", err)
	}

	// Verify length
	if len(result) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(result))
	}

	// Verify first record - XML converts numbers to strings
	if result[0]["id"] != "1" {
		t.Errorf("Expected id '1', got %v", result[0]["id"])
	}
	if result[0]["name"] != "Alice" {
		t.Errorf("Expected name 'Alice', got %v", result[0]["name"])
	}
}

func TestXMLWithSpecialCharacters(t *testing.T) {
	testData := map[string]any{
		"description": "Text with <tags> & \"quotes\"",
		"code":        "if (x > 5) { return true; }",
		"unicode":     "Hello 世界",
	}

	codec := &Codec{}

	// Test marshaling
	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal XML with special characters: %v", err)
	}

	// Test unmarshaling
	var result map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal XML with special characters: %v", err)
	}

	// Verify special characters are properly escaped/unescaped
	if result["description"] != "Text with <tags> & \"quotes\"" {
		t.Errorf("Special characters not preserved: %v", result["description"])
	}
	if result["code"] != "if (x > 5) { return true; }" {
		t.Errorf("Code with brackets not preserved: %v", result["code"])
	}
	if result["unicode"] != "Hello 世界" {
		t.Errorf("Unicode not preserved: %v", result["unicode"])
	}
}

func TestXMLNestedStructure(t *testing.T) {
	testData := map[string]any{
		"user": map[string]any{
			"personal": map[string]any{
				"name": "Alice",
				"age":  30,
			},
			"professional": map[string]any{
				"title":      "Engineer",
				"department": "Technology",
			},
		},
	}

	codec := &Codec{}

	// Test marshaling
	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal nested XML: %v", err)
	}

	// Test unmarshaling
	var result map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal nested XML: %v", err)
	}

	// Navigate nested structure
	user := result["user"].(map[string]any)
	personal := user["personal"].(map[string]any)
	professional := user["professional"].(map[string]any)

	if personal["name"] != "Alice" {
		t.Errorf("Expected nested name 'Alice', got %v", personal["name"])
	}
	if personal["age"] != "30" {
		t.Errorf("Expected nested age '30', got %v", personal["age"])
	}
	if professional["title"] != "Engineer" {
		t.Errorf("Expected title 'Engineer', got %v", professional["title"])
	}
}

func TestXMLEmptyAndNullValues(t *testing.T) {
	testData := map[string]any{
		"name":         "Alice",
		"middle_name":  "",
		"optional":     nil,
		"empty_object": map[string]any{},
	}

	codec := &Codec{}

	// Test marshaling
	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal XML with empty/null values: %v", err)
	}

	// Test unmarshaling
	var result map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal XML with empty/null values: %v", err)
	}

	// Verify handling of empty and null values
	if result["name"] != "Alice" {
		t.Errorf("Expected name 'Alice', got %v", result["name"])
	}
	if result["middle_name"] != "" {
		t.Errorf("Expected empty middle_name, got %v", result["middle_name"])
	}
}

func TestXMLRoundTrip(t *testing.T) {
	originalData := map[string]any{
		"string":  "test value",
		"number":  42,
		"boolean": true,
		"nested": map[string]any{
			"inner": "nested value",
		},
	}

	codec := &Codec{}

	// First marshal
	data1, err := codec.Marshal(originalData)
	if err != nil {
		t.Fatalf("Failed first marshal: %v", err)
	}

	// First unmarshal
	var result1 map[string]any
	err = codec.Unmarshal(data1, &result1)
	if err != nil {
		t.Fatalf("Failed first unmarshal: %v", err)
	}

	// Second marshal
	data2, err := codec.Marshal(result1)
	if err != nil {
		t.Fatalf("Failed second marshal: %v", err)
	}

	// Second unmarshal
	var result2 map[string]any
	err = codec.Unmarshal(data2, &result2)
	if err != nil {
		t.Fatalf("Failed second unmarshal: %v", err)
	}

	// Compare values (noting XML converts all to strings)
	if result2["string"] != "test value" {
		t.Errorf("String field changed: %v", result2["string"])
	}
	if result2["number"] != "42" {
		t.Errorf("Number field changed: %v", result2["number"])
	}
	if result2["boolean"] != "true" {
		t.Errorf("Boolean field changed: %v", result2["boolean"])
	}

	// Check nested structure
	nested2 := result2["nested"].(map[string]any)
	if nested2["inner"] != "nested value" {
		t.Errorf("Nested field changed: %v", nested2["inner"])
	}
}

func TestXMLValidFormat(t *testing.T) {
	testData := map[string]any{
		"item1": "value1",
		"item2": "value2",
	}

	codec := &Codec{}

	// Test marshaling
	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal XML: %v", err)
	}

	xmlStr := string(data)

	// Basic XML validation checks
	if !strings.HasPrefix(strings.TrimSpace(xmlStr), "<?xml") {
		t.Error("XML missing XML declaration")
	}

	if !strings.Contains(xmlStr, "<root>") {
		t.Error("XML missing root opening tag")
	}

	if !strings.Contains(xmlStr, "</root>") {
		t.Error("XML missing root closing tag")
	}

	// Count opening and closing tags for item1
	openCount := strings.Count(xmlStr, "<item1>")
	closeCount := strings.Count(xmlStr, "</item1>")
	if openCount != closeCount {
		t.Errorf("Mismatched item1 tags: %d open, %d close", openCount, closeCount)
	}
}
