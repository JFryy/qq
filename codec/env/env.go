package env

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Codec handles environment file parsing and marshaling
type Codec struct{}

// EnvVar represents a single environment variable with metadata
type EnvVar struct {
	Value    interface{} `json:"value"`
	Type     string      `json:"type"`
	Comment  string      `json:"comment,omitempty"`
	Exported bool        `json:"exported"`
}

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

	var envVars map[string]interface{}
	if err := json.Unmarshal(data, &envVars); err != nil {
		return nil, err
	}

	var lines []string
	for key, value := range envVars {
		// Handle both simple values and EnvVar structs
		switch val := value.(type) {
		case map[string]interface{}:
			// This is an EnvVar struct
			if v, exists := val["value"]; exists {
				line := fmt.Sprintf("%s=%s", key, c.formatValue(v))
				if comment, hasComment := val["comment"].(string); hasComment && comment != "" {
					line += " # " + comment
				}
				lines = append(lines, line)
			}
		default:
			// Simple key=value
			lines = append(lines, fmt.Sprintf("%s=%s", key, c.formatValue(val)))
		}
	}

	return []byte(strings.Join(lines, "\n")), nil
}

// Parse processes environment file content into structured data
func (c *Codec) Parse(content string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	lines := strings.Split(content, "\n")

	// Patterns for parsing
	varPattern := regexp.MustCompile(`^(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)=(.*)$`)
	commentPattern := regexp.MustCompile(`^#\s*(.*)$`)

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Handle comments
		if commentPattern.MatchString(line) {
			continue // Skip standalone comments for now
		}

		// Parse variable assignment
		if matches := varPattern.FindStringSubmatch(line); matches != nil {
			key := matches[1]
			valueWithComment := matches[2]

			// Check if exported
			exported := strings.HasPrefix(strings.TrimSpace(lines[lineNum]), "export")

			// Split value and inline comment
			value, comment := c.parseValueAndComment(valueWithComment)

			// Parse and type the value
			parsedValue, valueType := c.parseValue(value)

			result[key] = &EnvVar{
				Value:    parsedValue,
				Type:     valueType,
				Comment:  comment,
				Exported: exported,
			}
		}
	}

	return result, nil
}

// parseValueAndComment separates value from inline comments
func (c *Codec) parseValueAndComment(input string) (string, string) {
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
				// Found end quote
				value := input[:i+1]
				remainder := strings.TrimSpace(input[i+1:])
				if strings.HasPrefix(remainder, "#") {
					return value, strings.TrimSpace(remainder[1:])
				}
				return value, ""
			}
		}
		// No closing quote found, treat as is
		return input, ""
	}

	if strings.HasPrefix(input, "'") {
		// Similar logic for single quotes
		for i := 1; i < len(input); i++ {
			if input[i] == '\'' {
				value := input[:i+1]
				remainder := strings.TrimSpace(input[i+1:])
				if strings.HasPrefix(remainder, "#") {
					return value, strings.TrimSpace(remainder[1:])
				}
				return value, ""
			}
		}
		return input, ""
	}

	// Unquoted value - look for comment
	if idx := strings.Index(input, "#"); idx != -1 {
		value := strings.TrimSpace(input[:idx])
		comment := strings.TrimSpace(input[idx+1:])
		return value, comment
	}

	return strings.TrimSpace(input), ""
}

// parseValue attempts to parse a value into appropriate Go type
func (c *Codec) parseValue(value string) (interface{}, string) {
	value = strings.TrimSpace(value)

	// Handle quoted strings
	if (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) ||
		(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
		// Remove quotes and handle escape sequences for double quotes
		unquoted := value[1 : len(value)-1]
		if strings.HasPrefix(value, `"`) {
			unquoted = strings.ReplaceAll(unquoted, `\"`, `"`)
			unquoted = strings.ReplaceAll(unquoted, `\\`, `\`)
			unquoted = strings.ReplaceAll(unquoted, `\n`, "\n")
			unquoted = strings.ReplaceAll(unquoted, `\t`, "\t")
		}
		return unquoted, "string"
	}

	// Try to parse as integer
	if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
		return intVal, "integer"
	}

	// Try to parse as float
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal, "number"
	}

	// Try to parse as boolean
	switch strings.ToLower(value) {
	case "true", "yes", "1", "on":
		return true, "boolean"
	case "false", "no", "0", "off":
		return false, "boolean"
	}

	// Default to string
	return value, "string"
}

// formatValue converts a value back to env file format
func (c *Codec) formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		// Quote if contains spaces or special characters
		if strings.ContainsAny(v, " \t\n#=") || v == "" {
			escaped := strings.ReplaceAll(v, `\`, `\\`)
			escaped = strings.ReplaceAll(escaped, `"`, `\"`)
			escaped = strings.ReplaceAll(escaped, "\n", `\n`)
			escaped = strings.ReplaceAll(escaped, "\t", `\t`)
			return fmt.Sprintf(`"%s"`, escaped)
		}
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int, int32, int64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
