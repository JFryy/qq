package codec

import (
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/itchyny/gojq"
)

// queryNumbers returns every number in data as a JSON array, in gojq's order.
func queryNumbers(t *testing.T, data any) any {
	t.Helper()
	query, err := gojq.Parse(`[.. | numbers] | tojson`)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := query.Run(data).Next()
	if !ok {
		t.Fatal("query produced no result")
	}
	if err, isErr := got.(error); isErr {
		t.Fatal(err)
	}
	return got
}

func bigInt(t *testing.T, s string) *big.Int {
	t.Helper()
	n, ok := new(big.Int).SetString(s, 10)
	if !ok {
		t.Fatalf("invalid integer %q", s)
	}
	return n
}

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

func TestRoundTripsPreserveIntegersBeyondInt64(t *testing.T) {
	// Keys sort in the order the values are listed, so gojq reports them so.
	row := map[string]any{
		"a": 9223372036854775807,
		"b": bigInt(t, "9223372036854775808"),
		"c": -9223372036854775808,
		"d": bigInt(t, "-9223372036854775809"),
		"e": bigInt(t, "123456789012345678901234567890"),
	}
	want := "[9223372036854775807,9223372036854775808,-9223372036854775808,-9223372036854775809,123456789012345678901234567890]"
	for _, format := range []EncodingType{JSON, JSONC, JSONL, BASE64, YAML, HCL, CBOR, CSV, TSV, XML, GRON} {
		t.Run(format.String(), func(t *testing.T) {
			encoded, err := Marshal([]any{row}, format)
			if err != nil {
				t.Fatal(err)
			}
			var decoded any
			if err := Unmarshal(encoded, format, &decoded); err != nil {
				t.Fatal(err)
			}
			if got := queryNumbers(t, decoded); got != want {
				t.Fatalf("got %v, want %v\nencoded:\n%s", got, want, encoded)
			}
		})
	}
}

func TestMsgpackRoundTripsUint64(t *testing.T) {
	source := map[string]any{"max": bigInt(t, "18446744073709551615"), "min": -9223372036854775808}
	encoded, err := Marshal(source, MSGPACK)
	if err != nil {
		t.Fatal(err)
	}
	var decoded any
	if err := Unmarshal(encoded, MSGPACK, &decoded); err != nil {
		t.Fatal(err)
	}
	if got, want := queryNumbers(t, decoded), "[18446744073709551615,-9223372036854775808]"; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestMarshalRejectsIntegersOutOfFormatRange(t *testing.T) {
	for _, test := range []struct {
		format EncodingType
		value  string
	}{
		{TOML, "9223372036854775808"},
		{TOML, "-9223372036854775809"},
		{MSGPACK, "18446744073709551616"},
		{MSGPACK, "-9223372036854775809"},
	} {
		t.Run(test.format.String()+" "+test.value, func(t *testing.T) {
			_, err := Marshal(map[string]any{"n": []any{bigInt(t, test.value)}}, test.format)
			if err == nil || !strings.Contains(err.Error(), test.value) {
				t.Fatalf("got error %v, want one naming %s", err, test.value)
			}
		})
	}
}

func TestDecodeIntegersBeyondInt64(t *testing.T) {
	for _, test := range []struct {
		format EncodingType
		input  string
		want   string
	}{
		{JSON, `{"n": 12345678901234567890, "m": -9223372036854775809, "f": 1e30}`, "[1e+30,-9223372036854775809,12345678901234567890]"},
		{YAML, "n: 12345678901234567890\nm: -9223372036854775809\nf: 1e30\n", "[1e+30,-9223372036854775809,12345678901234567890]"},
		// Explicitly tagged and quoted scalars keep their types.
		{YAML, "f: !!float 123456789012345678901234567890\ns: \"123456789012345678901234567890\"\n", "[1.2345678901234568e+29]"},
		// Merged mappings and aliases are restored too.
		{YAML, "base: &b {n: 123456789012345678901234567890}\nmerged: {<<: *b}\nalias: *b\n", "[123456789012345678901234567890,123456789012345678901234567890,123456789012345678901234567890]"},
		{HCL, "n = 12345678901234567890\nm = -9223372036854775809\nf = 1e30\n", "[1e+30,-9223372036854775809,12345678901234567890]"},
		{INI, "[s]\nn = 12345678901234567890\n", "[12345678901234567890]"},
		{CSV, "n\n12345678901234567890\n9007199254740993\n", "[12345678901234567890,9007199254740993]"},
	} {
		t.Run(test.format.String(), func(t *testing.T) {
			var decoded any
			if err := Unmarshal([]byte(test.input), test.format, &decoded); err != nil {
				t.Fatal(err)
			}
			if got := queryNumbers(t, decoded); got != test.want {
				t.Fatalf("got %v, want %v", got, test.want)
			}
		})
	}
}

func TestDecodeLongIntegerLiterals(t *testing.T) {
	long := strings.Repeat("1", 200)
	tooLong := strings.Repeat("9", 400)
	for _, test := range []struct {
		format EncodingType
		input  string
		query  string
		want   any
	}{
		// HCL parses numbers with limited precision, so digits come from the source.
		{HCL, "n = " + long + "\nm = -" + long + "\n", `[.n, .m] | map(tostring)`, []any{long, "-" + long}},
		// Beyond the float64 range, cells stay strings as they always have.
		{CSV, "id\n" + tooLong + "\n", `.[0].id | type`, "string"},
	} {
		t.Run(test.format.String(), func(t *testing.T) {
			var decoded any
			if err := Unmarshal([]byte(test.input), test.format, &decoded); err != nil {
				t.Fatal(err)
			}
			query, err := gojq.Parse(test.query)
			if err != nil {
				t.Fatal(err)
			}
			got, _ := query.Run(decoded).Next()
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("got %v, want %v", got, test.want)
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
