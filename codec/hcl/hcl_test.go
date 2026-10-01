package hcl

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func TestObjectAttributeRoundTrip(t *testing.T) {
	for _, test := range []struct {
		name   string
		input  string
		labels []string
	}{
		{
			name: "required providers",
			input: `terraform {
  required_version = "~> 1.0"
  required_providers {
    aws = {
      source = "hashicorp/aws"
      version = "~> 6.0"
    }
  }
}`,
		},
		{
			name: "nested object attributes",
			input: `configuration = {
  enabled = true
  retries = 3
  nested = { name = "example" }
  empty = {}
  values = []
  optional = null
  entries = [{ name = "first" }, { name = "second" }]
}`,
		},
		{
			name: "one label",
			input: `output "example" {
  value = "constant"
}`,
			labels: []string{"example"},
		},
		{
			name: "two labels and repeated blocks",
			input: `data "aws_ami" "example" {
  owners = ["amazon"]
  tags = { Name = "example" }
  filter {
    name = "name"
    values = ["example-*"]
  }
  filter {
    name = "architecture"
    values = ["arm64"]
  }
}`,
			labels: []string{"aws_ami", "example"},
		},
		{
			name: "empty collections and null",
			input: `empty = {}
values = []
optional = null`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			codec := &Codec{}
			var original any
			if err := codec.Unmarshal([]byte(test.input), &original); err != nil {
				t.Fatal(err)
			}
			output, err := codec.Marshal(original)
			if err != nil {
				t.Fatal(err)
			}
			parsed, diagnostics := hclsyntax.ParseConfig(output, "roundtrip.tf", hcl.InitialPos)
			if diagnostics.HasErrors() {
				t.Fatalf("invalid HCL: %s\n%s", diagnostics, output)
			}
			if test.labels != nil {
				blocks := parsed.Body.(*hclsyntax.Body).Blocks
				if len(blocks) != 1 || !reflect.DeepEqual(blocks[0].Labels, test.labels) {
					t.Fatalf("block labels were not preserved:\n%s", output)
				}
			}
			var roundTrip any
			if err := codec.Unmarshal(output, &roundTrip); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(original, roundTrip) {
				t.Fatalf("round trip changed object attributes:\n%s\nwant %#v\ngot %#v", output, original, roundTrip)
			}
		})
	}
}

func TestObjectArraysFollowBlockConvention(t *testing.T) {
	// A JSON object array could also be an attribute; without a schema it is
	// interpreted as repeated blocks, following hcl2json's representation.
	input := map[string]any{"entry": []any{
		map[string]any{"name": "first"},
		map[string]any{"name": "second"},
	}}
	output, err := (&Codec{}).Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	parsed, diagnostics := hclsyntax.ParseConfig(output, "output.hcl", hcl.InitialPos)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}
	body := parsed.Body.(*hclsyntax.Body)
	if len(body.Blocks) != 2 || len(body.Attributes) != 0 {
		t.Fatalf("expected two blocks:\n%s", output)
	}
}

func TestMarshalUnsupportedValueReturnsError(t *testing.T) {
	for _, value := range []any{
		map[string]any{"attribute": make(chan int)},
		map[string]any{"object": map[string]any{"attribute": make(chan int)}},
		map[string]any{"block": []any{map[string]any{"attribute": make(chan int)}}},
	} {
		if _, err := (&Codec{}).Marshal(value); err == nil {
			t.Fatal("expected an error for unsupported HCL value")
		}
	}
}
