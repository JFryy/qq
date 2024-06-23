package codec

import (
	"fmt"
	"github.com/clbanning/mxj/v2"
)

func xmlMarshal(v interface{}) ([]byte, error) {
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

func xmlUnmarshal(input []byte, v interface{}) error {
	mv, err := mxj.NewMapXml(input)
	if err != nil {
		return fmt.Errorf("error unmarshaling XML: %v", err)
	}
	*v.(*interface{}) = mv.Old()
	return nil
}
