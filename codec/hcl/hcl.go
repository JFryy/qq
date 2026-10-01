package hcl

import (
	"fmt"
	"time"

	"github.com/goccy/go-json"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/tmccombs/hcl2json/convert"
	"github.com/zclconf/go-cty/cty"
)

type Codec struct{}

func (c *Codec) Unmarshal(input []byte, v any) error {
	opts := convert.Options{}
	content, err := convert.Bytes(input, "json", opts)
	if err != nil {
		return fmt.Errorf("error converting HCL to JSON: %v", err)
	}
	return json.Unmarshal(content, v)
}

func (c *Codec) Marshal(v any) ([]byte, error) {
	// Ensure the input is wrapped in a map if it's not already
	var data map[string]any
	switch v := v.(type) {
	case map[string]any:
		data = v
	default:
		data = map[string]any{
			"data": v,
		}
	}
	hclData, err := c.convertMapToHCL(data)
	if err != nil {
		return nil, fmt.Errorf("error converting map to HCL: %v", err)
	}

	return hclData, nil
}

func (c *Codec) convertMapToHCL(data map[string]any) ([]byte, error) {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	if err := c.populateBody(rootBody, data); err != nil {
		return nil, err
	}
	return f.Bytes(), nil
}

func (c *Codec) populateBody(body *hclwrite.Body, data map[string]any) error {
	for key, value := range data {
		if isBlockValue(value) {
			if err := c.appendBlocks(body, key, nil, value); err != nil {
				return fmt.Errorf("block %q: %w", key, err)
			}
			continue
		}
		attribute, err := c.convertToCtyValue(value)
		if err != nil {
			return fmt.Errorf("attribute %q: %w", key, err)
		}
		body.SetAttributeValue(key, attribute)
	}
	return nil
}

// isBlockValue recognizes hcl2json's block convention: nonempty arrays of
// objects, optionally nested beneath label keys. JSON alone cannot distinguish
// these from object-list attributes, so ambiguous values follow that convention.
func isBlockValue(value any) bool {
	switch v := value.(type) {
	case []any:
		if len(v) == 0 {
			return false
		}
		for _, item := range v {
			if _, ok := item.(map[string]any); !ok {
				return false
			}
		}
		return true
	case map[string]any:
		if len(v) == 0 {
			return false
		}
		for _, item := range v {
			if !isBlockValue(item) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// appendBlocks restores block labels and repeated bodies from hcl2json values.
func (c *Codec) appendBlocks(body *hclwrite.Body, name string, labels []string, value any) error {
	switch v := value.(type) {
	case map[string]any:
		for label, item := range v {
			if err := c.appendBlocks(body, name, append(labels, label), item); err != nil {
				return err
			}
		}
	case []any:
		for _, item := range v {
			block := body.AppendNewBlock(name, labels)
			if err := c.populateBody(block.Body(), item.(map[string]any)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Codec) convertToCtyValue(value any) (cty.Value, error) {
	switch v := value.(type) {
	case nil:
		return cty.NullVal(cty.DynamicPseudoType), nil
	case string:
		return cty.StringVal(v), nil
	case int:
		return cty.NumberIntVal(int64(v)), nil
	case int8:
		return cty.NumberIntVal(int64(v)), nil
	case int16:
		return cty.NumberIntVal(int64(v)), nil
	case int32:
		return cty.NumberIntVal(int64(v)), nil
	case int64:
		return cty.NumberIntVal(v), nil
	case uint:
		return cty.NumberUIntVal(uint64(v)), nil
	case uint8:
		return cty.NumberUIntVal(uint64(v)), nil
	case uint16:
		return cty.NumberUIntVal(uint64(v)), nil
	case uint32:
		return cty.NumberUIntVal(uint64(v)), nil
	case uint64:
		return cty.NumberUIntVal(v), nil
	case float32:
		return cty.NumberFloatVal(float64(v)), nil
	case float64:
		return cty.NumberFloatVal(v), nil
	case bool:
		return cty.BoolVal(v), nil
	case time.Time:
		// HCL has no timestamp type; match the timestamp's JSON representation.
		text, err := v.MarshalText()
		if err != nil {
			return cty.NilVal, fmt.Errorf("invalid timestamp: %w", err)
		}
		return cty.StringVal(string(text)), nil
	case []any:
		tuple := make([]cty.Value, len(v))
		for i, elem := range v {
			converted, err := c.convertToCtyValue(elem)
			if err != nil {
				return cty.NilVal, fmt.Errorf("element %d: %w", i, err)
			}
			tuple[i] = converted
		}
		return cty.TupleVal(tuple), nil
	case map[string]any:
		vals := make(map[string]cty.Value)
		for k, elem := range v {
			converted, err := c.convertToCtyValue(elem)
			if err != nil {
				return cty.NilVal, fmt.Errorf("key %q: %w", k, err)
			}
			vals[k] = converted
		}
		return cty.ObjectVal(vals), nil
	default:
		return cty.NilVal, fmt.Errorf("unsupported HCL value type %T", v)
	}
}
