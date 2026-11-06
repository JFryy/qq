package hcl

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestUnmarshalValues(t *testing.T) {
	for _, test := range []struct {
		name   string
		source string
		want   map[string]any
	}{
		{
			name:   "literals",
			source: "int = 1\nfloat = -2.5\nexponent = 1e3\nbool = true\nnull_value = null\nstring = \"s\"\n",
			want: map[string]any{
				"int": 1, "float": -2.5, "exponent": 1000, "bool": true, "null_value": nil, "string": "s",
			},
		},
		{
			name:   "unary operators",
			source: "negative = -3\nnot = !true\nreference = -var.x\nparenthesized = -(var.x)\n",
			want:   map[string]any{"negative": -3, "not": false, "reference": "${-var.x}", "parenthesized": "${-(var.x)}"},
		},
		{
			name:   "numbers at the int64 boundary",
			source: "max = 9223372036854775807\nmin = -9223372036854775808\nover = 9223372036854775808\nwhole_float = 2.0\nfraction = 0.1\n",
			want: map[string]any{
				"max": math.MaxInt64, "min": math.MinInt64, "over": 9223372036854775808.0, "whole_float": 2, "fraction": 0.1,
			},
		},
		{
			name: "expressions keep source text",
			source: `reference = var.x
call = upper("x")
nested_call = max(min(1, 2), 3)
conditional = var.on ? 1 : 2
comprehension = [for v in var.list : v * 2]
index = var.list[0]
parentheses = (1 + 2)
parenthesized_index = var.list[(0)]
splat = var.list[*].id
`,
			want: map[string]any{
				"reference":           "${var.x}",
				"call":                `${upper("x")}`,
				"nested_call":         "${max(min(1, 2), 3)}",
				"conditional":         "${var.on ? 1 : 2}",
				"comprehension":       "${[for v in var.list : v * 2]}",
				"index":               "${var.list[0]}",
				"parentheses":         "${(1 + 2)}",
				"parenthesized_index": "${var.list[(0)]}",
				"splat":               "${var.list[*].id}",
			},
		},
		{
			name: "templates",
			source: `interpolation = "ami-${var.x}"
wrapped = "${var.x}"
escaped = "lit $${x} %%{y}"
if_else = "%{if var.on}yes%{else}no%{endif}"
if_only = "%{if var.on}yes%{endif}"
for_value = "%{for v in var.list}${v},%{endfor}"
for_key_value = "%{for k, v in var.map}${k}=${v}%{endfor}"
ternary_interpolation = "x ${var.on ? "y" : "z"} w"
ternary_wrapped = "${var.on ? 1 : 2}"
nested_template = "a ${upper("b${var.c}")} d"
wrapped_literal = "${1}"
strip_markers = " a %{~ if var.on ~} b %{~ endif ~} c "
heredoc = <<EOT
line ${var.z}
EOT
`,
			want: map[string]any{
				"interpolation":         "ami-${var.x}",
				"wrapped":               "${var.x}",
				"escaped":               "lit $${x} %%{y}",
				"if_else":               "%{if var.on}yes%{else}no%{endif}",
				"if_only":               "%{if var.on}yes%{endif}",
				"for_value":             "%{for v in var.list}${v},%{endfor}",
				"for_key_value":         "%{for k, v in var.map}${k}=${v}%{endfor}",
				"ternary_interpolation": `x ${var.on ? "y" : "z"} w`,
				"ternary_wrapped":       "${var.on ? 1 : 2}",
				"nested_template":       `a ${upper("b${var.c}")} d`,
				"wrapped_literal":       1,
				"strip_markers":         " a%{if var.on}b%{endif}c ",
				"heredoc":               "line ${var.z}\n",
			},
		},
		{
			name: "collections",
			source: `empty_list = []
list = [1, "two", [3], { four = 4 }]
empty_object = {}
object = {
  ident = 1
  "quoted key" = 2
  env.x = 3
  (var.k) = 4
  "${var.t}" = 5
  true = 6
  7 = 7
}
`,
			want: map[string]any{
				"empty_list":   []any{},
				"list":         []any{1, "two", []any{3}, map[string]any{"four": 4}},
				"empty_object": map[string]any{},
				"object": map[string]any{
					"ident": 1, "quoted key": 2, "env.x": 3, "${(var.k)}": 4, "${var.t}": 5, "true": 6, "7": 7,
				},
			},
		},
		{
			name: "blocks",
			source: `locals { a = 1 }
locals { b = 2 }
terraform {}
resource "aws_instance" "web" {
  ami = "x"
  lifecycle { create_before_destroy = true }
}
resource "aws_instance" "db" {}
module "m" { source = "./a" }
module "m" { source = "./b" }
`,
			want: map[string]any{
				"locals":    []any{map[string]any{"a": 1}, map[string]any{"b": 2}},
				"terraform": []any{map[string]any{}},
				"resource": map[string]any{
					"aws_instance": map[string]any{
						"web": []any{map[string]any{
							"ami":       "x",
							"lifecycle": []any{map[string]any{"create_before_destroy": true}},
						}},
						"db": []any{map[string]any{}},
					},
				},
				"module": map[string]any{
					"m": []any{map[string]any{"source": "./a"}, map[string]any{"source": "./b"}},
				},
			},
		},
		{
			name: "flush heredoc with directives",
			source: `doc = <<-EOT
    %{ for v in var.l ~}
      item ${v} ${var.a ? "y" : "z"}
    %{ endfor ~}
    done
    EOT
`,
			want: map[string]any{
				// hclsyntax keeps this heredoc's indentation; HCL's own
				// evaluation gives the same text.
				"doc": `%{for v in var.l}      item ${v} ${var.a ? "y" : "z"}` + "\n%{endfor}    done\n",
			},
		},
		{
			name:   "mixed label depths under one type",
			source: "a \"x\" { n = 1 }\na \"y\" \"z\" { n = 2 }\na \"y\" \"z\" { n = 3 }\na \"x\" { n = 4 }\n",
			want: map[string]any{"a": map[string]any{
				"x": []any{map[string]any{"n": 1}, map[string]any{"n": 4}},
				"y": map[string]any{"z": []any{map[string]any{"n": 2}, map[string]any{"n": 3}}},
			}},
		},
		{
			name:   "attribute replaces block of the same name",
			source: "clash { a = 1 }\nclash = 2\n",
			want:   map[string]any{"clash": 2},
		},
		{
			name:   "empty input",
			source: "",
			want:   map[string]any{},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var got any
			if err := (&Codec{}).Unmarshal([]byte(test.source), &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, any(test.want)) {
				t.Fatalf("got %#v\nwant %#v", got, test.want)
			}
		})
	}
}

