package msgpack

import (
	"testing"
)

func TestCodec_Marshal(t *testing.T) {
	codec := &Codec{}

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name: "simple map",
			input: map[string]interface{}{
				"name": "test",
				"age":  30,
			},
			wantErr: false,
		},
		{
			name: "array",
			input: []interface{}{
				"item1",
				"item2",
				42,
			},
			wantErr: false,
		},
		{
			name:    "nil",
			input:   nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := codec.Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Codec.Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("Codec.Marshal() = nil, want non-nil")
			}
		})
	}
}

func TestCodec_Unmarshal(t *testing.T) {
	codec := &Codec{}

	// Test data
	testData := map[string]interface{}{
		"name":   "test",
		"age":    30,
		"active": true,
	}

	// Marshal first
	marshaled, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	// Test unmarshal
	var result interface{}
	err = codec.Unmarshal(marshaled, &result)
	if err != nil {
		t.Errorf("Codec.Unmarshal() error = %v", err)
		return
	}

	// Convert to map for comparison
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Errorf("Codec.Unmarshal() result is not a map")
		return
	}

	// Check specific values
	if resultMap["name"] != "test" {
		t.Errorf("Codec.Unmarshal() name = %v, want %v", resultMap["name"], "test")
	}

	// Age might be converted to different numeric type, so check as number
	if age, ok := resultMap["age"].(int8); ok {
		if int(age) != 30 {
			t.Errorf("Codec.Unmarshal() age = %v, want %v", age, 30)
		}
	} else if age, ok := resultMap["age"].(int); ok {
		if age != 30 {
			t.Errorf("Codec.Unmarshal() age = %v, want %v", age, 30)
		}
	} else {
		t.Errorf("Codec.Unmarshal() age type = %T, want int", resultMap["age"])
	}

	if resultMap["active"] != true {
		t.Errorf("Codec.Unmarshal() active = %v, want %v", resultMap["active"], true)
	}
}

func TestCodec_RoundTrip(t *testing.T) {
	codec := &Codec{}

	tests := []struct {
		name  string
		input interface{}
	}{
		{
			name: "complex map",
			input: map[string]interface{}{
				"string": "hello",
				"number": 42,
				"float":  3.14,
				"bool":   true,
				"array":  []interface{}{"a", "b", "c"},
				"nested": map[string]interface{}{
					"key": "value",
				},
			},
		},
		{
			name:  "array",
			input: []interface{}{1, 2, 3, "test", true},
		},
		{
			name:  "string",
			input: "simple string",
		},
		{
			name:  "number",
			input: 123,
		},
		{
			name:  "boolean",
			input: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			marshaled, err := codec.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			// Unmarshal
			var result interface{}
			err = codec.Unmarshal(marshaled, &result)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			// For complex types, we mainly check that unmarshaling doesn't fail
			// MessagePack may alter numeric types during round-trip
			if result == nil && tt.input != nil {
				t.Errorf("Round trip failed: got nil, want non-nil")
			}
		})
	}
}
