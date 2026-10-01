package hcl_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/JFryy/qq/codec"
)

func TestHCLConversionPreservesFixtureValues(t *testing.T) {
	for _, format := range []string{"toml", "msgpack"} {
		t.Run(format, func(t *testing.T) {
			input, err := os.ReadFile(filepath.Join("..", "..", "tests", "test."+format))
			if err != nil {
				t.Fatal(err)
			}
			encoding, err := codec.GetEncodingType(format)
			if err != nil {
				t.Fatal(err)
			}
			var source any
			if err := codec.Unmarshal(input, encoding, &source); err != nil {
				t.Fatal(err)
			}
			// JSON establishes the expected portable representation of timestamps
			// and numeric types produced by the source decoder.
			jsonData, err := codec.Marshal(source, codec.JSON)
			if err != nil {
				t.Fatal(err)
			}
			var expected any
			if err := codec.Unmarshal(jsonData, codec.JSON, &expected); err != nil {
				t.Fatal(err)
			}
			for _, outputFormat := range []string{"hcl", "tf"} {
				t.Run(outputFormat, func(t *testing.T) {
					outputEncoding, err := codec.GetEncodingType(outputFormat)
					if err != nil {
						t.Fatal(err)
					}
					output, err := codec.Marshal(source, outputEncoding)
					if err != nil {
						t.Fatal(err)
					}
					var actual any
					if err := codec.Unmarshal(output, codec.HCL, &actual); err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(actual, expected) {
						t.Fatalf("conversion lost values:\nwant %#v\ngot %#v\n%s", expected, actual, output)
					}
				})
			}
		})
	}
}
