package codec

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/goccy/go-json"
)

// Detect infers the encoding of input from its content. It backs the "auto"
// input type and is intentionally conservative: it only reports a format when a
// strict parse confirms it, and returns an error otherwise so callers can ask
// the user to be explicit rather than guess wrong.
//
// Detection covers the common structured text formats (json, xml/html, toml,
// yaml). Binary formats and ambiguous line based formats (csv, ini, env,
// properties, ...) are deliberately left out because their inputs overlap too
// much to tell apart reliably.
func Detect(input []byte) (EncodingType, error) {
	trimmed := bytes.TrimSpace(input)
	if len(trimmed) == 0 {
		return JSON, fmt.Errorf("cannot detect input format: empty input")
	}

	switch trimmed[0] {
	case '{', '[':
		if json.Valid(input) {
			return JSON, nil
		}
	case '<':
		lower := bytes.ToLower(trimmed)
		if bytes.HasPrefix(lower, []byte("<!doctype html")) || bytes.Contains(lower, []byte("<html")) {
			return HTML, nil
		}
		if _, err := tryDecode(input, XML); err == nil {
			return XML, nil
		}
		if _, err := tryDecode(input, HTML); err == nil {
			return HTML, nil
		}
	}

	// TOML is strict enough that arbitrary text will not parse, so try it before
	// falling back to YAML which accepts almost anything.
	if v, err := tryDecode(input, TOML); err == nil && isStructured(v) {
		return TOML, nil
	}

	// YAML is the catch all for structured config. Require a map or sequence so
	// that plain scalars and prose are not claimed as YAML.
	if v, err := tryDecode(input, YAML); err == nil && isStructured(v) {
		return YAML, nil
	}

	return JSON, fmt.Errorf("could not detect input format, please specify it with -i/--input")
}

func tryDecode(input []byte, t EncodingType) (any, error) {
	var v any
	if err := Codecs[t].Unmarshal(input, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func isStructured(v any) bool {
	if v == nil {
		return false
	}
	switch reflect.ValueOf(v).Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		return true
	default:
		return false
	}
}
