package hcl

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/tmccombs/hcl2json/convert"
	"github.com/zclconf/go-cty/cty"
	"log"
)

type Codec struct{}

func (c *Codec) Unmarshal(input []byte, v interface{}) error {
	opts := convert.Options{}
	content, err := convert.Bytes(input, "json", opts)
	if err != nil {
		return fmt.Errorf("error converting HCL to JSON: %v", err)
	}
	return json.Unmarshal(content, v)
}

func (c *Codec) Marshal(v interface{}) ([]byte, error) {
	// Ensure the input is wrapped in a map if it's not already
	var data map[string]interface{}
	switch v := v.(type) {
	case map[string]interface{}:
		data = v
	default:
		data = map[string]interface{}{
			"data": v,
		}
	}
	hclData, err := c.convertMapToHCL(data)
	if err != nil {
		return nil, fmt.Errorf("error converting map to HCL: %v", err)
	}

	return hclData, nil
}

func (c *Codec) convertMapToHCL(data map[string]interface{}) ([]byte, error) {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	c.populateBody(rootBody, data)
	return f.Bytes(), nil
}

func (c *Codec) populateBody(body *hclwrite.Body, data map[string]interface{}) {
	for key, value := range data {
		switch v := value.(type) {
		case map[string]interface{}:
			block := body.AppendNewBlock(key, nil)
			c.populateBody(block.Body(), v)

		case []interface{}:
			if len(v) == 1 {
				if singleMap, ok := v[0].(map[string]interface{}); ok {
					block := body.AppendNewBlock(key, nil)
					c.populateBody(block.Body(), singleMap)
					continue
				}
			}
			if len(v) == 0 {
				continue
			}
			tuple := make([]cty.Value, len(v))
			for i, elem := range v {
				tuple[i] = c.convertToCtyValue(elem)
			}
			body.SetAttributeValue(key, cty.TupleVal(tuple))

		case string:
			body.SetAttributeValue(key, cty.StringVal(v))
		case int:
			body.SetAttributeValue(key, cty.NumberIntVal(int64(v)))
		case int64:
			body.SetAttributeValue(key, cty.NumberIntVal(v))
		case float64:
			body.SetAttributeValue(key, cty.NumberFloatVal(v))
		case bool:
			body.SetAttributeValue(key, cty.BoolVal(v))
		default:
			log.Printf("Unsupported type: %T", v)
		}
	}
}

func (c *Codec) convertToCtyValue(value interface{}) cty.Value {
	switch v := value.(type) {
	case string:
		return cty.StringVal(v)
	case int:
		return cty.NumberIntVal(int64(v))
	case int64:
		return cty.NumberIntVal(v)
	case float64:
		return cty.NumberFloatVal(v)
	case bool:
		return cty.BoolVal(v)
	case []interface{}:
		tuple := make([]cty.Value, len(v))
		for i, elem := range v {
			tuple[i] = c.convertToCtyValue(elem)
		}
		return cty.TupleVal(tuple)
	case map[string]interface{}:
		vals := make(map[string]cty.Value)
		for k, elem := range v {
			vals[k] = c.convertToCtyValue(elem)
		}
		return cty.ObjectVal(vals)
	default:
		log.Printf("Unsupported type: %T", v)
		return cty.NilVal
	}
}
