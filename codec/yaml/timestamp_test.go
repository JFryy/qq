package yaml

import (
	"reflect"
	"testing"

	"github.com/itchyny/gojq"
)

func TestYAMLTimestampValues(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  any
	}{
		{"date", "2024-01-01", "2024-01-01"},
		{"UTC", "2024-01-01T10:20:30Z", "2024-01-01T10:20:30Z"},
		{"offset and fraction", "2024-01-01T10:20:30.125+02:00", "2024-01-01T10:20:30.125+02:00"},
		{"nested mapping", "a:\n  b: 2024-01-01\n", map[string]any{"a": map[string]any{"b": "2024-01-01"}}},
		{"sequence", "- 2024-01-01\n- 2024-01-01T10:20:30Z\n", []any{"2024-01-01", "2024-01-01T10:20:30Z"}},
		{"non-string key", "200: 2024-01-01\n", map[string]any{"200": "2024-01-01"}},
		{"multi-document", "2024-01-01\n---\n2024-01-01T10:20:30Z\n", []any{"2024-01-01", "2024-01-01T10:20:30Z"}},
		{"quoted timestamp", "'2024-01-01 10:20:30Z'", "2024-01-01 10:20:30Z"},
		{"null", "null", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got any
			if err := (Codec{}).Unmarshal([]byte(tt.input), &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %#v (%T), want %#v (%T)", got, got, tt.want, tt.want)
			}
			// gojq's encoder rejects time.Time, unlike encoding/json.
			if _, err := gojq.Marshal(got); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestYAMLTimestampKeyValueParity(t *testing.T) {
	got := unmarshalTestMap(t, "test:\n  2024-01-01: 2024-01-01\n  2024-01-01T10:20:30.5+02:00: 2024-01-01T10:20:30.5+02:00\n")
	for key, value := range got {
		if value != key {
			t.Errorf("timestamp key %q and value %#v disagree", key, value)
		}
	}
}
