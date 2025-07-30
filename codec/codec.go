package codec

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/goccy/go-json"
	"github.com/mattn/go-isatty"
	"os"
	"strings"
	// dedicated codec packages and wrappers where appropriate
	"github.com/JFryy/qq/codec/csv"
	"github.com/JFryy/qq/codec/env"
	"github.com/JFryy/qq/codec/gron"
	"github.com/JFryy/qq/codec/hcl"
	"github.com/JFryy/qq/codec/html"
	"github.com/JFryy/qq/codec/ini"
	qqjson "github.com/JFryy/qq/codec/json"
	"github.com/JFryy/qq/codec/line"
	"github.com/JFryy/qq/codec/msgpack"
	"github.com/JFryy/qq/codec/parquet"
	proto "github.com/JFryy/qq/codec/proto"
	"github.com/JFryy/qq/codec/xml"
	"github.com/JFryy/qq/codec/yaml"
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
	PROTO
	ENV
	PARQUET
	MSGPACK
	MPK
)

func (e EncodingType) String() string {
	return [...]string{"json", "yaml", "yml", "toml", "hcl", "tf", "csv", "xml", "ini", "gron", "html", "line", "txt", "proto", "env", "parquet", "msgpack", "mpk"}[e]
}

type Encoding struct {
	Ext       EncodingType
	Unmarshal func([]byte, any) error
	Marshal   func(any) ([]byte, error)
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

var (
	htm          = html.Codec{}
	jsn          = qqjson.Codec{} // wrapper for go-json marshal
	grn          = gron.Codec{}
	hcltf        = hcl.Codec{}
	xmll         = xml.Codec{}
	inii         = ini.Codec{}
	lines        = line.Codec{}
	sv           = csv.Codec{}
	pb           = proto.Codec{}
	yml          = yaml.Codec{}
	envCodec     = env.Codec{}
	parquetCodec = parquet.Codec{}
	msgpackCodec = msgpack.Codec{}
)
var SupportedFileTypes = []Encoding{
	{JSON, json.Unmarshal, jsn.Marshal},
	{YAML, yml.Unmarshal, yml.Marshal},
	{YML, yml.Unmarshal, yml.Marshal},
	{TOML, toml.Unmarshal, toml.Marshal},
	{HCL, hcltf.Unmarshal, hcltf.Marshal},
	{TF, hcltf.Unmarshal, hcltf.Marshal},
	{CSV, sv.Unmarshal, sv.Marshal},
	{XML, xmll.Unmarshal, xmll.Marshal},
	{INI, inii.Unmarshal, inii.Marshal},
	{GRON, grn.Unmarshal, grn.Marshal},
	{HTML, htm.Unmarshal, xmll.Marshal},
	{LINE, lines.Unmarshal, jsn.Marshal},
	{TXT, lines.Unmarshal, jsn.Marshal},
	{PROTO, pb.Unmarshal, jsn.Marshal},
	{ENV, envCodec.Unmarshal, envCodec.Marshal},
	{PARQUET, parquetCodec.Unmarshal, parquetCodec.Marshal},
	{MSGPACK, msgpackCodec.Unmarshal, msgpackCodec.Marshal},
	{MPK, msgpackCodec.Unmarshal, msgpackCodec.Marshal},
}

func Unmarshal(input []byte, inputFileType EncodingType, data any) error {
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

func Marshal(v any, outputFileType EncodingType) ([]byte, error) {
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

func PrettyFormat(s string, fileType EncodingType, raw bool, monochrome bool) (string, error) {
	if raw {
		var v any
		err := Unmarshal([]byte(s), fileType, &v)
		if err != nil {
			return "", err
		}
		switch v.(type) {
		case map[string]any:
			break
		case []any:
			break
		default:
			return strings.ReplaceAll(s, "\"", ""), nil
		}
	}

	if !isatty.IsTerminal(os.Stdout.Fd()) || monochrome {
		return s, nil
	}

	var lexer chroma.Lexer
	// this a workaround for json lexer while we don't have a marshal function dedicated for these formats.
	if fileType == CSV || fileType == HTML || fileType == LINE || fileType == TXT || fileType == ENV || fileType == PARQUET || fileType == MSGPACK || fileType == MPK {
		lexer = lexers.Get("json")
	} else {
		lexer = lexers.Get(fileType.String())
		if lexer == nil {
			lexer = lexers.Fallback
		}
	}

	if lexer == nil {
		return "", fmt.Errorf("unsupported file type for formatting: %v", fileType)
	}

	iterator, err := lexer.Tokenise(nil, s)
	if err != nil {
		return "", fmt.Errorf("error tokenizing input: %v", err)
	}

	style := styles.Get("nord")
	formatter := formatters.Get("terminal256")
	var buffer bytes.Buffer

	err = formatter.Format(&buffer, style, iterator)
	if err != nil {
		return "", fmt.Errorf("error formatting output: %v", err)
	}

	return buffer.String(), nil
}

func IsBinaryFormat(fileType EncodingType) bool {
	return fileType == PARQUET || fileType == MSGPACK || fileType == MPK
}