func TestUnmarshalErrors(t *testing.T) {
	for _, test := range []struct {
		name   string
		source string
		want   []string
	}{
		{
			name:   "parse error reports position without a filename",
			source: "a = \n",
			want:   []string{"error converting HCL to JSON: line 1, column 5: Invalid expression"},
		},
		{
			name:   "parse error on a later line",
			source: "a = 1\nb = = 2\n",
			want:   []string{"line 2, column 5: Invalid expression"},
		},
		{
			name:   "unlabeled then labeled block",
			source: "a {}\na \"x\" {}\n",
			want:   []string{`block a "x": cannot mix labeled and unlabeled blocks`},
		},
		{
			name:   "labeled then unlabeled block",
			source: "a \"x\" {}\na {}\n",
			want:   []string{"block a: cannot mix labeled and unlabeled blocks"},
		},
		{
			name:   "error inside nested block names the path",
			source: "outer \"o\" {\n  inner {}\n  inner \"i\" {}\n}\n",
			want:   []string{`block outer "o": block inner "i": cannot mix labeled and unlabeled blocks`},
		},
		{
			name:   "number too large for float64",
			source: "a = 1e400\n",
			want:   []string{`attribute "a": number 1e+400 is out of range`},
		},
		{
			name:   "negative number too large for float64",
			source: "a = -1e400\n",
			want:   []string{`attribute "a": number -1e+400 is out of range`},
		},
		{
			name:   "invalid unary literal",
			source: "a = -true\n",
			want:   []string{`attribute "a": line 1, column 6: Invalid operand`},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var got any
			err := (&Codec{}).Unmarshal([]byte(test.source), &got)
			if err == nil {
				t.Fatalf("expected error, got %#v", got)
			}
			for _, want := range test.want {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("error %q does not contain %q", err, want)
				}
			}
			if strings.Contains(err.Error(), "json:") {
				t.Fatalf("error %q contains a misleading filename", err)
			}
		})
	}
}

func TestUnmarshalIntoTypedTarget(t *testing.T) {
	var got struct {
		Name  string         `json:"name"`
		Count int            `json:"count"`
		Tags  map[string]any `json:"tags"`
	}
	source := "name = \"web\"\ncount = 2\ntags = { env = var.env }\n"
	if err := (&Codec{}).Unmarshal([]byte(source), &got); err != nil {
		t.Fatal(err)
	}
	if got.Name != "web" || got.Count != 2 || got.Tags["env"] != "${var.env}" {
		t.Fatalf("got %#v", got)
	}
}

func TestUnmarshalTypedTargetKeepsExactNumbers(t *testing.T) {
	var got struct {
		Max    uint64      `json:"max"`
		Number json.Number `json:"number"`
	}
	source := "max = 18446744073709551615\nnumber = 12345678901234567890.5\n"
	if err := (&Codec{}).Unmarshal([]byte(source), &got); err != nil {
		t.Fatal(err)
	}
	if got.Max != math.MaxUint64 || got.Number != "12345678901234567890.5" {
		t.Fatalf("got %#v", got)
	}
}
