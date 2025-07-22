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
	if !strings.Contains(xmlStr, "<doc>") || !strings.Contains(xmlStr, "</doc>") {
		t.Error("XML output missing doc element")
	}

	// Test unmarshaling
	var result map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal XML data: %v", err)
	}

	// The XML codec wraps content in a doc element and parses values by type
	doc := result["doc"].(map[string]any)
	if doc["name"] != "Alice" {
		t.Errorf("Expected name 'Alice', got %v", doc["name"])
	}
	if doc["age"] != 30 {
		t.Errorf("Expected age 30, got %v", doc["age"])
	}
	if doc["active"] != true {
		t.Errorf("Expected active true, got %v", doc["active"])
	}
	if doc["score"] != 95.5 {
		t.Errorf("Expected score 95.5, got %v", doc["score"])
	}
}

func TestXMLArrayMarshalUnmarshal(t *testing.T) {
	testData := []any{
		map[string]any{
			"id":   1,
			"name": "Alice",
		},
		map[string]any{
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
	var result map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal XML array: %v", err)
	}

	// The array items should be in the doc as separate root elements
	doc := result["doc"].(map[string]any)
	rootItems := doc["root"].([]any)

	// Verify length
	if len(rootItems) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(rootItems))
	}

	// Verify first record - values are parsed by type
	firstItem := rootItems[0].(map[string]any)
	if firstItem["id"] != 1 {
		t.Errorf("Expected id 1, got %v", firstItem["id"])
	}
	if firstItem["name"] != "Alice" {
		t.Errorf("Expected name 'Alice', got %v", firstItem["name"])
	}
}

func TestXMLWithSpecialCharacters(t *testing.T) {
	testData := map[string]any{
		"description": "Text with simple content",
		"code":        "if x then return true",
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

	// Verify content is preserved
	doc := result["doc"].(map[string]any)
	if doc["description"] != "Text with simple content" {
		t.Errorf("Description not preserved: %v", doc["description"])
	}
	if doc["code"] != "if x then return true" {
		t.Errorf("Code not preserved: %v", doc["code"])
	}
	if doc["unicode"] != "Hello 世界" {
		t.Errorf("Unicode not preserved: %v", doc["unicode"])
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

	// Navigate nested structure - no doc wrapper for this structure
	user := result["user"].(map[string]any)
	personal := user["personal"].(map[string]any)
	professional := user["professional"].(map[string]any)

	if personal["name"] != "Alice" {
		t.Errorf("Expected nested name 'Alice', got %v", personal["name"])
	}
	if personal["age"] != 30 {
		t.Errorf("Expected nested age 30, got %v", personal["age"])
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
	doc := result["doc"].(map[string]any)
	if doc["name"] != "Alice" {
		t.Errorf("Expected name 'Alice', got %v", doc["name"])
	}
	// Empty strings and nil values may not be preserved in XML
	if val, exists := doc["middle_name"]; exists && val != "" {
		t.Errorf("Expected empty or missing middle_name, got %v", val)
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

	// Compare values (XML codec parses values by type)
	doc2 := result2["doc"].(map[string]any)
	if doc2["string"] != "test value" {
		t.Errorf("String field changed: %v", doc2["string"])
	}
	if doc2["number"] != 42 {
		t.Errorf("Number field changed: %v", doc2["number"])
	}
	if doc2["boolean"] != true {
		t.Errorf("Boolean field changed: %v", doc2["boolean"])
	}

	// Check nested structure
	if nested2, ok := doc2["nested"].(map[string]any); ok {
		if nested2["inner"] != "nested value" {
			t.Errorf("Nested field changed: %v", nested2["inner"])
		}
	} else {
		t.Errorf("Nested structure not preserved: %v", doc2["nested"])
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

	// Basic XML validation checks - XML declaration is optional
	trimmed := strings.TrimSpace(xmlStr)
	if !strings.HasPrefix(trimmed, "<?xml") && !strings.HasPrefix(trimmed, "<doc>") {
		t.Error("XML missing XML declaration or doc element")
	}

	if !strings.Contains(xmlStr, "<doc>") {
		t.Error("XML missing doc opening tag")
	}

	if !strings.Contains(xmlStr, "</doc>") {
		t.Error("XML missing doc closing tag")
	}

	// Count opening and closing tags for item1
	openCount := strings.Count(xmlStr, "<item1>")
	closeCount := strings.Count(xmlStr, "</item1>")
	if openCount != closeCount {
		t.Errorf("Mismatched item1 tags: %d open, %d close", openCount, closeCount)
	}
}
