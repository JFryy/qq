package codec

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"strings"
)

// EncodingType represents the supported encoding types as an enum with a string representation
type EncodingType int

const (
	JSON EncodingType = iota
	YAML
	YML
	TOML
	HCL
	TF
	CSV
	XML
	INI
	GRON
	HTML
	LINE
	TXT
    MD
)

func (e EncodingType) String() string {
	return [...]string{"json", "yaml", "yml", "toml", "hcl", "tf", "csv", "xml", "ini", "gron", "html", "line", "txt", "md"}[e]
}

type Encoding struct {
	Ext       EncodingType
	Unmarshal func([]byte, interface{}) error
	Marshal   func(interface{}) ([]byte, error)
}

func jsonMarshalIndent(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

func GetEncodingType(fileType string) (EncodingType, error) {
	fileType = strings.ToLower(fileType)
	for _, t := range SupportedFileTypes {
		if fileType == t.Ext.String() {
			return t.Ext, nil
		}
	}
	return JSON, fmt.Errorf("unsupported file type: %v", fileType)
}

var SupportedFileTypes = []Encoding{
	{JSON, json.Unmarshal, jsonMarshalIndent},
	{YAML, yaml.Unmarshal, yaml.Marshal},
	{YML, yaml.Unmarshal, yaml.Marshal},
	{TOML, toml.Unmarshal, tomlMarshal},
	{HCL, hclUnmarshal, hclMarshal},
	{TF, hclUnmarshal, hclMarshal},
	{CSV, csvUnmarshal, jsonMarshalIndent},
	{XML, xmlUnmarshal, xmlMarshal},
	{INI, iniUnmarshal, iniMarshal},
	{GRON, gronUnmarshal, gronMarshal},
	{HTML, htmlUnmarshal, jsonMarshalIndent},
	{LINE, lineUnmarshal, jsonMarshalIndent},
	{TXT, lineUnmarshal, jsonMarshalIndent},
    {MD, markdownUnmarshal, jsonMarshalIndent},
}

func Unmarshal(input []byte, inputFileType EncodingType, data interface{}) error {
	for _, t := range SupportedFileTypes {
		if t.Ext == inputFileType {
			err := t.Unmarshal(input, data)
			if err != nil {
				return fmt.Errorf("error parsing input: %v", err)
			}
			return nil
		}
	}
	return fmt.Errorf("unsupported input file type: %v", inputFileType)
}

func Marshal(v interface{}, outputFileType EncodingType) ([]byte, error) {
	for _, t := range SupportedFileTypes {
		if t.Ext == outputFileType {
			var err error
			b, err := t.Marshal(v)
			if err != nil {
				return b, fmt.Errorf("error marshaling result to %s: %v", outputFileType, err)
			}
			return b, nil
		}
	}
	return nil, fmt.Errorf("unsupported output file type: %v", outputFileType)
}
