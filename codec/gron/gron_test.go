package gron

import (
	"testing"
)

func TestBasicGronMarshalUnmarshal(t *testing.T) {
	testData := map[string]any{
		"name":  "test",
		"value": 42,
	}

	codec := &Codec{}

	// Test marshaling
	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal gron data: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Marshaled data is empty")
	}

	// Test unmarshaling
	var result map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal gron data: %v", err)
	}

	if result["name"] != "test" {
		t.Errorf("Expected name 'test', got %v", result["name"])
	}
}
