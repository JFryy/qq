package yaml

import (
	"testing"
)

// Helper function to check numeric values that might be int, int64, uint64, or float64
func assertNumericEqual(t *testing.T, actual any, expected float64, fieldName string) {
	switch v := actual.(type) {
	case int:
		if float64(v) != expected {
			t.Errorf("Expected %s %.0f, got %v", fieldName, expected, v)
		}
	case int64:
		if float64(v) != expected {
			t.Errorf("Expected %s %.0f, got %v", fieldName, expected, v)
		}
	case uint64:
		if float64(v) != expected {
			t.Errorf("Expected %s %.0f, got %v", fieldName, expected, v)
		}
	case float64:
		if v != expected {
			t.Errorf("Expected %s %v, got %v", fieldName, expected, v)
		}
	default:
		t.Errorf("Expected %s as number, got %v (type: %T)", fieldName, v, v)
	}
}

func TestBasicYAMLMarshalUnmarshal(t *testing.T) {
	testData := map[string]any{
		"name":   "Alice",
		"age":    30,
		"active": true,
		"score":  95.5,
		"tags":   []string{"engineer", "golang"},
		"metadata": map[string]any{
			"department": "Engineering",
			"level":      "Senior",
		},
	}

	codec := &Codec{}

	// Test marshaling
	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal YAML data: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Marshaled data is empty")
	}

	// Test unmarshaling
	var result map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML data: %v", err)
	}

	// Verify basic fields
	// Note: YAML may preserve int types, unlike pure JSON round-trip
	if result["name"] != "Alice" {
		t.Errorf("Expected name 'Alice', got %v", result["name"])
	}
	assertNumericEqual(t, result["age"], 30, "age")
	if result["active"] != true {
		t.Errorf("Expected active true, got %v", result["active"])
	}
	if result["score"] != 95.5 {
		t.Errorf("Expected score 95.5, got %v", result["score"])
	}

	// Verify array
	tags := result["tags"].([]any)
	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tags))
	}
	if tags[0] != "engineer" {
		t.Errorf("Expected first tag 'engineer', got %v", tags[0])
	}

	// Verify nested object
	metadata := result["metadata"].(map[string]any)
	if metadata["department"] != "Engineering" {
		t.Errorf("Expected department 'Engineering', got %v", metadata["department"])
	}
}

func TestYAMLArrayMarshalUnmarshal(t *testing.T) {
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
		t.Fatalf("Failed to marshal YAML array: %v", err)
	}

	// Test unmarshaling
	var result []map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML array: %v", err)
	}

	// Verify length
	if len(result) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(result))
	}

	// Verify first record
	assertNumericEqual(t, result[0]["id"], 1, "id")
	if result[0]["name"] != "Alice" {
		t.Errorf("Expected name 'Alice', got %v", result[0]["name"])
	}
}

func TestYAMLWithNullValues(t *testing.T) {
	testData := map[string]any{
		"name":        "Alice",
		"middle_name": nil,
		"age":         30,
		"optional":    nil,
	}

	codec := &Codec{}

	// Test marshaling
	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal YAML with null values: %v", err)
	}

	// Test unmarshaling
	var result map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML with null values: %v", err)
	}

	// Verify null values are preserved
	if result["middle_name"] != nil {
		t.Errorf("Expected middle_name nil, got %v", result["middle_name"])
	}
	if result["optional"] != nil {
		t.Errorf("Expected optional nil, got %v", result["optional"])
	}
	if result["name"] != "Alice" {
		t.Errorf("Expected name 'Alice', got %v", result["name"])
	}
}

func TestYAMLMultilineStrings(t *testing.T) {
	testData := map[string]any{
		"description": "This is a\nmultiline\nstring",
		"code":        "func main() {\n\tfmt.Println(\"Hello\")\n}",
		"simple":      "single line",
	}

	codec := &Codec{}

	// Test marshaling
	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal YAML with multiline strings: %v", err)
	}

	// Test unmarshaling
	var result map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML with multiline strings: %v", err)
	}

	// Verify multiline strings are preserved
	if result["description"] != "This is a\nmultiline\nstring" {
		t.Errorf("Multiline string not preserved: %v", result["description"])
	}
	if result["simple"] != "single line" {
		t.Errorf("Simple string not preserved: %v", result["simple"])
	}
}

func TestYAMLBooleans(t *testing.T) {
	testData := map[string]any{
		"true_bool":  true,
		"false_bool": false,
		"true_str":   "true",
		"false_str":  "false",
		"yes_str":    "yes",
		"no_str":     "no",
	}

	codec := &Codec{}

	// Test marshaling
	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal YAML with booleans: %v", err)
	}

	// Test unmarshaling
	var result map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML with booleans: %v", err)
	}

	// Verify boolean values
	if result["true_bool"] != true {
		t.Errorf("Expected true_bool true, got %v", result["true_bool"])
	}
	if result["false_bool"] != false {
		t.Errorf("Expected false_bool false, got %v", result["false_bool"])
	}

	// String representations should remain as strings
	if result["true_str"] != "true" {
		t.Errorf("Expected true_str 'true', got %v", result["true_str"])
	}
}

func TestYAMLNumbers(t *testing.T) {
	testData := map[string]any{
		"int":            42,
		"negative_int":   -17,
		"float":          3.14159,
		"negative_float": -2.718,
		"zero":           0,
		"scientific":     1.23e10,
	}

	codec := &Codec{}

	// Test marshaling
	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal YAML with numbers: %v", err)
	}

	// Test unmarshaling
	var result map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML with numbers: %v", err)
	}

	// Verify numbers (YAML may preserve int types)
	assertNumericEqual(t, result["int"], 42, "int")
	assertNumericEqual(t, result["negative_int"], -17, "negative_int")
	assertNumericEqual(t, result["float"], 3.14159, "float")
	assertNumericEqual(t, result["zero"], 0, "zero")
}

