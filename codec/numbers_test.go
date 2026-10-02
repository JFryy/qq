package codec

import (
	"testing"

	"github.com/itchyny/gojq"
)

func TestRoundTripsPreserveLargeIntegers(t *testing.T) {
	source := map[string]any{"n": 9007199254740993, "nested": map[string]any{"m": -3}}
	for _, format := range []EncodingType{JSON, JSONC, JSONL, BASE64, YAML, TOML, HCL, MSGPACK, CBOR} {
		t.Run(format.String(), func(t *testing.T) {
			encoded, err := Marshal(source, format)
			if err != nil {
				t.Fatal(err)
			}
			var decoded any
			if err := Unmarshal(encoded, format, &decoded); err != nil {
				t.Fatal(err)
			}
			query, err := gojq.Parse(`[.. | numbers] | tojson`)
			if err != nil {
				t.Fatal(err)
			}
			got, ok := query.Run(decoded).Next()
			if !ok {
				t.Fatal("query produced no result")
			}
			if err, isErr := got.(error); isErr {
				t.Fatal(err)
			}
			if got != "[9007199254740993,-3]" {
				t.Fatalf("got %v, want [9007199254740993,-3]", got)
			}
		})
	}
}

func TestMarshalIsDeterministic(t *testing.T) {
	nested := map[string]any{}
	data := map[string]any{"nested": nested}
	for _, key := range []string{"k", "c", "x", "a", "m", "q", "e", "t", "b", "z"} {
		data[key] = key
		nested[key] = key
	}
	data["resource"] = map[string]any{"type": map[string]any{
		"zeta": []any{map[string]any{"a": 1}},
		"beta": []any{map[string]any{"b": 2}},
	}}
	flat := map[string]any{}
	for key, value := range nested {
		flat[key] = value
	}
	for _, test := range []struct {
		format EncodingType
		input  map[string]any
	}{{HCL, data}, {INI, data}, {GRON, data}, {ENV, flat}} {
		t.Run(test.format.String(), func(t *testing.T) {
			first, err := Marshal(test.input, test.format)
			if err != nil {
				t.Fatal(err)
			}
			for range 20 {
				next, err := Marshal(test.input, test.format)
				if err != nil {
					t.Fatal(err)
				}
				if string(next) != string(first) {
					t.Fatalf("output changed between runs:\n%s\n---\n%s", first, next)
				}
			}
		})
	}
}
