package codec

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/goccy/go-json"

	// dedicated codec packages and wrappers where appropriate
	"github.com/JFryy/qq/codec/base64"
	"github.com/JFryy/qq/codec/csv"
	"github.com/JFryy/qq/codec/env"
	"github.com/JFryy/qq/codec/gron"
	"github.com/JFryy/qq/codec/hcl"
	"github.com/JFryy/qq/codec/html"
	"github.com/JFryy/qq/codec/ini"
	qqjson "github.com/JFryy/qq/codec/json"
	"github.com/JFryy/qq/codec/jsonc"
	"github.com/JFryy/qq/codec/jsonl"
	"github.com/JFryy/qq/codec/line"
	"github.com/JFryy/qq/codec/msgpack"
	"github.com/JFryy/qq/codec/parquet"
	"github.com/JFryy/qq/codec/properties"
	proto "github.com/JFryy/qq/codec/proto"
	"github.com/JFryy/qq/codec/tsv"
	"github.com/JFryy/qq/codec/xml"
	"github.com/JFryy/qq/codec/yaml"
)

// EncodingType represents the supported encoding types as an enum with a string representation
type EncodingType int

const (
	JSON EncodingType = iota
	YAML
	TOML
	HCL
	CSV
	TSV
	XML
	INI
	GRON
	HTML
	LINE
	TXT
	PROTO
	ENV
	PARQUET
	MSGPACK
	PROPERTIES
	JSONL
	JSONC
	BASE64
)

// String implements the Stringer interface, converting the enum to its canonical string name.
// This is used for:
//   - Human-readable error messages and logs (e.g., "json" instead of "0")
//   - Lexer lookup for syntax highlighting
//   - Test output and debugging
//
// Note: These names are duplicated in the Extensions field of Codecs map. This duplication
// is intentional for performance - O(1) array lookup here vs O(n) map iteration.
// The array indices must match the iota order in the const block above.
func (e EncodingType) String() string {
	return [...]string{"json", "yaml", "toml", "hcl", "csv", "tsv", "xml", "ini", "gron", "html", "line", "txt", "proto", "env", "parquet", "msgpack", "properties", "jsonl", "jsonc", "base64"}[e]
}

// General Encoding struct to hold unmarshal/marshal functions and associated file extensions for each encoding type
// This allows for a clean separation of concerns and makes it easy to add new encodings in the future by simply implementing the Codec interface and adding an entry to the Codecs map.
// The Extensions field is used for file type detection and mapping to the appropriate encoding type.
type Encoding struct {
	Unmarshal  func([]byte, any) error
	Marshal    func(any) ([]byte, error)
	Extensions []string
}

func GetEncodingType(fileType string) (EncodingType, error) {
	fileType = strings.ToLower(fileType)
	for encType, enc := range Codecs {
		for _, ext := range enc.Extensions {
			if fileType == ext {
				return encType, nil
			}
		}
	}
	return JSON, fmt.Errorf("unsupported file type: %v", fileType)
}

// GetExtensionMap returns a map of file extensions to their encoding types
func GetExtensionMap() map[string]EncodingType {
	extMap := make(map[string]EncodingType)
	for encType, enc := range Codecs {
		for _, ext := range enc.Extensions {
			// Only set if not already set (earlier entries win)
			if _, exists := extMap[ext]; !exists {
				extMap[ext] = encType
			}
		}
	}
	return extMap
}

// GetSupportedExtensions returns a sorted list of unique supported extensions
func GetSupportedExtensions() []string {
	seen := make(map[string]bool)
	var exts []string
	for _, enc := range Codecs {
		for _, ext := range enc.Extensions {
			if !seen[ext] {
				seen[ext] = true
				exts = append(exts, ext)
			}
		}
	}
	return exts
}

