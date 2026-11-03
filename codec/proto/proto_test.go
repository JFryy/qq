package codec

import (
	"reflect"
	"testing"
)

func TestCodecUnmarshal(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string]any
	}{
		{
			name:  "message fields and comments",
			input: "syntax = \"proto3\";\npackage people; // comment\n/* ignored */ message Person { string name = 1; int32 age = 2; bool active = 3; }",
			want: map[string]any{"people": map[string]any{
				"message": map[string]any{"Person": []any{
					map[string]any{"name": "name", "type": "string", "number": float64(1)},
					map[string]any{"name": "age", "type": "int32", "number": float64(2)},
					map[string]any{"name": "active", "type": "bool", "number": float64(3)},
				}},
				"enum": map[string]any{},
			}},
		},
		{
			name:  "enum with negative value",
			input: "package state; enum Status { UNKNOWN = 0; ACTIVE = 1; INVALID = -1; }",
			want: map[string]any{"state": map[string]any{
				"message": map[string]any{},
				"enum": map[string]any{"Status": map[string]any{
					"UNKNOWN": float64(0), "ACTIVE": float64(1), "INVALID": float64(-1),
				}},
			}},
		},
		{
			name:  "message reference",
			input: "package team; message Address { string city = 1; } message Person { Address address = 1; }",
			want: map[string]any{"team": map[string]any{
				"message": map[string]any{
					"Address": []any{map[string]any{"name": "city", "type": "string", "number": float64(1)}},
					"Person":  []any{map[string]any{"name": "address", "type": "Address", "number": float64(1)}},
				},
				"enum": map[string]any{},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got map[string]any
			if err := (&Codec{}).Unmarshal([]byte(tt.input), &got); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			// Field order is unspecified because fields are stored in a map.
			packageName := ""
			for name := range tt.want {
				packageName = name
			}
			actual := got[packageName].(map[string]any)
			expected := tt.want[packageName].(map[string]any)
			actualMessages := actual["message"].(map[string]any)
			expectedMessages := expected["message"].(map[string]any)
			for name, fields := range expectedMessages {
				gotFields, ok := actualMessages[name].([]any)
				if !ok {
					t.Fatalf("message %s = %T, want fields", name, actualMessages[name])
				}
				if len(gotFields) != len(fields.([]any)) {
					t.Fatalf("message %s has %d fields, want %d", name, len(gotFields), len(fields.([]any)))
				}
				for _, field := range fields.([]any) {
					found := false
					for _, candidate := range gotFields {
						if reflect.DeepEqual(candidate, field) {
							found = true
						}
					}
					if !found {
						t.Errorf("message %s missing field %v in %v", name, field, gotFields)
					}
				}
			}
			if len(actualMessages) != len(expectedMessages) || !reflect.DeepEqual(actual["enum"], expected["enum"]) || len(got) != 1 {
				t.Errorf("Unmarshal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCodecUnmarshalInvalidDestination(t *testing.T) {
	if err := (&Codec{}).Unmarshal([]byte("package demo;"), nil); err == nil {
		t.Error("Unmarshal() with nil destination returned no error")
	}
}

func TestRemoveComments(t *testing.T) {
	input := " // heading\n package demo; /* block */\n message A { string name = 1; // field\n } "
	want := "package demo; \n message A { string name = 1; \n }"
	if got := removeComments(input); got != want {
		t.Errorf("removeComments() = %q, want %q", got, want)
	}
}
