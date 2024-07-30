package codec

import (
	"bytes"
	"fmt"
	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/mattn/go-isatty"
	"os"
	"strings"
    "strconv"
    "time"

)

func PrettyFormat(s string, fileType EncodingType, raw bool) (string, error) {
	if raw {
		var v interface{}
		err := Unmarshal([]byte(s), fileType, &v)
		if err != nil {
			return "", err
		}
		switch v.(type) {
		case map[string]interface{}:
			break
		case []interface{}:
			break
		default:
			return strings.ReplaceAll(s, "\"", ""), nil
		}
	}

	if !isatty.IsTerminal(os.Stdout.Fd()) {
		return s, nil
	}

	var lexer chroma.Lexer
	if fileType == CSV {
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

func parseValue(value string) interface{} {
	value = strings.TrimSpace(value)

	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}
	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		return floatValue
	}
	if boolValue, err := strconv.ParseBool(value); err == nil {
		return boolValue
	}
	if dateValue, err := time.Parse(time.RFC3339, value); err == nil {
		return dateValue
	}
	if dateValue, err := time.Parse("2006-01-02", value); err == nil {
		return dateValue
	}
	return value
}

