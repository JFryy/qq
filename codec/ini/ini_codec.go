package ini

import (
	"fmt"
	"github.com/JFryy/qq/codec/util"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/ini.v1"
	"strings"
)

type Codec struct{}

func (c *Codec) Unmarshal(input []byte, v interface{}) error {
	cfg, err := ini.Load(input)
	if err != nil {
		return fmt.Errorf("error unmarshaling INI: %v", err)
	}

	data := make(map[string]interface{})
	for _, section := range cfg.Sections() {
		if section.Name() == ini.DefaultSection {
			continue
		}
		sectionMap := make(map[string]interface{})
		for _, key := range section.Keys() {
			sectionMap[key.Name()] = util.ParseValue(key.Value())
		}
		data[section.Name()] = sectionMap
	}

	return mapstructure.Decode(data, v)
}

func (c *Codec) Marshal(v interface{}) ([]byte, error) {
	data, ok := v.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("input data is not a map")
	}

	cfg := ini.Empty()
	for section, sectionValue := range data {
		sectionMap, ok := sectionValue.(map[string]interface{})
		if !ok {
			continue
		}

		sec, err := cfg.NewSection(section)
		if err != nil {
			return nil, err
		}

		for key, value := range sectionMap {
			_, err := sec.NewKey(key, fmt.Sprintf("%v", value))
			if err != nil {
				return nil, err
			}
		}
	}

	var b strings.Builder
	_, err := cfg.WriteTo(&b)
	if err != nil {
		return nil, fmt.Errorf("error writing INI data: %v", err)
	}
	return []byte(b.String()), nil
}
