package json

import (
	"reflect"
	"testing"
)

func TestUnmarshalPreservesIntegers(t *testing.T) {
	var got any
	input := `{"big":9007199254740993,"neg":-9223372036854775808,"float":1.5,"whole":1.0,"exp":1e3,"list":[1,{"n":2}]}`
	if err := Unmarshal([]byte(input), &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"big":   9007199254740993,
		"neg":   -9223372036854775808,
		"float": 1.5,
		"whole": float64(1),
		"exp":   float64(1000),
		"list":  []any{1, map[string]any{"n": 2}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestUnmarshalRejectsInvalidInput(t *testing.T) {
	for _, input := range []string{`{} {}`, `{"a":1e999}`, `{`} {
		var got any
		if err := Unmarshal([]byte(input), &got); err == nil {
			t.Errorf("Unmarshal(%q) succeeded, want error", input)
		}
	}
}

func TestUnmarshalTypedDestination(t *testing.T) {
	var got map[string]float64
	if err := Unmarshal([]byte(`{"n":2}`), &got); err != nil {
		t.Fatal(err)
	}
	if got["n"] != 2 {
		t.Fatalf("got %v, want 2", got["n"])
	}
}
