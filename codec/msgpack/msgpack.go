package msgpack

import (
	"github.com/vmihailenco/msgpack/v5"
)

type Codec struct{}

func (c *Codec) Unmarshal(data []byte, v any) error {
	return msgpack.Unmarshal(data, v)
}

func (c *Codec) Marshal(v any) ([]byte, error) {
	return msgpack.Marshal(v)
}
