package util

import (
	"sync"
)

var (
	keyOrderingMu sync.RWMutex
	keyOrdering   = make(map[uintptr][]string)

	// PreserveKeyOrder determines if the JSON and YAML codecs preserve the original key order.
	PreserveKeyOrder bool
)

// SetKeyOrder records the original order of keys for a map pointer.
func SetKeyOrder(ptr uintptr, keys []string) {
	keyOrderingMu.Lock()
	keyOrdering[ptr] = keys
	keyOrderingMu.Unlock()
}

// GetKeyOrder returns the recorded key order for a map pointer if it exists.
func GetKeyOrder(ptr uintptr) ([]string, bool) {
	keyOrderingMu.RLock()
	keys, ok := keyOrdering[ptr]
	keyOrderingMu.RUnlock()
	return keys, ok
}

// ClearKeyOrder resets the key order registry to prevent memory growth.
func ClearKeyOrder() {
	keyOrderingMu.Lock()
	keyOrdering = make(map[uintptr][]string)
	keyOrderingMu.Unlock()
}
