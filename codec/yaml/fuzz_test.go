package yaml

import (
	"fmt"
	"math/big"
	"os"
	"testing"
)

// FuzzUnmarshal checks that any input either fails to decode or decodes to
// values gojq accepts, which YAML can then encode.
//
//	go test ./codec/yaml -run='^$' -fuzz=FuzzUnmarshal -fuzztime=60s
func FuzzUnmarshal(f *testing.F) {
	fixture, err := os.ReadFile("../../tests/test.yaml")
	if err != nil {
		f.Fatal(err)
	}
	f.Add(fixture)
	for _, seed := range []string{
		"",
		"a: 1",
		"- 1\n- [2, {b: c}]",
		"---\na: 1\n---\nb: 2\n",
		"base: &b {x: 1}\nmerged: {<<: *b, y: 2}\nalias: *b\n",
		"1: int key\ntrue: bool key\n~: null key\n",
		"date: 2024-01-02\ntime: 2024-01-02T03:04:05Z\n",
		"big: 12345678901234567890\nfloat: 1.5e300\nhex: 0x1F\n",
		"text: |\n  line one\n  line two\n",
		"bin: !!binary aGVsbG8=\n",
	} {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		var v any
		if err := (Codec{}).Unmarshal(data, &v); err != nil {
			return
		}
		if err := checkQueryValue(v); err != nil {
			t.Fatalf("decoded value cannot be queried: %v", err)
		}
		if _, err := (Codec{}).Marshal(v); err != nil {
			t.Fatalf("decoded value cannot be encoded: %v", err)
		}
	})
}

// checkQueryValue reports a value of a type gojq does not accept; gojq panics
// on such values.
func checkQueryValue(v any) error {
	switch v := v.(type) {
	case nil, bool, int, float64, *big.Int, string:
		return nil
	case []any:
		for i, item := range v {
			if err := checkQueryValue(item); err != nil {
				return fmt.Errorf("[%d]: %w", i, err)
			}
		}
		return nil
	case map[string]any:
		for key, item := range v {
			if err := checkQueryValue(item); err != nil {
				return fmt.Errorf("%q: %w", key, err)
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported type %T", v)
	}
}
