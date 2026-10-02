package util

import "math"

// NormalizeNumbers converts sized Go integer and float types from binary
// decoders into int and float64, the numeric types gojq accepts.
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
		return float64(v)
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
	return float64(v)
}
