package xml

import (
	"fmt"
	"github.com/JFryy/qq/codec/util"
	"github.com/clbanning/mxj/v2"
	"reflect"
)

type Codec struct{}

func (c *Codec) Marshal(v interface{}) ([]byte, error) {
	switch v := v.(type) {
	case map[string]interface{}:
		mv := mxj.Map(v)
		return mv.XmlIndent("", "  ")
	case []interface{}:
		mv := mxj.Map(map[string]interface{}{"root": v})
		return mv.XmlIndent("", "  ")
	default:
		mv := mxj.Map(map[string]interface{}{"value": v})
		return mv.XmlIndent("", "  ")
	}
}

func (c *Codec) Unmarshal(input []byte, v interface{}) error {
	mv, err := mxj.NewMapXml(input)
	if err != nil {
		return fmt.Errorf("error unmarshaling XML: %v", err)
	}

	parsedData := c.parseXMLValues(mv.Old())

	// reflection of values required for type assertions on interface
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("provided value must be a non-nil pointer")
	}
	rv.Elem().Set(reflect.ValueOf(parsedData))

	return nil
}

// infer the type of the value and parse it accordingly
func (c *Codec) parseXMLValues(v interface{}) interface{} {
	switch v := v.(type) {
	case map[string]interface{}:
		for key, val := range v {
			v[key] = c.parseXMLValues(val)
		}
		return v
	case []interface{}:
		for i, val := range v {
			v[i] = c.parseXMLValues(val)
		}
		return v
	case string:
		return util.ParseValue(v)
	default:
		return v
	}
}
