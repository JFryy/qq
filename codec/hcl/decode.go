package hcl

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/JFryy/qq/codec/util"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// Unmarshal decodes HCL into plain Go values. Blocks are grouped by type and
// then by each label, ending in an array of bodies so repeated blocks are
// kept. Literals become values; any other expression becomes a "${...}"
// string holding its source text, which Marshal turns back into HCL.
func (c *Codec) Unmarshal(input []byte, v any) error {
	file, diags := hclsyntax.ParseConfig(input, "", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return fmt.Errorf("error converting HCL to JSON: %w", diagnosticsError(diags))
	}
	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return fmt.Errorf("error converting HCL to JSON: unexpected body type %T", file.Body)
	}
	ptr, generic := v.(*any)
	d := decoder{source: file.Bytes, number: goNumber}
	if !generic {
		// Typed targets decode exact numbers, as encoding/json would.
		d.number = exactNumber
	}
	value, err := d.body(body)
	if err != nil {
		return fmt.Errorf("error converting HCL to JSON: %w", err)
	}
	if generic {
		*ptr = value
		return nil
	}
	content, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(content, v)
}

type decoder struct {
	source []byte
	// number converts a numeric literal; digits is its source text when that
	// is a plain integer, without a sign, fraction or exponent, and "" otherwise.
	number func(n *big.Float, digits string) (any, error)
}

func (d *decoder) body(body *hclsyntax.Body) (map[string]any, error) {
	out := make(map[string]any, len(body.Attributes)+len(body.Blocks))
	for _, block := range body.Blocks {
		content, err := d.body(block.Body)
		if err != nil {
			return nil, fmt.Errorf("block %s: %w", blockPath(block), err)
		}
		if err := insertBlock(out, block, content); err != nil {
			return nil, err
		}
	}
	// Attributes are applied after blocks, so an attribute wins a name clash.
	for name, attribute := range body.Attributes {
		value, err := d.value(attribute.Expr)
		if err != nil {
			return nil, fmt.Errorf("attribute %q: %w", name, err)
		}
		out[name] = value
	}
	return out, nil
}

// insertBlock stores content at out[type][label1]...[labelN], which holds an
// array of every body declared with that type and those labels.
func insertBlock(out map[string]any, block *hclsyntax.Block, content map[string]any) error {
	path := append([]string{block.Type}, block.Labels...)
	parent := out
	for _, name := range path[:len(path)-1] {
		child, exists := parent[name]
		if !exists {
			child = make(map[string]any)
			parent[name] = child
		}
		next, ok := child.(map[string]any)
		if !ok {
			return fmt.Errorf("block %s: cannot mix labeled and unlabeled blocks", blockPath(block))
		}
		parent = next
	}
	last := path[len(path)-1]
	switch existing := parent[last].(type) {
	case nil:
		parent[last] = []any{content}
	case []any:
		parent[last] = append(existing, content)
	default:
		return fmt.Errorf("block %s: cannot mix labeled and unlabeled blocks", blockPath(block))
	}
	return nil
}

func blockPath(block *hclsyntax.Block) string {
	path := block.Type
	for _, label := range block.Labels {
		path += fmt.Sprintf(" %q", label)
	}
	return path
}

func (d *decoder) value(expr hclsyntax.Expression) (any, error) {
	switch e := expr.(type) {
	case *hclsyntax.LiteralValueExpr:
		return d.literal(e.Val, d.text(e.SrcRange))
	case *hclsyntax.UnaryOpExpr:
		// Negative numbers and negated booleans are literals in practice.
		if operand, literal := e.Val.(*hclsyntax.LiteralValueExpr); literal {
			result, diags := e.Value(nil)
			if diags.HasErrors() {
				return nil, diagnosticsError(diags)
			}
			return d.literal(result, d.text(operand.SrcRange))
		}
	case *hclsyntax.TemplateWrapExpr:
		// "${expr}" is the expression itself rather than a string.
		return d.value(e.Wrapped)
	case *hclsyntax.TemplateExpr:
		return d.template(e)
	case *hclsyntax.TupleConsExpr:
		items := make([]any, len(e.Exprs))
		for i, item := range e.Exprs {
			value, err := d.value(item)
			if err != nil {
				return nil, fmt.Errorf("index %d: %w", i, err)
			}
			items[i] = value
		}
		return items, nil
	case *hclsyntax.ObjectConsExpr:
		object := make(map[string]any, len(e.Items))
		for _, item := range e.Items {
			key, err := d.objectKey(item.KeyExpr)
			if err != nil {
				return nil, err
			}
			value, err := d.value(item.ValueExpr)
			if err != nil {
				return nil, fmt.Errorf("key %q: %w", key, err)
			}
			object[key] = value
		}
		return object, nil
	}
	return d.interpolation(expr), nil
}

// objectKey keeps bare and dotted keys as written; any other key is rendered
// like a template, so computed keys such as (var.k) become "${(var.k)}".
func (d *decoder) objectKey(expr hclsyntax.Expression) (string, error) {
	if key, ok := expr.(*hclsyntax.ObjectConsKeyExpr); ok {
		if _, bare := key.Wrapped.(*hclsyntax.ScopeTraversalExpr); bare {
			return d.text(key.Wrapped.Range()), nil
		}
		expr = key.Wrapped
	}
	return d.template(expr)
}

// template renders expr as HCL template text: literal text is escaped so it
// is not read back as interpolation, and everything else is written as
// interpolation or directive syntax.
func (d *decoder) template(expr hclsyntax.Expression) (string, error) {
	var out strings.Builder
	if err := d.writeTemplate(&out, expr); err != nil {
		return "", err
	}
	return out.String(), nil
}

