package cli

import (
	"testing"

	"github.com/JFryy/qq/codec"
	"github.com/JFryy/qq/codec/indent"
)

func TestE2E_Indent(t *testing.T) {
	t.Cleanup(indent.Reset)

	tests := []struct {
		name             string
		input            string
		width            int
		outputFormat     codec.EncodingType
		preserveKeyOrder bool
		want             string
	}{
		{"compact json", `{"a":{"b":1}}`, 0, codec.JSON, false, "{\"a\":{\"b\":1}}\n"},
		{"default json", `{"a":{"b":1}}`, indent.Default, codec.JSON, false, "{\n  \"a\": {\n    \"b\": 1\n  }\n}\n"},
		{"four json with key order", `{"z":1,"a":{"y":2,"b":3}}`, 4, codec.JSON, true, "{\n    \"z\": 1,\n    \"a\": {\n        \"y\": 2,\n        \"b\": 3\n    }\n}\n"},
		{"compact json with key order", `{"z":1,"a":{"y":2,"b":3}}`, 0, codec.JSON, true, "{\"z\":1,\"a\":{\"y\":2,\"b\":3}}\n"},
		{"four yaml", `{"a":{"b":1}}`, 4, codec.YAML, false, "a:\n    b: 1\n\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := indent.Set(tt.width); err != nil {
				t.Fatal(err)
			}
			got, err := runPipeline(t, []byte(tt.input), codec.JSON, tt.outputFormat, ".", tt.preserveKeyOrder, false)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("output = %q, want %q", got, tt.want)
			}
		})
	}
}
