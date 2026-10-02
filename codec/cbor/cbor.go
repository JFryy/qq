package cbor

import (
	"reflect"

	"github.com/JFryy/qq/codec/util"
	"github.com/fxamacker/cbor/v2"
)

var decMode cbor.DecMode

func init() {
	var err error
	decMode, err = cbor.DecOptions{
		// Decode CBOR maps to map[string]any instead of map[interface{}]any
		// so downstream JSON marshaling works correctly.
		DefaultMapType: reflect.TypeOf(map[string]any{}),
	}.DecMode()
	if err != nil {
		panic(err)
	}
}

type Codec struct{}

func (c *Codec) Unmarshal(data []byte, v any) error {
	if err := decMode.Unmarshal(data, v); err != nil {
		return err
	}
	if ptr, ok := v.(*any); ok {
		*ptr = util.NormalizeNumbers(*ptr)
	}
	return nil
}

func (c *Codec) Marshal(v any) ([]byte, error) {
	return cbor.Marshal(v)
}
