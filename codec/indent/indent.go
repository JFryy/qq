// Package indent holds the output indentation width chosen with --indent.
//
// Encoders that pretty-print (JSON, JSONC and YAML) read the width when they
// marshal, so the CLI sets it once instead of passing it through every codec.
package indent

import (
	"fmt"
	"strings"
	"sync/atomic"
)

const (
	// Default is the width used when --indent is not given.
	Default = 2
	// Max is the largest accepted width, matching jq.
	Max = 7
)

var width atomic.Int32

func init() {
	width.Store(Default)
}

// Validate reports whether n is an accepted width.
func Validate(n int) error {
	if n < 0 || n > Max {
		return fmt.Errorf("--indent must be between 0 and %d", Max)
	}
	return nil
}

// Set changes the width. It returns an error and leaves the width unchanged
// when n is out of range.
func Set(n int) error {
	if err := Validate(n); err != nil {
		return err
	}
	width.Store(int32(n))
	return nil
}

// Width returns the current width. Zero means compact output.
func Width() int {
	return int(width.Load())
}

// String returns the indentation unit: Width spaces.
func String() string {
	return strings.Repeat(" ", Width())
}

// Reset restores Default.
func Reset() {
	width.Store(Default)
}
