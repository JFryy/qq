package cli

import (
	"strings"
	"testing"

	"github.com/JFryy/qq/codec"
	"github.com/JFryy/qq/codec/keyorder"
	"github.com/itchyny/gojq"
)

// runPreservingKeyOrder decodes input with key order recording on, runs query,
// and encodes each result, as handleCommand does with --preserve-key-order.
func runPreservingKeyOrder(t *testing.T, input string, in codec.EncodingType, query string, out codec.EncodingType, slurp bool) string {
	t.Helper()
	keyorder.Enable()
	t.Cleanup(keyorder.Disable)

	var data any
	var err error
	if slurp {
		data, err = slurpInputs([]byte(input), in)
	} else {
		err = codec.Unmarshal([]byte(input), in, &data)
	}
	if err != nil {
		t.Fatalf("decoding input: %v", err)
	}
	keyorder.StopRecording()

	parsed, err := gojq.Parse(query)
	if err != nil {
		t.Fatalf("parsing query: %v", err)
	}
	var results []string
	iter := parsed.Run(data)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			t.Fatalf("running query: %v", err)
		}
		b, err := codec.Marshal(v, out)
		if err != nil {
			t.Fatalf("encoding output: %v", err)
		}
		results = append(results, strings.TrimSpace(string(b)))
	}
	return strings.Join(results, "\n")
}

