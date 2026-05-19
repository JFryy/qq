package toml

import (
	"math"

	"github.com/BurntSushi/toml"
)

type Codec struct{}

func (c Codec) Unmarshal(data []byte, v any) error {
	return toml.Unmarshal(data, v)
}

// Marshal serializes v as TOML. JSON-derived inputs encode all numbers as
// float64, which would otherwise render whole-valued numbers as "120.0".
// Convert whole-valued float64 values to int64 first so the TOML output
// preserves integer typing for round-trips through type-checked TOML.
func (c Codec) Marshal(v any) ([]byte, error) {
	return toml.Marshal(normalizeIntegerFloats(v))
}

func normalizeIntegerFloats(val any) any {
	switch v := val.(type) {
	case float64:
		if isWholeFloat(v) {
			return int64(v)
		}
		return v
	case float32:
		f := float64(v)
		if isWholeFloat(f) {
			return int64(f)
		}
		return v
	case map[string]any:
		result := make(map[string]any, len(v))
		for k, item := range v {
			result[k] = normalizeIntegerFloats(item)
		}
		return result
	case []any:
		// Keep arrays homogeneous: if any element is a non-whole float, leave
		// every float element alone so we don't emit a mixed int/float array
		// (invalid in TOML <1.0, surprising in 1.0). Otherwise normalize.
		if hasNonWholeFloat(v) {
			result := make([]any, len(v))
			for i, item := range v {
				if _, isFloat := item.(float64); isFloat {
					result[i] = item
					continue
				}
				if _, isFloat := item.(float32); isFloat {
					result[i] = item
					continue
				}
				result[i] = normalizeIntegerFloats(item)
			}
			return result
		}
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = normalizeIntegerFloats(item)
		}
		return result
	default:
		return v
	}
}

func hasNonWholeFloat(items []any) bool {
	for _, item := range items {
		switch f := item.(type) {
		case float64:
			if !isWholeFloat(f) {
				return true
			}
		case float32:
			if !isWholeFloat(float64(f)) {
				return true
			}
		}
	}
	return false
}

func isWholeFloat(f float64) bool {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return false
	}
	if f != math.Trunc(f) {
		return false
	}
	// int64 range
	return f >= math.MinInt64 && f <= math.MaxInt64
}
