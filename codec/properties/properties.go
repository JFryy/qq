package properties

import (
	"bufio"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/goccy/go-json"
)

// Codec handles Java properties file parsing and marshaling
type Codec struct{}

// Unmarshal parses properties file data into the provided interface
func (c *Codec) Unmarshal(data []byte, v interface{}) error {
	if v == nil {
		return errors.New("v cannot be nil")
	}

	result, err := c.Parse(string(data))
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, v)
}

// Marshal converts data back to properties file format
func (c *Codec) Marshal(v interface{}) ([]byte, error) {
	// Convert to our expected format first
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var props map[string]string
	if err := json.Unmarshal(data, &props); err != nil {
		return nil, errors.New("properties format only supports simple key-value pairs, cannot convert complex nested structures")
	}

	var lines []string
	// Sort keys for consistent output
	keys := make([]string, 0, len(props))
	for key := range props {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := props[key]
		lines = append(lines, fmt.Sprintf("%s=%s", c.escapeKey(key), c.escapeValue(value)))
	}

	return []byte(strings.Join(lines, "\n")), nil
}

// Parse processes properties file content into key-value pairs
func (c *Codec) Parse(content string) (map[string]string, error) {
	result := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(content))

	var continuedLine strings.Builder
	var continuedKey string

	for scanner.Scan() {
		line := scanner.Text()

		// If we have a continued line from previous iteration
		if continuedLine.Len() > 0 {
			line = continuedLine.String() + line
			continuedLine.Reset()
		}

		// Check if line ends with backslash (continuation)
		if strings.HasSuffix(strings.TrimRightFunc(line, unicode.IsSpace), "\\") {
			// Remove trailing backslash and spaces before it
			line = strings.TrimRightFunc(line, unicode.IsSpace)
			line = line[:len(line)-1]
			continuedLine.WriteString(line)
			continue
		}

		// Skip empty lines and comments
		trimmed := strings.TrimLeftFunc(line, unicode.IsSpace)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "!") {
			continue
		}

		// Parse key-value pair
		key, value, found := c.parseKeyValue(line)
		if found {
			if continuedKey != "" {
				// This is continuation of a previous value
				result[continuedKey] = result[continuedKey] + value
				continuedKey = ""
			} else {
				result[key] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning properties: %v", err)
	}

	return result, nil
}

// parseKeyValue extracts key and value from a properties line
// Handles: key=value, key:value, key value
func (c *Codec) parseKeyValue(line string) (string, string, bool) {
	// Skip leading whitespace
	line = strings.TrimLeftFunc(line, unicode.IsSpace)
	if line == "" {
		return "", "", false
	}

	var key strings.Builder
	var value string
	escaped := false
	i := 0

	// Parse key (until we hit separator or space)
	for i < len(line) {
		ch := line[i]

		if escaped {
			key.WriteByte(ch)
			escaped = false
			i++
			continue
		}

		if ch == '\\' {
			escaped = true
			i++
			continue
		}

		// Check for separator
		if ch == '=' || ch == ':' || unicode.IsSpace(rune(ch)) {
			// Found separator
			i++
			// Skip any additional separators or whitespace
			for i < len(line) && (line[i] == '=' || line[i] == ':' || unicode.IsSpace(rune(line[i]))) {
				i++
			}
			value = line[i:]
			break
		}

		key.WriteByte(ch)
		i++
	}

	if key.Len() == 0 {
		return "", "", false
	}

	return c.unescapeString(key.String()), c.unescapeString(strings.TrimSpace(value)), true
}

// unescapeString handles Java properties escape sequences
func (c *Codec) unescapeString(s string) string {
	var result strings.Builder
	escaped := false

	for i := 0; i < len(s); i++ {
		ch := s[i]

		if escaped {
			switch ch {
			case 't':
				result.WriteByte('\t')
			case 'n':
				result.WriteByte('\n')
			case 'r':
				result.WriteByte('\r')
			case 'f':
				result.WriteByte('\f')
			case '\\':
				result.WriteByte('\\')
			case '=':
				result.WriteByte('=')
			case ':':
				result.WriteByte(':')
			case ' ':
				result.WriteByte(' ')
			default:
				// Unknown escape, keep as-is
				result.WriteByte('\\')
				result.WriteByte(ch)
			}
			escaped = false
		} else if ch == '\\' {
			escaped = true
		} else {
			result.WriteByte(ch)
		}
	}

	return result.String()
}

// escapeKey escapes special characters in property keys
func (c *Codec) escapeKey(s string) string {
	return c.escapeCommon(s, true)
}

// escapeValue escapes special characters in property values
func (c *Codec) escapeValue(s string) string {
	return c.escapeCommon(s, false)
}

// escapeCommon handles common escape sequences
func (c *Codec) escapeCommon(s string, isKey bool) string {
	var result strings.Builder

	for _, ch := range s {
		switch ch {
		case '\\':
			result.WriteString(`\\`)
		case '\t':
			result.WriteString(`\t`)
		case '\n':
			result.WriteString(`\n`)
		case '\r':
			result.WriteString(`\r`)
		case '\f':
			result.WriteString(`\f`)
		case '=':
			if isKey {
				result.WriteString(`\=`)
			} else {
				result.WriteRune(ch)
			}
		case ':':
			if isKey {
				result.WriteString(`\:`)
			} else {
				result.WriteRune(ch)
			}
		case ' ':
			if isKey {
				result.WriteString(`\ `)
			} else {
				result.WriteRune(ch)
			}
		default:
			result.WriteRune(ch)
		}
	}

	return result.String()
}
