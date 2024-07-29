package codec

import (
	"fmt"
	"github.com/goccy/go-json"
	"strings"
)

func lineUnmarshal(input []byte, v interface{}) error {
	lines := strings.Split(strings.TrimSpace(string(input)), "\n")

	// Marshal the lines to JSON and then unmarshal into the provided interface
	jsonData, err := json.Marshal(lines)
	if err != nil {
		return fmt.Errorf("error marshaling to JSON: %v", err)
	}

	if err := json.Unmarshal(jsonData, v); err != nil {
		return fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	return nil
}