func TestPreserveKeyOrder(t *testing.T) {
	tests := []struct {
		name  string
		input string
		in    codec.EncodingType
		query string
		out   codec.EncodingType
		slurp bool
		want  string
	}{
		{
			name:  "json nested objects and arrays",
			input: `{"z":1,"a":{"y":[{"q":1,"c":2}],"b":"<&>"}}`,
			in:    codec.JSON, query: ".", out: codec.JSON,
			want: "{\n  \"z\": 1,\n  \"a\": {\n    \"y\": [\n      {\n        \"q\": 1,\n        \"c\": 2\n      }\n    ],\n    \"b\": \"<&>\"\n  }\n}",
		},
		{
			name:  "json path into unchanged object",
			input: `{"x":{"z":1,"a":2}}`,
			in:    codec.JSON, query: ".x", out: codec.JSON,
			want: "{\n  \"z\": 1,\n  \"a\": 2\n}",
		},
		{
			name:  "json del falls back to sorted keys",
			input: `{"z":1,"m":2,"a":3}`,
			in:    codec.JSON, query: "del(.m)", out: codec.JSON,
			want: "{\n  \"a\": 3,\n  \"z\": 1\n}",
		},
		{
			name:  "json modified object falls back to sorted keys",
			input: `{"z":1,"a":2}`,
			in:    codec.JSON, query: ".z = 5", out: codec.JSON,
			want: "{\n  \"a\": 2,\n  \"z\": 5\n}",
		},
		{
			name:  "json constructed object falls back to sorted keys",
			input: `{"z":1,"a":2}`,
			in:    codec.JSON, query: "{z, a}", out: codec.JSON,
			want: "{\n  \"a\": 2,\n  \"z\": 1\n}",
		},
		{
			name:  "json duplicate key keeps first position and last value",
			input: `{"z":1,"a":2,"z":3}`,
			in:    codec.JSON, query: ".", out: codec.JSON,
			want: "{\n  \"z\": 3,\n  \"a\": 2\n}",
		},
		{
			name:  "json slurp",
			input: `{"z":1,"a":2} {"y":1,"b":2}`,
			in:    codec.JSON, query: ".", out: codec.JSONL, slurp: true,
			want: "{\"z\":1,\"a\":2}\n{\"y\":1,\"b\":2}",
		},
		{
			name:  "jsonl",
			input: "{\"z\":1,\"a\":\"<\"}\n{\"y\":1,\"b\":2}\n",
			in:    codec.JSONL, query: ".", out: codec.JSONL,
			want: "{\"z\":1,\"a\":\"\\u003c\"}\n{\"y\":1,\"b\":2}",
		},
		{
			name:  "yaml multi-document",
			input: "z: 1\na: 2\n---\ny: 1\nb: 2\n",
			in:    codec.YAML, query: ".[1]", out: codec.YAML,
			want: "\"y\": 1\nb: 2",
		},
		{
			name:  "yaml aliases and merge keys",
			input: "base: &base\n  z: 1\n  a: 2\ncopy: *base\nmerged:\n  m: 0\n  <<: *base\n  a: 9\n",
			in:    codec.YAML, query: ".", out: codec.YAML,
			want: "base:\n  z: 1\n  a: 2\ncopy:\n  z: 1\n  a: 2\nmerged:\n  m: 0\n  z: 1\n  a: 9",
		},
		{
			name:  "yaml non-string keys",
			input: "2: b\n1: a\n",
			in:    codec.YAML, query: ".", out: codec.JSON,
			want: "{\n  \"2\": \"b\",\n  \"1\": \"a\"\n}",
		},
		{
			name:  "xml attributes, text and repeated elements",
			input: `<r z="1" a="2"><y>1</y><b k="v">text</b><y>2</y></r>`,
			in:    codec.XML, query: ".", out: codec.XML,
			want: "<r z=\"1\" a=\"2\">\n  <y>1</y>\n  <y>2</y>\n  <b k=\"v\">text</b>\n</r>",
		},
		{
			name:  "xml to json",
			input: `<r z="1" a="2"><y>1</y><b>x</b></r>`,
			in:    codec.XML, query: ".r", out: codec.JSON,
			want: "{\n  \"-z\": 1,\n  \"-a\": 2,\n  \"y\": 1,\n  \"b\": \"x\"\n}",
		},
		{
			name:  "csv header order",
			input: "z,a\n1,2\n3,4\n",
			in:    codec.CSV, query: ".", out: codec.CSV,
			want: "z,a\n1,2\n3,4",
		},
		{
			name:  "tsv header order",
			input: "z\ta\n1\t2\n",
			in:    codec.TSV, query: ".", out: codec.TSV,
			want: "z\ta\n1\t2",
		},
		{
			name:  "json to csv",
			input: `[{"z":1,"a":2}]`,
			in:    codec.JSON, query: ".", out: codec.CSV,
			want: "z,a\n1,2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runPreservingKeyOrder(t, tt.input, tt.in, tt.query, tt.out, tt.slurp)
			if got != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestPreserveKeyOrderDisabledSortsKeys(t *testing.T) {
	keyorder.Disable()
	var data any
	if err := codec.Unmarshal([]byte(`{"z":1,"a":2}`), codec.JSON, &data); err != nil {
		t.Fatalf("decoding input: %v", err)
	}
	b, err := codec.Marshal(data, codec.JSON)
	if err != nil {
		t.Fatalf("encoding output: %v", err)
	}
	if want := "{\n  \"a\": 2,\n  \"z\": 1\n}"; strings.TrimSpace(string(b)) != want {
		t.Errorf("got:\n%s\nwant:\n%s", b, want)
	}
}

func TestPreserveKeyOrderStopsRecordingAfterInput(t *testing.T) {
	keyorder.Enable()
	t.Cleanup(keyorder.Disable)
	keyorder.StopRecording()

	var data any
	if err := codec.Unmarshal([]byte(`{"z":1,"a":2}`), codec.JSON, &data); err != nil {
		t.Fatalf("decoding input: %v", err)
	}
	if got := keyorder.Keys(data.(map[string]any)); got[0] != "a" {
		t.Errorf("expected no recorded order after StopRecording, got %v", got)
	}
}

func TestPreserveKeyOrderRejectsTrailingJSON(t *testing.T) {
	keyorder.Enable()
	t.Cleanup(keyorder.Disable)
	var data any
	if err := codec.Unmarshal([]byte(`{"z":1} {"a":2}`), codec.JSON, &data); err == nil {
		t.Fatal("expected an error for trailing JSON input")
	}
}