func (d *decoder) writeTemplate(out *strings.Builder, expr hclsyntax.Expression) error {
	switch e := expr.(type) {
	case *hclsyntax.LiteralValueExpr:
		text, err := literalText(e.Val)
		if err != nil {
			return err
		}
		out.WriteString(escapeTemplate(text))
	case *hclsyntax.TemplateExpr:
		for _, part := range e.Parts {
			if err := d.writeTemplate(out, part); err != nil {
				return err
			}
		}
	case *hclsyntax.TemplateWrapExpr:
		return d.writeTemplate(out, e.Wrapped)
	case *hclsyntax.ConditionalExpr:
		// Only %{if} directives become directives; ${a ? b : c} is kept as
		// an interpolation.
		if !d.isDirective(e) {
			out.WriteString(d.interpolation(e))
			return nil
		}
		return d.writeIf(out, e)
	case *hclsyntax.TemplateJoinExpr:
		if loop, ok := e.Tuple.(*hclsyntax.ForExpr); ok {
			return d.writeFor(out, loop)
		}
		out.WriteString(d.interpolation(e))
	default:
		out.WriteString(d.interpolation(expr))
	}
	return nil
}

func (d *decoder) writeIf(out *strings.Builder, e *hclsyntax.ConditionalExpr) error {
	whenTrue, err := d.template(e.TrueResult)
	if err != nil {
		return err
	}
	whenFalse, err := d.template(e.FalseResult)
	if err != nil {
		return err
	}
	out.WriteString("%{if " + d.text(e.Condition.Range()) + "}" + whenTrue)
	if whenFalse != "" {
		out.WriteString("%{else}" + whenFalse)
	}
	out.WriteString("%{endif}")
	return nil
}

func (d *decoder) writeFor(out *strings.Builder, e *hclsyntax.ForExpr) error {
	content, err := d.template(e.ValExpr)
	if err != nil {
		return err
	}
	vars := e.ValVar
	if e.KeyVar != "" {
		vars = e.KeyVar + ", " + e.ValVar
	}
	out.WriteString("%{for " + vars + " in " + d.text(e.CollExpr.Range()) + "}" + content + "%{endfor}")
	return nil
}

func (d *decoder) isDirective(expr hclsyntax.Expression) bool {
	return strings.HasPrefix(d.text(expr.Range()), "%{")
}

func (d *decoder) interpolation(expr hclsyntax.Expression) string {
	return "${" + d.text(expr.Range()) + "}"
}

func (d *decoder) text(r hcl.Range) string {
	return string(d.source[r.Start.Byte:r.End.Byte])
}

// literal converts a literal value; text is its source text.
func (d *decoder) literal(value cty.Value, text string) (any, error) {
	if value.IsNull() {
		return nil, nil
	}
	switch value.Type() {
	case cty.String:
		return value.AsString(), nil
	case cty.Bool:
		return value.True(), nil
	case cty.Number:
		digits := ""
		if isIntegerLiteral(text) {
			digits = text
		}
		return d.number(value.AsBigFloat(), digits)
	default:
		return nil, fmt.Errorf("unsupported literal of type %s", value.Type().FriendlyName())
	}
}

// goNumber matches the JSON codec's numbers: whole numbers within int range
// become int, larger integer literals *big.Int, and others float64.
func goNumber(n *big.Float, digits string) (any, error) {
	if n.IsInt() {
		if i, accuracy := n.Int64(); accuracy == big.Exact && i >= math.MinInt && i <= math.MaxInt {
			return int(i), nil
		}
		// Parse the source text: n has limited precision, so long literals
		// would lose digits.
		if i, ok := util.ParseInteger(digits); ok {
			if n.Sign() < 0 {
				i.Neg(i)
			}
			return i, nil
		}
	}
	f, _ := n.Float64()
	if math.IsInf(f, 0) {
		return nil, fmt.Errorf("number %s is out of range", n.Text('g', 10))
	}
	return f, nil
}

func exactNumber(n *big.Float, _ string) (any, error) {
	return json.Number(n.Text('f', -1)), nil
}

// isIntegerLiteral reports whether text is written as decimal digits alone.
func isIntegerLiteral(text string) bool {
	if text == "" {
		return false
	}
	for _, r := range text {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// literalText is the text a literal contributes to a template.
func literalText(value cty.Value) (string, error) {
	if value.IsNull() {
		return "null", nil
	}
	switch value.Type() {
	case cty.String:
		return value.AsString(), nil
	case cty.Bool:
		if value.True() {
			return "true", nil
		}
		return "false", nil
	case cty.Number:
		return value.AsBigFloat().Text('f', -1), nil
	default:
		return "", fmt.Errorf("unsupported template literal of type %s", value.Type().FriendlyName())
	}
}

func escapeTemplate(text string) string {
	return strings.NewReplacer("${", "$${", "%{", "%%{").Replace(text)
}

// diagnosticsError reports error diagnostics by position. Input has no
// meaningful filename, so hcl's "file:line,col" prefix is omitted.
func diagnosticsError(diags hcl.Diagnostics) error {
	var messages []string
	for _, diag := range diags {
		if diag.Severity != hcl.DiagError {
			continue
		}
		message := diag.Summary
		if diag.Detail != "" {
			message += ": " + diag.Detail
		}
		if diag.Subject != nil {
			message = fmt.Sprintf("line %d, column %d: %s", diag.Subject.Start.Line, diag.Subject.Start.Column, message)
		}
		messages = append(messages, message)
	}
	return errors.New(strings.Join(messages, "; "))
}