func TestYAMLDeepNesting(t *testing.T) {
	testData := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": map[string]any{
					"level4": "deep value",
					"array": []any{
						map[string]any{"item": 1},
						map[string]any{"item": 2},
					},
				},
			},
		},
	}

	codec := &Codec{}

	// Test marshaling
	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal deeply nested YAML: %v", err)
	}

	// Test unmarshaling
	var result map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal deeply nested YAML: %v", err)
	}

	// Navigate to deep value
	level1 := result["level1"].(map[string]any)
	level2 := level1["level2"].(map[string]any)
	level3 := level2["level3"].(map[string]any)
	deepValue := level3["level4"]

	if deepValue != "deep value" {
		t.Errorf("Expected 'deep value', got %v", deepValue)
	}

	// Check nested array
	array := level3["array"].([]any)
	if len(array) != 2 {
		t.Errorf("Expected 2 array items, got %d", len(array))
	}

	firstItem := array[0].(map[string]any)
	assertNumericEqual(t, firstItem["item"], 1, "item")
}

func TestYAMLRoundTrip(t *testing.T) {
	originalData := map[string]any{
		"string":  "test",
		"number":  42,
		"float":   3.14,
		"boolean": true,
		"null":    nil,
		"array":   []any{1, 2, 3},
		"object": map[string]any{
			"nested": "value",
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

	// Compare key values
	if result2["string"] != "test" {
		t.Errorf("String field changed: %v", result2["string"])
	}
	assertNumericEqual(t, result2["number"], 42, "number")
	if result2["boolean"] != true {
		t.Errorf("Boolean field changed: %v", result2["boolean"])
	}
	if result2["null"] != nil {
		t.Errorf("Null field changed: %v", result2["null"])
	}
}

func TestMultiDocumentYAML(t *testing.T) {
	multiDocYAML := `---
name: document1
value: 100
---
name: document2
value: 200
---
name: document3
value: 300`

	codec := &Codec{}

	// Test unmarshaling multi-document YAML
	var result any
	err := codec.Unmarshal([]byte(multiDocYAML), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal multi-document YAML: %v", err)
	}

	// Should return an array of documents
	docs, ok := result.([]any)
	if !ok {
		t.Fatalf("Expected []any, got %T", result)
	}

	if len(docs) != 3 {
		t.Fatalf("Expected 3 documents, got %d", len(docs))
	}

	// Verify first document
	doc1 := docs[0].(map[string]any)
	if doc1["name"] != "document1" {
		t.Errorf("Expected name 'document1', got %v", doc1["name"])
	}
	assertNumericEqual(t, doc1["value"], 100, "value")

	// Verify second document
	doc2 := docs[1].(map[string]any)
	if doc2["name"] != "document2" {
		t.Errorf("Expected name 'document2', got %v", doc2["name"])
	}
	assertNumericEqual(t, doc2["value"], 200, "value")
}

func TestSingleDocumentYAMLNotWrappedInArray(t *testing.T) {
	singleDocYAML := `---
name: single
value: 123`

	codec := &Codec{}

	// Test unmarshaling single-document YAML
	var result any
	err := codec.Unmarshal([]byte(singleDocYAML), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal single-document YAML: %v", err)
	}

	// Should return a map, NOT an array
	doc, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("Expected map[string]any, got %T", result)
	}

	if doc["name"] != "single" {
		t.Errorf("Expected name 'single', got %v", doc["name"])
	}
	assertNumericEqual(t, doc["value"], 123, "value")
}

func TestMultiDocumentYAMLTypeNormalization(t *testing.T) {
	// Test that uint64 types from YAML are normalized to int
	multiDocYAML := `---
id: 1
count: 100
---
id: 2
count: 200`

	codec := &Codec{}

	var result any
	err := codec.Unmarshal([]byte(multiDocYAML), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	docs := result.([]any)
	doc1 := docs[0].(map[string]any)

	// Verify types are normalized (should be int, not uint64)
	switch doc1["id"].(type) {
	case int, int64, float64:
		// Good - normalized type
	case uint, uint64:
		t.Errorf("Expected normalized int type, got uint type: %T", doc1["id"])
	default:
		t.Errorf("Unexpected type for id: %T", doc1["id"])
	}
}

func TestYAMLWithoutDelimiter(t *testing.T) {
	// YAML without --- delimiter should work
	yamlData := `name: test
value: 42
active: true`

	codec := &Codec{}

	var result any
	err := codec.Unmarshal([]byte(yamlData), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML without delimiter: %v", err)
	}

	doc := result.(map[string]any)
	if doc["name"] != "test" {
		t.Errorf("Expected name 'test', got %v", doc["name"])
	}
	assertNumericEqual(t, doc["value"], 42, "value")
	if doc["active"] != true {
		t.Errorf("Expected active true, got %v", doc["active"])
	}
}

func TestEmptyDocuments(t *testing.T) {
	// Test handling of empty documents
	multiDocYAML := `---
name: doc1
---
---
name: doc2`

	codec := &Codec{}

	var result any
	err := codec.Unmarshal([]byte(multiDocYAML), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Result should be an array if multiple documents
	if docs, ok := result.([]any); ok {
		// Should handle empty document (might be nil or empty map)
		if len(docs) < 2 {
			t.Errorf("Expected at least 2 documents, got %d", len(docs))
		}
	} else {
		// Or might be a single map if empty docs are skipped
		if _, ok := result.(map[string]any); !ok {
			t.Errorf("Expected array or map, got %T", result)
		}
	}
}
