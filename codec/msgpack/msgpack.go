package msgpack

import (
	"github.com/JFryy/qq/codec/util"
	"github.com/vmihailenco/msgpack/v5"
)

type Codec struct{}

func (c *Codec) Unmarshal(data []byte, v any) error {
	if err := msgpack.Unmarshal(data, v); err != nil {
		return err
	}
	if ptr, ok := v.(*any); ok {
		*ptr = util.NormalizeNumbers(*ptr)
	}
	return nil
}

func (c *Codec) Marshal(v any) ([]byte, error) {
	return msgpack.Marshal(v)
}
