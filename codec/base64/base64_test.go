package base64

import (
	"testing"
)

func TestBase64MarshalUnmarshal(t *testing.T) {
	testData := map[string]any{
		"name":   "Alice",
		"age":    30,
		"active": true,
	}

	codec := &Codec{}

	// Test marshaling
	encoded, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal to base64: %v", err)
	}

	if len(encoded) == 0 {
		t.Fatal("Encoded data is empty")
	}

	// Test unmarshaling
	var result map[string]any
	err = codec.Unmarshal(encoded, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal base64: %v", err)
	}

	// Verify data
	if result["name"] != "Alice" {
		t.Errorf("Expected name 'Alice', got %v", result["name"])
	}
	if result["age"].(float64) != 30 {
		t.Errorf("Expected age 30, got %v", result["age"])
	}
	if result["active"] != true {
		t.Errorf("Expected active true, got %v", result["active"])
	}
}

func TestBase64UnmarshalKnownValue(t *testing.T) {
	// {"message":"hello world"}
	base64Data := "eyJtZXNzYWdlIjoiaGVsbG8gd29ybGQifQ=="

	codec := &Codec{}
	var result map[string]any
	err := codec.Unmarshal([]byte(base64Data), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal known base64: %v", err)
	}

	if result["message"] != "hello world" {
		t.Errorf("Expected message 'hello world', got %v", result["message"])
	}
}

func TestBase64Array(t *testing.T) {
	testData := []any{
		map[string]any{"id": 1, "name": "Alice"},
		map[string]any{"id": 2, "name": "Bob"},
	}

	codec := &Codec{}

	// Marshal
	encoded, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal array: %v", err)
	}

	// Unmarshal
	var result []any
	err = codec.Unmarshal(encoded, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal array: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(result))
	}
}

func TestBase64InvalidData(t *testing.T) {
	codec := &Codec{}
	var result any

	// Invalid base64
	err := codec.Unmarshal([]byte("not-valid-base64!!!"), &result)
	if err == nil {
		t.Error("Expected error for invalid base64")
	}
}

func TestBase64RoundTrip(t *testing.T) {
	originalData := map[string]any{
		"nested": map[string]any{
			"key": "value",
			"num": 42,
		},
		"array":  []any{1, 2, 3},
		"string": "test",
	}

	codec := &Codec{}

	// Marshal
	encoded, err := codec.Marshal(originalData)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal
	var result map[string]any
	err = codec.Unmarshal(encoded, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify nested structure
	nested := result["nested"].(map[string]any)
	if nested["key"] != "value" {
		t.Errorf("Nested key mismatch")
	}
	if nested["num"].(float64) != 42 {
		t.Errorf("Nested num mismatch")
	}
}
