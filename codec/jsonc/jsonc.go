package jsonc

import (
	"errors"
	"strings"

	"github.com/goccy/go-json"
)

// Codec handles JSON with Comments (JSONC) format
type Codec struct{}

// Unmarshal parses JSONC data by stripping comments and parsing as JSON
func (c *Codec) Unmarshal(data []byte, v interface{}) error {
	if v == nil {
		return errors.New("v cannot be nil")
	}

	// Strip comments from JSONC
	cleaned := c.stripComments(string(data))

	// Parse as regular JSON
	return json.Unmarshal([]byte(cleaned), v)
}

// Marshal converts data to JSON format (comments are not preserved in output)
func (c *Codec) Marshal(v interface{}) ([]byte, error) {
	// JSONC output is just pretty-printed JSON
	return json.MarshalIndent(v, "", "  ")
}

// stripComments removes single-line (//) and multi-line (/* */) comments from JSONC
// while preserving strings that may contain comment-like sequences
func (c *Codec) stripComments(input string) string {
	var result strings.Builder
	var i int
	inString := false
	escaped := false

	for i < len(input) {
		ch := input[i]

		// Handle escape sequences in strings
		if inString && escaped {
			result.WriteByte(ch)
			escaped = false
			i++
			continue
		}

		if inString && ch == '\\' {
			result.WriteByte(ch)
			escaped = true
			i++
			continue
		}

		// Track string boundaries
		if ch == '"' && !escaped {
			inString = !inString
			result.WriteByte(ch)
			i++
			continue
		}

		// If we're in a string, keep everything as-is
		if inString {
			result.WriteByte(ch)
			i++
			continue
		}

		// Check for single-line comment
		if ch == '/' && i+1 < len(input) && input[i+1] == '/' {
			// Skip until end of line
			i += 2
			for i < len(input) && input[i] != '\n' {
				i++
			}
			if i < len(input) {
				result.WriteByte('\n') // Preserve newline
				i++
			}
			continue
		}

		// Check for multi-line comment
		if ch == '/' && i+1 < len(input) && input[i+1] == '*' {
			// Skip until */
			i += 2
			for i < len(input)-1 {
				if input[i] == '*' && input[i+1] == '/' {
					i += 2
					break
				}
				// Preserve newlines in multi-line comments for line number accuracy
				if input[i] == '\n' {
					result.WriteByte('\n')
				}
				i++
			}
			continue
		}

		// Regular character
		result.WriteByte(ch)
		i++
	}

	return result.String()
}
