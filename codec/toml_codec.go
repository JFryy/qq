package codec

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"strings"
)

func tomlMarshal(v interface{}) ([]byte, error) {
	buf := new(strings.Builder)
	err := toml.NewEncoder(buf).Encode(v)
	if err != nil {
		return nil, fmt.Errorf("error marshaling result to TOML: %v", err)
	}
	return []byte(buf.String()), nil
}
