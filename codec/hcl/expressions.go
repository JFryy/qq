package hcl

import (
	"bytes"
	"fmt"
	"sort"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// expressionTokens reconstructs expressions encoded by hcl2json, including
// expressions nested in object attributes and tuple elements.
func (c *Codec) expressionTokens(value any) (hclwrite.Tokens, error) {
	switch v := value.(type) {
	case string:
		return templateTokens(v)
	case []any:
		elements := make([]hclwrite.Tokens, len(v))
		for i, item := range v {
			tokens, err := c.expressionTokens(item)
			if err != nil {
				return nil, fmt.Errorf("element %d: %w", i, err)
			}
			elements[i] = tokens
		}
		return hclwrite.TokensForTuple(elements), nil
	case map[string]any:
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		attributes := make([]hclwrite.ObjectAttrTokens, 0, len(v))
		for _, key := range keys {
			name, err := objectKeyTokens(key)
			if err != nil {
				return nil, fmt.Errorf("object key %q: %w", key, err)
			}
			tokens, err := c.expressionTokens(v[key])
			if err != nil {
				return nil, fmt.Errorf("key %q: %w", key, err)
			}
			attributes = append(attributes, hclwrite.ObjectAttrTokens{Name: name, Value: tokens})
		}
		return hclwrite.TokensForObject(attributes), nil
	default:
		literal, err := c.convertToCtyValue(value)
		if err != nil {
			return nil, err
		}
		return hclwrite.TokensForValue(literal), nil
	}
}

// objectKeyTokens preserves static names and parenthesizes computed keys so
// HCL does not interpret a traversal as a literal object attribute name.
func objectKeyTokens(key string) (hclwrite.Tokens, error) {
	if hclsyntax.ValidIdentifier(key) {
		return hclwrite.TokensForIdentifier(key), nil
	}
	tokens, err := templateTokens(key)
	if err != nil {
		return nil, err
	}
	if len(tokens) > 0 && tokens[0].Type == hclsyntax.TokenOQuote {
		return tokens, nil
	}
	return append(append(hclwrite.Tokens{{Type: hclsyntax.TokenOParen, Bytes: []byte("(")}}, tokens...),
		&hclwrite.Token{Type: hclsyntax.TokenCParen, Bytes: []byte(")")}), nil
}

// templateTokens interprets hcl2json's template encoding, rather than escaping
// interpolations as string data. Literal introducers remain escaped.
func templateTokens(value string) (hclwrite.Tokens, error) {
	source := []byte(value)
	expression, diagnostics := hclsyntax.ParseTemplate(source, "template", hcl.InitialPos)
	if diagnostics.HasErrors() {
		return nil, fmt.Errorf("invalid HCL template: %s", diagnostics)
	}
	if wrapped, ok := expression.(*hclsyntax.TemplateWrapExpr); ok {
		return parsedExpressionTokens(wrapped.Wrapped.Range().SliceBytes(source))
	}
	if literal, ok := expression.(*hclsyntax.TemplateExpr); ok && literal.IsStringLiteral() {
		value, diagnostics := literal.Value(nil)
		if diagnostics.HasErrors() {
			return nil, fmt.Errorf("invalid template literal: %s", diagnostics)
		}
		return hclwrite.TokensForValue(value), nil
	}
	lexerTokens, diagnostics := hclsyntax.LexTemplate(source, "template", hcl.InitialPos)
	if diagnostics.HasErrors() {
		return nil, fmt.Errorf("invalid HCL template: %s", diagnostics)
	}
	var quoted bytes.Buffer
	quoted.WriteByte('"')
	depth, offset := 0, 0
	for _, token := range lexerTokens {
		if token.Type == hclsyntax.TokenEOF {
			break
		}
		quoted.Write(source[offset:token.Range.Start.Byte])
		if token.Type == hclsyntax.TokenStringLit && depth == 0 {
			// Decode escaped introducers before applying quoted-string escaping.
			literal, diagnostics := hclsyntax.ParseTemplate(token.Bytes, "literal", hcl.InitialPos)
			if diagnostics.HasErrors() {
				return nil, fmt.Errorf("invalid template literal: %s", diagnostics)
			}
			value, diagnostics := literal.Value(nil)
			if diagnostics.HasErrors() {
				return nil, fmt.Errorf("invalid template literal: %s", diagnostics)
			}
			tokens := hclwrite.TokensForValue(cty.StringVal(value.AsString()))
			quoted.Write(tokens[1 : len(tokens)-1].Bytes())
		} else {
			quoted.Write(token.Bytes)
		}
		switch token.Type {
		case hclsyntax.TokenTemplateInterp, hclsyntax.TokenTemplateControl:
			depth++
		case hclsyntax.TokenTemplateSeqEnd:
			depth--
		}
		offset = token.Range.End.Byte
	}
	quoted.WriteByte('"')
	return parsedExpressionTokens(quoted.Bytes())
}

// parsedExpressionTokens validates generated source before handing raw tokens
// to hclwrite, which otherwise accepts invalid expressions unchecked.
func parsedExpressionTokens(source []byte) (hclwrite.Tokens, error) {
	if _, diagnostics := hclsyntax.ParseExpression(source, "expression", hcl.InitialPos); diagnostics.HasErrors() {
		return nil, fmt.Errorf("invalid HCL expression: %s", diagnostics)
	}
	assignment := append([]byte("value = "), source...)
	assignment = append(assignment, '\n')
	file, diagnostics := hclwrite.ParseConfig(assignment, "expression", hcl.InitialPos)
	if diagnostics.HasErrors() {
		return nil, fmt.Errorf("invalid HCL expression: %s", diagnostics)
	}
	return file.Body().GetAttribute("value").Expr().BuildTokens(nil), nil
}
