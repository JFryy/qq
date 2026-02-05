package env

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/goccy/go-json"
)

// Codec handles environment file parsing and marshaling
type Codec struct{}

// Unmarshal parses environment file data into the provided interface
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

// Marshal converts data back to environment file format
func (c *Codec) Marshal(v interface{}) ([]byte, error) {
	// Convert to our expected format first
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var envVars map[string]string
	if err := json.Unmarshal(data, &envVars); err != nil {
		return nil, errors.New("env format only supports simple key-value pairs, cannot convert complex nested structures")
	}

	var lines []string
	for key, value := range envVars {
		lines = append(lines, fmt.Sprintf("%s=%s", key, c.formatValue(value)))
	}

	return []byte(strings.Join(lines, "\n")), nil
}

// Parse processes environment file content into simple key-value pairs
func (c *Codec) Parse(content string) (map[string]string, error) {
	result := make(map[string]string)
	lines := strings.Split(content, "\n")

	// Pattern for parsing variable assignments
	varPattern := regexp.MustCompile(`^(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)=(.*)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse variable assignment
		if matches := varPattern.FindStringSubmatch(line); matches != nil {
			key := matches[1]
			valueWithComment := matches[2]

			// Extract just the value, ignoring comments
			value := c.extractValue(valueWithComment)
			result[key] = value
		}
	}

	return result, nil
}

// extractValue extracts the value part from a line, handling quotes and ignoring comments
func (c *Codec) extractValue(input string) string {
	input = strings.TrimSpace(input)

	// Handle quoted strings first
	if strings.HasPrefix(input, `"`) {
		// Find the closing quote, handling escaped quotes
		escaped := false
		for i := 1; i < len(input); i++ {
			if escaped {
				escaped = false
				continue
			}
			if input[i] == '\\' {
				escaped = true
				continue
			}
			if input[i] == '"' {
				// Found end quote, remove outer quotes and handle escapes
				unquoted := input[1:i]
				// Handle specific escape sequences only
				// Use a more careful approach to avoid double-processing
				result := ""
				for j := 0; j < len(unquoted); j++ {
					if j < len(unquoted)-1 && unquoted[j] == '\\' {
						switch unquoted[j+1] {
						case '"':
							result += `"`
							j++ // skip next char
						case 'n':
							result += "\n"
							j++ // skip next char
						case 't':
							result += "\t"
							j++ // skip next char
						case '\\':
							result += `\`
							j++ // skip next char
						default:
							result += string(unquoted[j])
						}
					} else {
						result += string(unquoted[j])
					}
				}
				return result
			}
		}
		// No closing quote found, return as is without outer quotes
		return input[1:]
	}

	if strings.HasPrefix(input, "'") {
		// Similar logic for single quotes (no escape processing)
		for i := 1; i < len(input); i++ {
			if input[i] == '\'' {
				return input[1:i]
			}
		}
		return input[1:]
	}

	// Unquoted value - look for comment and strip it
	if idx := strings.Index(input, "#"); idx != -1 {
		return strings.TrimSpace(input[:idx])
	}

	return input
}

// formatValue converts a string value back to env file format
func (c *Codec) formatValue(value string) string {
	// Quote if contains spaces or special characters
	if strings.ContainsAny(value, " \t\n#=") || value == "" {
		escaped := strings.ReplaceAll(value, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		escaped = strings.ReplaceAll(escaped, "\n", `\n`)
		escaped = strings.ReplaceAll(escaped, "\t", `\t`)
		return fmt.Sprintf(`"%s"`, escaped)
	}
	return value
}
