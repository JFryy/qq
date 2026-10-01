package hcl

import (
	"fmt"
	"math"
	"os"
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

func TestExpressionRoundTripSemantics(t *testing.T) {
	fixture, err := os.ReadFile("../../tests/test.tf")
	if err != nil {
		t.Fatal(err)
	}
	context := &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"var": cty.ObjectVal(map[string]cty.Value{
				"name": cty.StringVal("world"), "enabled": cty.True,
				"instance_type": cty.StringVal("t3.micro"),
				"values":        cty.TupleVal([]cty.Value{cty.NumberIntVal(1), cty.NumberIntVal(2)}),
			}),
			"aws_instance": cty.ObjectVal(map[string]cty.Value{
				"example": cty.ObjectVal(map[string]cty.Value{
					"id": cty.StringVal("i-123"), "public_ip": cty.StringVal("192.0.2.1"),
				}),
			}),
			"data": cty.ObjectVal(map[string]cty.Value{
				"aws_ami": cty.ObjectVal(map[string]cty.Value{
					"latest_amazon_linux": cty.ObjectVal(map[string]cty.Value{"id": cty.StringVal("ami-123")}),
				}),
			}),
		},
		Functions: map[string]function.Function{"upper": stdlib.UpperFunc},
	}
	for _, test := range []struct{ name, source string }{
		{"repository Terraform fixture", string(fixture)},
		{"expressions and literals", `reference = var.name
number = var.values[0] + 1
boolean = var.enabled
list = var.values
object = var
conditional = var.enabled ? "yes" : "no"
call = upper("hello")
comprehension = [for value in var.values : value * 2]
wrapped_object = ({ key = [var.name, { value = var.enabled }] })
computed_key = { (var.name) = var.enabled }
static_keys = { "literal.key" = 1, "$${literal}" = 2 }
template = "hello ${upper(var.name)}!"
quoted_expression = "${upper("hello")}"
nested = { reference = var.name, list = [var.enabled, var.values] }
literal = "keep $${var.name} and %%{if true}"
plain = "quote: \" slash: \\ newline:\n tab:\t"
directive = "%{if var.enabled}yes%{else}no%{endif}"
loop = "%{for value in var.values}${value},%{endfor}"
trimmed = " before %{~ if var.enabled ~} yes %{~ endif ~} after "
mixed = "literal $${name}; actual ${var.name}; \"${var.name}\"; slash \\${var.name}; snowman ☃"
heredoc = <<EOF
hello ${var.name}
literal $${var.name}
EOF
`},
	} {
		t.Run(test.name, func(t *testing.T) {
			var decoded any
			codec := &Codec{}
			if err := codec.Unmarshal([]byte(test.source), &decoded); err != nil {
				t.Fatal(err)
			}
			output, err := codec.Marshal(decoded)
			if err != nil {
				t.Fatal(err)
			}
			original := parseTestBody(t, []byte(test.source))
			roundTrip := parseTestBody(t, output)
			compareBodyValues(t, original, roundTrip, context)
		})
	}
}

func TestMalformedExpressionsReturnErrors(t *testing.T) {
	for _, value := range []any{"${var.name", "${var.}", "%{if true}missing endif", math.NaN(), math.Inf(1), float32(math.Inf(-1))} {
		if _, err := (&Codec{}).Marshal(map[string]any{"value": value}); err == nil {
			t.Fatalf("expected error for %#v", value)
		}
	}
}

func parseTestBody(t *testing.T, source []byte) *hclsyntax.Body {
	t.Helper()
	file, diagnostics := hclsyntax.ParseConfig(source, "test.tf", hcl.InitialPos)
	if diagnostics.HasErrors() {
		t.Fatalf("invalid HCL: %s\n%s", diagnostics, source)
	}
	return file.Body.(*hclsyntax.Body)
}

// compareBodyValues checks block structure and evaluated attribute types/values,
// ignoring formatting and declaration order.
func compareBodyValues(t *testing.T, original, actual *hclsyntax.Body, context *hcl.EvalContext) {
	t.Helper()
	if len(original.Attributes) != len(actual.Attributes) || len(original.Blocks) != len(actual.Blocks) {
		t.Fatal("round trip changed the number of attributes or blocks")
	}
	for name, attribute := range original.Attributes {
		other, ok := actual.Attributes[name]
		if !ok {
			t.Fatalf("missing attribute %q", name)
		}
		want, diagnostics := attribute.Expr.Value(context)
		if diagnostics.HasErrors() {
			t.Fatal(diagnostics)
		}
		got, diagnostics := other.Expr.Value(context)
		if diagnostics.HasErrors() {
			t.Fatal(diagnostics)
		}
		if !got.RawEquals(want) {
			t.Errorf("attribute %q changed meaning: got %s, want %s", name, got.GoString(), want.GoString())
		}
	}
	blocks := make(map[string][]*hclsyntax.Block)
	for _, block := range actual.Blocks {
		key := fmt.Sprintf("%s:%q", block.Type, block.Labels)
		blocks[key] = append(blocks[key], block)
	}
	for _, block := range original.Blocks {
		key := fmt.Sprintf("%s:%q", block.Type, block.Labels)
		matches := blocks[key]
		if len(matches) == 0 {
			t.Fatalf("missing block %s", key)
		}
		compareBodyValues(t, block.Body, matches[0].Body, context)
		blocks[key] = matches[1:]
	}
}
