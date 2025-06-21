package yaml

import (
	"github.com/goccy/go-yaml"
	"strings"
)

type Codec struct{}

// Unmarshal strips document delimiter and unmarshals YAML
func (c Codec) Unmarshal(data []byte, v any) error {
	// Strip document delimiter if present at the beginning
	content := strings.TrimSpace(string(data))
	if strings.HasPrefix(content, "---") {
		lines := strings.SplitN(content, "\n", 2)
		if len(lines) > 1 {
			content = lines[1]
		} else {
			content = ""
		}
	}
	return yaml.Unmarshal([]byte(content), v)
}

func (c Codec) Marshal(v any) ([]byte, error) {
	return yaml.Marshal(v)
}
