package tsv

import (
	"strings"
	"testing"
)

func TestBasicTSVMarshalUnmarshal(t *testing.T) {
	testData := []map[string]any{
		{
			"ID":         1,
			"Name":       "Alice",
			"Age":        30,
			"Active":     true,
			"Score":      95.5,
			"Department": "Engineering",
		},
		{
			"ID":         2,
			"Name":       "Bob",
			"Age":        25,
			"Active":     false,
			"Score":      87.2,
			"Department": "Sales",
		},
		{
			"ID":         3,
			"Name":       "Charlie",
			"Age":        35,
			"Active":     true,
			"Score":      92.0,
			"Department": "Engineering",
		},
	}

	codec := &Codec{}

	// Test marshaling
	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal TSV data: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Marshaled data is empty")
	}

	// Check that it contains headers and uses tabs
	tsvStr := string(data)
	if !strings.Contains(tsvStr, "ID") || !strings.Contains(tsvStr, "Name") {
		t.Error("TSV output missing expected headers")
	}
	if !strings.Contains(tsvStr, "\t") {
		t.Error("TSV output should contain tab characters")
	}

	// Test unmarshaling
	var result []map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal TSV data: %v", err)
	}

	// Verify length
	if len(result) != 3 {
		t.Fatalf("Expected 3 records, got %d", len(result))
	}

	// Verify first record structure
	first := result[0]
	if first["ID"].(float64) != 1 {
		t.Errorf("Expected ID 1, got %v", first["ID"])
	}
	if first["Name"] != "Alice" {
		t.Errorf("Expected Name 'Alice', got %v", first["Name"])
	}
	if first["Age"].(float64) != 30 {
		t.Errorf("Expected Age 30, got %v", first["Age"])
	}
	if first["Active"] != true {
		t.Errorf("Expected Active true, got %v", first["Active"])
	}
	if first["Score"].(float64) != 95.5 {
		t.Errorf("Expected Score 95.5, got %v", first["Score"])
	}
}

func TestTSVWithTabsInData(t *testing.T) {
	testData := []map[string]any{
		{
			"Name":  "John",
			"Notes": "Has\ttabs\tin\ttext",
		},
		{
			"Name":  "Jane",
			"Notes": "Normal text",
		},
	}

	codec := &Codec{}

	// Test marshaling
	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal TSV data with tabs: %v", err)
	}

	// Test unmarshaling
	var result []map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal TSV data with tabs: %v", err)
	}

	// Verify length
	if len(result) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(result))
	}
}

func TestEmptyTSVData(t *testing.T) {
	codec := &Codec{}

	// Test empty slice
	emptyData := []map[string]any{}
	_, err := codec.Marshal(emptyData)
	if err == nil {
		t.Error("Expected error for empty data, got nil")
	}
}

func TestTSVWithMissingFields(t *testing.T) {
	tsvData := "Name\tAge\tCity\nAlice\t30\tNew York\nBob\t\tLondon\nCharlie\t35\t"

	codec := &Codec{}
	var result []map[string]any
	err := codec.Unmarshal([]byte(tsvData), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal TSV with missing fields: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("Expected 3 records, got %d", len(result))
	}

	// Check Bob's missing age
	if result[1]["Age"] != "" {
		t.Errorf("Expected empty Age for Bob, got %v", result[1]["Age"])
	}

	// Check Charlie's missing city
	if result[2]["City"] != "" {
		t.Errorf("Expected empty City for Charlie, got %v", result[2]["City"])
	}
}

func TestTSVRoundTrip(t *testing.T) {
	originalData := []map[string]any{
		{
			"StringField": "test string",
			"NumberField": 42,
			"FloatField":  3.14159,
			"BoolField":   true,
		},
		{
			"StringField": "another string",
			"NumberField": -17,
			"FloatField":  -2.718,
			"BoolField":   false,
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
	if first["StringField"] != "test string" {
		t.Errorf("String field mismatch: %v", first["StringField"])
	}
	if first["NumberField"].(float64) != 42 {
		t.Errorf("Number field mismatch: %v", first["NumberField"])
	}
	if first["BoolField"] != true {
		t.Errorf("Bool field mismatch: %v", first["BoolField"])
	}
}
