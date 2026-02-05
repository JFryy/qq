package jsonl

import (
	"testing"
)

func TestBasicJSONLMarshalUnmarshal(t *testing.T) {
	testData := []map[string]any{
		{
			"name": "Alice",
			"age":  30,
			"city": "NYC",
		},
		{
			"name": "Bob",
			"age":  25,
			"city": "London",
		},
		{
			"name": "Charlie",
			"age":  35,
			"city": "Paris",
		},
	}

	codec := &Codec{}

	// Test marshaling
	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal JSONL data: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Marshaled data is empty")
	}

	// Test unmarshaling
	var result []map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSONL data: %v", err)
	}

	// Verify length
	if len(result) != 3 {
		t.Fatalf("Expected 3 records, got %d", len(result))
	}

	// Verify first record
	first := result[0]
	if first["name"] != "Alice" {
		t.Errorf("Expected name 'Alice', got %v", first["name"])
	}
	if first["age"].(float64) != 30 {
		t.Errorf("Expected age 30, got %v", first["age"])
	}
}

func TestJSONLWithEmptyLines(t *testing.T) {
	jsonlData := `{"name": "Alice", "age": 30}

{"name": "Bob", "age": 25}

`

	codec := &Codec{}
	var result []map[string]any
	err := codec.Unmarshal([]byte(jsonlData), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSONL with empty lines: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(result))
	}
}

func TestJSONLWithDifferentTypes(t *testing.T) {
	jsonlData := `{"type": "user", "name": "Alice"}
{"type": "product", "id": 123, "price": 99.99}
{"type": "order", "items": [1, 2, 3]}`

	codec := &Codec{}
	var result []any
	err := codec.Unmarshal([]byte(jsonlData), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSONL with different types: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("Expected 3 records, got %d", len(result))
	}
}

func TestJSONLRoundTrip(t *testing.T) {
	originalData := []map[string]any{
		{
			"id":     1,
			"status": "active",
			"tags":   []any{"important", "urgent"},
		},
		{
			"id":     2,
			"status": "pending",
			"tags":   []any{"review"},
		},
	}

	codec := &Codec{}

	// Marshal original data
	data, err := codec.Marshal(originalData)
	if err != nil {
		t.Fatalf("Failed to marshal data: %v", err)
	}

	// Unmarshal to get result
	var result []map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal data: %v", err)
	}

	// Verify structure is preserved
	if len(result) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(result))
	}

	// Check first record
	first := result[0]
	if first["id"].(float64) != 1 {
		t.Errorf("ID mismatch: %v", first["id"])
	}
	if first["status"] != "active" {
		t.Errorf("Status mismatch: %v", first["status"])
	}
}

func TestJSONLInvalidJSON(t *testing.T) {
	jsonlData := `{"valid": "json"}
{invalid json}
{"another": "valid"}`

	codec := &Codec{}
	var result []any
	err := codec.Unmarshal([]byte(jsonlData), &result)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestJSONLSingleObject(t *testing.T) {
	singleObj := map[string]any{
		"name": "Alice",
		"age":  30,
	}

	codec := &Codec{}

	// Marshal single object
	data, err := codec.Marshal(singleObj)
	if err != nil {
		t.Fatalf("Failed to marshal single object: %v", err)
	}

	// Should produce single line
	var result []map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(result))
	}
}
