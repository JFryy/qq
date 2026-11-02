// Package keyorder remembers the key order of objects decoded from input, so
// encoders can write objects that a query passes through unchanged in their
// original order.
//
// Order is attached to a specific map value, not to its contents. gojq builds a
// new map for any object a query creates or modifies, and such maps have no
// recorded order, so encoders fall back to sorted keys for them.
package keyorder

import (
	"reflect"
	"slices"
	"sync"
)

type entry struct {
	// m keeps the map reachable for as long as its order is recorded, so its
	// address cannot be reused by an unrelated map.
	m    map[string]any
	keys []string
}

var (
	mu        sync.RWMutex
	entries   map[uintptr]entry // nil while disabled
	recording bool
)

// Enable starts a new, empty registry and begins recording.
func Enable() {
	mu.Lock()
	defer mu.Unlock()
	entries = make(map[uintptr]entry)
	recording = true
}

// Disable discards the registry. Encoders return to sorted keys.
func Disable() {
	mu.Lock()
	defer mu.Unlock()
	entries = nil
	recording = false
}

// StopRecording keeps the recorded orders but ignores later Record calls, so
// decoding done after the input is read, such as re-parsing output for raw
// formatting, does not grow the registry.
func StopRecording() {
	mu.Lock()
	defer mu.Unlock()
	recording = false
}

// Enabled reports whether encoders should write keys in recorded order.
func Enabled() bool {
	mu.RLock()
	defer mu.RUnlock()
	return entries != nil
}

// Recording reports whether decoders should record key order.
func Recording() bool {
	mu.RLock()
	defer mu.RUnlock()
	return recording
}

// Record sets the key order of m. keys may contain duplicates or keys that are
// not in m; Keys ignores both.
func Record(m map[string]any, keys []string) {
	if len(m) < 2 {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	if !recording {
		return
	}
	entries[address(m)] = entry{m: m, keys: keys}
}

// RecordRows sets keys as the order of every object in rows, for formats where
// one header row orders all records.
func RecordRows(rows any, keys []string) {
	items, ok := rows.([]any)
	if !ok {
		return
	}
	for _, item := range items {
		if m, ok := item.(map[string]any); ok {
			Record(m, keys)
		}
	}
}

// Keys returns the keys of m in recorded order, followed by any keys that have
// no recorded position, sorted. Without a recorded order all keys are sorted.
func Keys(m map[string]any) []string {
	var recorded []string
	if len(m) > 1 {
		mu.RLock()
		if e, ok := entries[address(m)]; ok {
			recorded = e.keys
		}
		mu.RUnlock()
	}

	keys := make([]string, 0, len(m))
	seen := make(map[string]bool, len(m))
	for _, k := range recorded {
		if _, ok := m[k]; ok && !seen[k] {
			seen[k] = true
			keys = append(keys, k)
		}
	}
	rest := make([]string, 0, len(m)-len(keys))
	for k := range m {
		if !seen[k] {
			rest = append(rest, k)
		}
	}
	slices.Sort(rest)
	return append(keys, rest...)
}

func address(m map[string]any) uintptr {
	return uintptr(reflect.ValueOf(m).UnsafePointer())
}
