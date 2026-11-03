package avro

import (
	"reflect"
	"strings"
	"testing"
)

func TestCodecUnmarshal(t *testing.T) {
	codec := &Codec{}
	tests := []struct {
		name  string
		input any
		want  []map[string]any
	}{
		{
			name: "single record with scalar types",
			input: map[string]any{
				"name": "alice", "count": 42, "ratio": 1.5,
				"active": true, "missing": nil,
			},
			want: []map[string]any{{
				"name": "alice", "count": float64(42), "ratio": 1.5,
				"active": true, "missing": nil,
			}},
		},
		{
			name: "multiple records",
			input: []map[string]any{
				{"name": "first", "active": true},
				{"name": "second", "active": false},
			},
			want: []map[string]any{
				{"name": "first", "active": true},
				{"name": "second", "active": false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := codec.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			var got []map[string]any
			if err := codec.Unmarshal(encoded, &got); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unmarshal() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestCodecMarshalErrors(t *testing.T) {
	codec := &Codec{}
	tests := []struct {
		name  string
		input any
	}{
		{"empty records", []map[string]any{}},
		{"scalar", 42},
		{"array of scalars", []any{1, 2}},
		{"unsupported value", map[string]any{"bad": make(chan int)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := codec.Marshal(tt.input); err == nil {
				t.Error("Marshal() error = nil, want error")
			}
		})
	}
}

func TestCodecUnmarshalErrors(t *testing.T) {
	codec := &Codec{}
	tests := []struct {
		name string
		data []byte
	}{
		{"empty input", nil},
		{"not avro", []byte("not an Avro container")},
		{"truncated header", []byte("Obj\x01")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got any
			if err := codec.Unmarshal(tt.data, &got); err == nil {
				t.Error("Unmarshal() error = nil, want error")
			}
		})
	}
}

func TestCodecUnmarshalDestinationError(t *testing.T) {
	codec := &Codec{}
	encoded, err := codec.Marshal(map[string]any{"name": "alice"})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	var destination map[string]any
	if err := codec.Unmarshal(encoded, &destination); err == nil {
		t.Error("Unmarshal() into a map = nil error, want error")
	}
}

func TestInferSchema(t *testing.T) {
	schema, err := inferSchema(map[string]any{
		"z": nil, "b": true, "a": float64(1), "m": "text",
	})
	if err != nil {
		t.Fatalf("inferSchema() error = %v", err)
	}
	want := `{"type":"record","name":"Root","fields":[{"name":"a","type":["null","double"],"default":null},{"name":"b","type":["null","boolean"],"default":null},{"name":"m","type":["null","string"],"default":null},{"name":"z","type":["null","string"],"default":null}]}`
	if schema != want {
		t.Errorf("inferSchema() = %s, want %s", schema, want)
	}
}

func TestStringifyComplex(t *testing.T) {
	input := map[string]any{
		"nested": map[string]any{"key": "value"},
		"array":  []any{true, 2},
		"null":   nil,
	}
	got := stringifyComplex(input)
	want := map[string]any{
		"nested": `{"key":"value"}`,
		"array":  `[true,2]`,
		"null":   nil,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("stringifyComplex() = %#v, want %#v", got, want)
	}
	if _, ok := input["nested"].(map[string]any); !ok {
		t.Error("stringifyComplex() changed the input")
	}
	if !strings.Contains(got["nested"].(string), "value") {
		t.Error("stringifyComplex() lost nested data")
	}
}
