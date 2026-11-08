package msgpack

import (
	"math/big"

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
	// MessagePack integers are at most 64 bits; the encoder would otherwise
	// write larger ones as strings.
	v, err := util.ReplaceBigInts(v, func(n *big.Int) (any, error) {
		return util.FixedSizeInt(n, "MessagePack")
	})
	if err != nil {
		return nil, err
	}
	return msgpack.Marshal(v)
}
