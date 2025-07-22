package parquet

import (
	"testing"
)

func TestBasicParquetMarshalUnmarshal(t *testing.T) {
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
		t.Fatalf("Failed to marshal parquet data: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Marshaled data is empty")
	}

	// Test unmarshaling
	var result []map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal parquet data: %v", err)
	}

	// Verify length
	if len(result) != 3 {
		t.Fatalf("Expected 3 records, got %d", len(result))
	}

	// Verify first record structure
	first := result[0]
	if first["ID"] != "1" {
		t.Errorf("Expected ID '1', got %v", first["ID"])
	}
	if first["Name"] != "Alice" {
		t.Errorf("Expected Name 'Alice', got %v", first["Name"])
	}
	if first["Age"] != "30" {
		t.Errorf("Expected Age '30', got %v", first["Age"])
	}
	if first["Active"] != "true" {
		t.Errorf("Expected Active 'true', got %v", first["Active"])
	}
	if first["Score"] != "95.5" {
		t.Errorf("Expected Score '95.5', got %v", first["Score"])
	}
	if first["Department"] != "Engineering" {
		t.Errorf("Expected Department 'Engineering', got %v", first["Department"])
	}
}

func TestEmptyDataHandling(t *testing.T) {
	codec := &Codec{}

	// Test empty slice
	emptyData := []map[string]any{}
	_, err := codec.Marshal(emptyData)
	if err == nil {
		t.Error("Expected error for empty data, got nil")
	}
}

func TestNilValueHandling(t *testing.T) {
	testData := []map[string]any{
		{
			"ID":            1,
			"Name":          "Alice",
			"OptionalField": nil,
			"Department":    "Engineering",
		},
		{
			"ID":            2,
			"Name":          "Bob",
			"OptionalField": "Some Value",
			"Department":    "Sales",
		},
	}

	codec := &Codec{}

	// Test marshaling with nil values
	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal parquet data with nil values: %v", err)
	}

	// Test unmarshaling
	var result []map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal parquet data with nil values: %v", err)
	}

	// Verify length
	if len(result) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(result))
	}

	// The nil value should be handled (converted to null string or similar)
	first := result[0]
	if first["ID"] != "1" {
		t.Errorf("Expected ID '1', got %v", first["ID"])
	}

	// Second record should have the value
	second := result[1]
	if second["OptionalField"] != "Some Value" {
		t.Errorf("Expected OptionalField 'Some Value', got %v", second["OptionalField"])
	}
}

func TestMissingFieldsHandling(t *testing.T) {
	testData := []map[string]any{
		{
			"ID":   1,
			"Name": "Alice",
			"Age":  30,
		},
		{
			"ID":         2,
			"Name":       "Bob",
			"Department": "Sales", // Missing Age field
		},
		{
			"ID":         3,
			"Name":       "Charlie",
			"Age":        35,
			"Department": "Engineering", // Has all fields
		},
	}

	codec := &Codec{}

	// Test marshaling with inconsistent fields
	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal parquet data with missing fields: %v", err)
	}

	// Test unmarshaling
	var result []map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal parquet data with missing fields: %v", err)
	}

	// Verify length
	if len(result) != 3 {
		t.Fatalf("Expected 3 records, got %d", len(result))
	}

	// All records should have the same field structure (with null values for missing fields)
	for i, record := range result {
		if record["ID"] == nil {
			t.Errorf("Record %d missing ID field", i)
		}
		if record["Name"] == nil {
			t.Errorf("Record %d missing Name field", i)
		}
		// Age and Department might be null for some records, which is fine
	}
}

func TestLargeDataSet(t *testing.T) {
	// Create a larger dataset to test performance and memory handling
	testData := make([]map[string]any, 1000)
	for i := 0; i < 1000; i++ {
		testData[i] = map[string]any{
			"ID":       i + 1,
			"Name":     "User" + string(rune(i%26+65)),
			"Value":    float64(i) * 1.5,
			"Active":   i%2 == 0,
			"Category": "Category" + string(rune(i%5+65)),
		}
	}

	codec := &Codec{}

	// Test marshaling large dataset
	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal large parquet dataset: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Marshaled large dataset is empty")
	}

	// Test unmarshaling large dataset
	var result []map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal large parquet dataset: %v", err)
	}

	// Verify length
	if len(result) != 1000 {
		t.Fatalf("Expected 1000 records, got %d", len(result))
	}

	// Spot check a few records
	if result[0]["ID"] != "1" {
		t.Errorf("First record ID incorrect: %v", result[0]["ID"])
	}
	if result[999]["ID"] != "1000" {
		t.Errorf("Last record ID incorrect: %v", result[999]["ID"])
	}
}

func TestInvalidInputTypes(t *testing.T) {
	codec := &Codec{}

	// Test non-slice input
	invalidData := map[string]any{"key": "value"}
	_, err := codec.Marshal(invalidData)
	if err == nil {
		t.Error("Expected error for non-slice input, got nil")
	}

	// Test slice of non-map elements
	invalidSlice := []string{"item1", "item2"}
	_, err = codec.Marshal(invalidSlice)
	if err == nil {
		t.Error("Expected error for slice of non-map elements, got nil")
	}
}

func TestRoundTripConsistency(t *testing.T) {
	originalData := []map[string]any{
		{
			"StringField": "test string",
			"NumberField": 42,
			"FloatField":  3.14159,
			"BoolField":   true,
			"NullField":   nil,
		},
		{
			"StringField": "another string",
			"NumberField": -17,
			"FloatField":  -2.718,
			"BoolField":   false,
			"NullField":   nil,
		},
	}

	codec := &Codec{}

	// Marshal original data
	data, err := codec.Marshal(originalData)
	if err != nil {
		t.Fatalf("Failed to marshal data: %v", err)
	}

	// Unmarshal to get result
	var result1 []map[string]any
	err = codec.Unmarshal(data, &result1)
	if err != nil {
		t.Fatalf("Failed to unmarshal data: %v", err)
	}

	// Marshal the result again
	data2, err := codec.Marshal(result1)
	if err != nil {
		t.Fatalf("Failed to marshal result data: %v", err)
	}

	// Unmarshal again
	var result2 []map[string]any
	err = codec.Unmarshal(data2, &result2)
	if err != nil {
		t.Fatalf("Failed to unmarshal data second time: %v", err)
	}

	// Results should be consistent
	if len(result1) != len(result2) {
		t.Fatalf("Round-trip changed length: %d vs %d", len(result1), len(result2))
	}

	for i := range result1 {
		if len(result1[i]) != len(result2[i]) {
			t.Errorf("Record %d field count changed: %d vs %d", i, len(result1[i]), len(result2[i]))
		}

		for key := range result1[i] {
			if result1[i][key] != result2[i][key] {
				t.Errorf("Record %d field %s changed: %v vs %v", i, key, result1[i][key], result2[i][key])
			}
		}
	}
}