var (
	htmlCodec       = html.Codec{}
	jsonCodec       = qqjson.Codec{} // wrapper for go-json marshal
	gronCodec       = gron.Codec{}
	hclCodec        = hcl.Codec{}
	xmlCodec        = xml.Codec{}
	iniCodec        = ini.Codec{}
	lineCodec       = line.Codec{}
	csvCodec        = csv.Codec{}
	tsvCodec        = tsv.Codec{}
	protoCodec      = proto.Codec{}
	yamlCodec       = yaml.Codec{}
	envCodec        = env.Codec{}
	parquetCodec    = parquet.Codec{}
	msgpackCodec    = msgpack.Codec{}
	propertiesCodec = properties.Codec{}
	jsonlCodec      = jsonl.Codec{}
	jsoncCodec      = jsonc.Codec{}
	base64Codec     = base64.Codec{}
)

var Codecs = map[EncodingType]Encoding{
	JSON:       {json.Unmarshal, jsonCodec.Marshal, []string{"json"}},
	YAML:       {yamlCodec.Unmarshal, yamlCodec.Marshal, []string{"yaml", "yml"}},
	TOML:       {toml.Unmarshal, toml.Marshal, []string{"toml"}},
	HCL:        {hclCodec.Unmarshal, hclCodec.Marshal, []string{"hcl", "tf"}},
	CSV:        {csvCodec.Unmarshal, csvCodec.Marshal, []string{"csv"}},
	TSV:        {tsvCodec.Unmarshal, tsvCodec.Marshal, []string{"tsv"}},
	XML:        {xmlCodec.Unmarshal, xmlCodec.Marshal, []string{"xml"}},
	INI:        {iniCodec.Unmarshal, iniCodec.Marshal, []string{"ini"}},
	GRON:       {gronCodec.Unmarshal, gronCodec.Marshal, []string{"gron"}},
	HTML:       {htmlCodec.Unmarshal, xmlCodec.Marshal, []string{"html"}},
	LINE:       {lineCodec.Unmarshal, jsonCodec.Marshal, []string{"line"}},
	TXT:        {lineCodec.Unmarshal, jsonCodec.Marshal, []string{"txt", "text"}},
	PROTO:      {protoCodec.Unmarshal, jsonCodec.Marshal, []string{"proto"}},
	ENV:        {envCodec.Unmarshal, envCodec.Marshal, []string{"env"}},
	PARQUET:    {parquetCodec.Unmarshal, parquetCodec.Marshal, []string{"parquet"}},
	MSGPACK:    {msgpackCodec.Unmarshal, msgpackCodec.Marshal, []string{"msgpack", "mpk"}},
	PROPERTIES: {propertiesCodec.Unmarshal, propertiesCodec.Marshal, []string{"properties"}},
	JSONL:      {jsonlCodec.Unmarshal, jsonlCodec.Marshal, []string{"jsonl", "ndjson", "jsonlines"}},
	JSONC:      {jsoncCodec.Unmarshal, jsoncCodec.Marshal, []string{"jsonc"}},
	BASE64:     {base64Codec.Unmarshal, base64Codec.Marshal, []string{"base64", "b64"}},
}

func Unmarshal(input []byte, inputFileType EncodingType, data any) error {
	if data == nil {
		return fmt.Errorf("data parameter cannot be nil")
	}
	codec, ok := Codecs[inputFileType]
	if !ok {
		return fmt.Errorf("unsupported input file type: %v", inputFileType)
	}
	if err := codec.Unmarshal(input, data); err != nil {
		return fmt.Errorf("error parsing input: %v", err)
	}
	return nil
}

func Marshal(v any, outputFileType EncodingType) ([]byte, error) {
	if v == nil {
		return nil, fmt.Errorf("input data cannot be nil")
	}
	codec, ok := Codecs[outputFileType]
	if !ok {
		return nil, fmt.Errorf("unsupported output file type: %v", outputFileType)
	}
	data, err := codec.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("error marshaling result to %s: %v", outputFileType, err)
	}
	return data, nil
}

func IsBinaryFormat(fileType EncodingType) bool {
	return fileType == PARQUET || fileType == MSGPACK
}
