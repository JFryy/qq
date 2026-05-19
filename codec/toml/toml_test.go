package toml

import (
	"math"
	"strings"
	"testing"
)

func TestMarshalWholeFloatsBecomeIntegers(t *testing.T) {
	// JSON unmarshaling into `any` produces float64 for all numbers, so the
	// TOML marshaler must treat whole-valued float64s as integers to avoid
	// emitting `120.0` for what was originally `120`.
	c := Codec{}

	tests := []struct {
		name  string
		input any
		want  string
	}{
		{
			name:  "whole float64 at top level",
			input: map[string]any{"a": float64(120), "b": float64(22)},
			want:  "a = 120\nb = 22\n",
		},
		{
			name:  "real float preserved",
			input: map[string]any{"pi": 3.14},
			want:  "pi = 3.14\n",
		},
		{
			name:  "negative whole float",
			input: map[string]any{"n": float64(-7)},
			want:  "n = -7\n",
		},
		{
			name:  "zero whole float",
			input: map[string]any{"z": float64(0)},
			want:  "z = 0\n",
		},
		{
			name: "nested map",
			input: map[string]any{
				"outer": map[string]any{"x": float64(10), "y": 1.5},
			},
			want: "[outer]\n  x = 10\n  y = 1.5\n",
		},
		{
			name:  "array of whole floats",
			input: map[string]any{"xs": []any{float64(1), float64(2), float64(3)}},
			want:  "xs = [1, 2, 3]\n",
		},
		{
			name:  "array of mixed floats",
			input: map[string]any{"xs": []any{float64(1), 1.5}},
			want:  "xs = [1.0, 1.5]\n", // BurntSushi promotes to floats when mixed
		},
		{
			name:  "float32 whole value",
			input: map[string]any{"f": float32(42)},
			want:  "f = 42\n",
		},
		{
			name:  "int64 unchanged",
			input: map[string]any{"i": int64(99)},
			want:  "i = 99\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("Marshal mismatch:\n  want: %q\n   got: %q", tt.want, string(got))
			}
		})
	}
}

func TestMarshalDoesNotMutateInput(t *testing.T) {
	c := Codec{}
	input := map[string]any{
		"a": float64(120),
		"nested": map[string]any{
			"b":  float64(5),
			"xs": []any{float64(1), 1.5},
		},
	}

	if _, err := c.Marshal(input); err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if _, ok := input["a"].(float64); !ok {
		t.Errorf("expected input[\"a\"] to remain float64, got %T", input["a"])
	}
	nested := input["nested"].(map[string]any)
	if _, ok := nested["b"].(float64); !ok {
		t.Errorf("expected nested[\"b\"] to remain float64, got %T", nested["b"])
	}
	xs := nested["xs"].([]any)
	if _, ok := xs[0].(float64); !ok {
		t.Errorf("expected xs[0] to remain float64, got %T", xs[0])
	}
}

func TestIsWholeFloat(t *testing.T) {
	tests := []struct {
		in   float64
		want bool
	}{
		{0, true},
		{1, true},
		{-1, true},
		{120, true},
		{1.5, false},
		{0.1, false},
		{math.NaN(), false},
		{math.Inf(1), false},
		{math.Inf(-1), false},
		{math.MaxInt64 + 1e20, false}, // out of int64 range
	}

	for _, tt := range tests {
		if got := isWholeFloat(tt.in); got != tt.want {
			t.Errorf("isWholeFloat(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestRoundTripPreservesIntegerType(t *testing.T) {
	// Simulate the bug-report flow: a value originating as float64 (as JSON
	// decode produces) marshals to TOML and round-trips back as an integer.
	c := Codec{}
	src := map[string]any{"a": float64(120), "b": float64(22)}

	encoded, err := c.Marshal(src)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if strings.Contains(string(encoded), ".0") {
		t.Errorf("expected no .0 in output, got %q", encoded)
	}

	var decoded map[string]any
	if err := c.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	for _, k := range []string{"a", "b"} {
		switch decoded[k].(type) {
		case int64, int:
			// ok
		default:
			t.Errorf("expected %s to decode as integer, got %T (%v)", k, decoded[k], decoded[k])
		}
	}
}
