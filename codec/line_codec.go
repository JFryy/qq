package codec

import (
	"fmt"
	"github.com/goccy/go-json"
	"reflect"
	"strings"
)

func lineUnmarshal(input []byte, v interface{}) error {
	lines := strings.Split(strings.TrimSpace(string(input)), "\n")
	var parsedLines []interface{}

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		parsedValue := parseValue(trimmedLine)
		parsedLines = append(parsedLines, parsedValue)
	}

	jsonData, err := json.Marshal(parsedLines)
	if err != nil {
		return fmt.Errorf("error marshaling to JSON: %v", err)
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("provided value must be a non-nil pointer")
	}

	if err := json.Unmarshal(jsonData, rv.Interface()); err != nil {
		return fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	return nil
}
