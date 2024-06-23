package codec

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/tmccombs/hcl2json/convert"
	"github.com/zclconf/go-cty/cty"
	"log"
)

func hclUnmarshal(input []byte, v interface{}) error {
	opts := convert.Options{}
	content, err := convert.Bytes(input, "json", opts)
	if err != nil {
		return fmt.Errorf("error converting HCL to JSON: %v", err)
	}
	return json.Unmarshal(content, v)
}

func hclMarshal(v interface{}) ([]byte, error) {
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

	// Convert map to HCL
	hclData, err := convertMapToHCL(data)
	if err != nil {
		return nil, fmt.Errorf("error converting map to HCL: %v", err)
	}

	return hclData, nil
}

func convertMapToHCL(data map[string]interface{}) ([]byte, error) {
	// Create a new HCL file
	f := hclwrite.NewEmptyFile()

	// Create the root body of the file
	rootBody := f.Body()

	// Populate the body with the data
	populateBody(rootBody, data)

	// Return the HCL data
	return f.Bytes(), nil
}

func populateBody(body *hclwrite.Body, data map[string]interface{}) {
	for key, value := range data {
		switch v := value.(type) {
		case map[string]interface{}:
			block := body.AppendNewBlock(key, nil)
			populateBody(block.Body(), v)
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
		case []interface{}:
			tuple := make([]cty.Value, len(v))
			for i, elem := range v {
				tuple[i] = convertToCtyValue(elem)
			}
			body.SetAttributeValue(key, cty.TupleVal(tuple))
		default:
			log.Printf("Unsupported type: %T", v)
		}
	}
}

func convertToCtyValue(value interface{}) cty.Value {
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
			tuple[i] = convertToCtyValue(elem)
		}
		return cty.TupleVal(tuple)
	case map[string]interface{}:
		vals := make(map[string]cty.Value)
		for k, elem := range v {
			vals[k] = convertToCtyValue(elem)
		}
		return cty.ObjectVal(vals)
	default:
		log.Printf("Unsupported type: %T", v)
		return cty.NilVal
	}
}
