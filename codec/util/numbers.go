package util

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
)

// NormalizeNumbers converts sized Go integer and float types from binary
// decoders into int, *big.Int and float64, the numeric types gojq accepts.
func NormalizeNumbers(value any) any {
	switch v := value.(type) {
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		if v >= math.MinInt && v <= math.MaxInt {
			return int(v)
		}
		return big.NewInt(v)
	case uint:
		return normalizeUint(uint64(v))
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return normalizeUint(uint64(v))
	case uint64:
		return normalizeUint(v)
	case float32:
		return float64(v)
	case big.Int:
		return NormalizeBigInt(&v)
	case *big.Int:
		return NormalizeBigInt(v)
	case map[string]any:
		for key, item := range v {
			v[key] = NormalizeNumbers(item)
		}
	case []any:
		for i, item := range v {
			v[i] = NormalizeNumbers(item)
		}
	}
	return value
}

func normalizeUint(v uint64) any {
	if v <= math.MaxInt {
		return int(v)
	}
	return new(big.Int).SetUint64(v)
}

// NormalizeBigInt returns n as an int when it fits, and n itself otherwise.
func NormalizeBigInt(n *big.Int) any {
	if n.IsInt64() {
		if i := n.Int64(); i >= math.MinInt && i <= math.MaxInt {
			return int(i)
		}
	}
	return n
}

// ParseInteger parses s as a base 10 integer literal. Literals beyond the
// float64 range are rejected: they were not numbers before integers were kept
// exact, and the JSON encoder cannot write them.
func ParseInteger(s string) (*big.Int, bool) {
	if _, err := strconv.ParseFloat(s, 64); err != nil {
		return nil, false
	}
	return new(big.Int).SetString(s, 10)
}

// ReplaceBigInts returns v with each *big.Int replaced by replace(n). Containers
// are copied only when something inside them is replaced, so v is not modified.
func ReplaceBigInts(v any, replace func(*big.Int) (any, error)) (any, error) {
	out, _, err := replaceBigInts(v, replace)
	return out, err
}

func replaceBigInts(v any, replace func(*big.Int) (any, error)) (any, bool, error) {
	switch val := v.(type) {
	case *big.Int:
		replaced, err := replace(val)
		return replaced, true, err
	case map[string]any:
		var out map[string]any
		for key, item := range val {
			replaced, changed, err := replaceBigInts(item, replace)
			if err != nil {
				return nil, false, err
			}
			if changed && out == nil {
				out = make(map[string]any, len(val))
				for k, existing := range val {
					out[k] = existing
				}
			}
			if changed {
				out[key] = replaced
			}
		}
		if out != nil {
			return out, true, nil
		}
	case []any:
		var out []any
		for i, item := range val {
			replaced, changed, err := replaceBigInts(item, replace)
			if err != nil {
				return nil, false, err
			}
			if changed && out == nil {
				out = make([]any, len(val))
				copy(out, val)
			}
			if changed {
				out[i] = replaced
			}
		}
		if out != nil {
			return out, true, nil
		}
	}
	return v, false, nil
}

// FixedSizeInt returns n as an int64 or uint64 for formats whose integers are
// at most 64 bits, or an error naming the format when n does not fit.
func FixedSizeInt(n *big.Int, format string) (any, error) {
	if n.IsInt64() {
		return n.Int64(), nil
	}
	if n.IsUint64() {
		return n.Uint64(), nil
	}
	return nil, fmt.Errorf("integer %s is out of range for %s", n, format)
}
