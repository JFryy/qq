package codec

import (
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

func PrettyFormat(s string, fileType EncodingType, raw bool, monochrome bool) (string, error) {
	if raw {
		var v any
		err := Unmarshal([]byte(s), fileType, &v)
		if err != nil {
			return "", err
		}
		switch val := v.(type) {
		case map[string]any:
			break
		case []any:
			break
		case string:
			// For strings, return directly (escapes already decoded)
			return val, nil
		default:
			// For numbers, booleans, null - convert to string representation
			return fmt.Sprintf("%v", val), nil
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

	return colorizeTokens(iterator), nil
}

// tokenColorMap maps chroma token types to ANSI color attributes.
// This uses semantic ANSI colors that respect the terminal's color theme.
var tokenColorMap = map[chroma.TokenType]color.Attribute{
	// Strings
	chroma.String:         color.FgGreen,
	chroma.StringDouble:   color.FgGreen,
	chroma.StringSingle:   color.FgGreen,
	chroma.StringBacktick: color.FgGreen,
	chroma.StringChar:     color.FgGreen,
	chroma.StringHeredoc:  color.FgGreen,
	chroma.StringInterpol: color.FgGreen,
	chroma.StringOther:    color.FgGreen,

	// Numbers
	chroma.Number:        color.FgYellow,
	chroma.NumberFloat:   color.FgYellow,
	chroma.NumberInteger: color.FgYellow,
	chroma.NumberHex:     color.FgYellow,
	chroma.NumberOct:     color.FgYellow,
	chroma.NumberBin:     color.FgYellow,

	// Keywords
	chroma.Keyword:            color.FgBlue,
	chroma.KeywordConstant:    color.FgBlue,
	chroma.KeywordDeclaration: color.FgBlue,
	chroma.KeywordNamespace:   color.FgBlue,
	chroma.KeywordPseudo:      color.FgBlue,
	chroma.KeywordReserved:    color.FgBlue,
	chroma.KeywordType:        color.FgBlue,

	// Names (keys, identifiers)
	chroma.Name:          color.FgCyan,
	chroma.NameAttribute: color.FgCyan,
	chroma.NameBuiltin:   color.FgCyan,
	chroma.NameClass:     color.FgCyan,
	chroma.NameConstant:  color.FgCyan,
	chroma.NameDecorator: color.FgCyan,
	chroma.NameEntity:    color.FgCyan,
	chroma.NameException: color.FgCyan,
	chroma.NameFunction:  color.FgCyan,
	chroma.NameLabel:     color.FgCyan,
	chroma.NameNamespace: color.FgCyan,
	chroma.NameOther:     color.FgCyan,
	chroma.NameTag:       color.FgCyan,
	chroma.NameVariable:  color.FgCyan,

	// Booleans
	chroma.LiteralStringBoolean: color.FgMagenta,

	// Comments
	chroma.Comment:          color.FgHiBlack,
	chroma.CommentHashbang:  color.FgHiBlack,
	chroma.CommentMultiline: color.FgHiBlack,
	chroma.CommentSingle:    color.FgHiBlack,
	chroma.CommentSpecial:   color.FgHiBlack,
	chroma.CommentPreproc:   color.FgHiBlack,

	// Operators
	chroma.Operator:     color.FgMagenta,
	chroma.OperatorWord: color.FgMagenta,

	// Punctuation
	chroma.Punctuation: color.FgWhite,

	// Errors
	chroma.Error: color.FgRed,
}

// colorizeTokens applies semantic ANSI colors to tokens based on their type.
// This uses the terminal's color theme instead of hardcoded RGB values.
func colorizeTokens(iterator chroma.Iterator) string {
	var result strings.Builder

	for token := iterator(); token != chroma.EOF; token = iterator() {
		value := token.Value
		if value == "" {
			continue
		}

		// Look up color for this token type
		if colorAttr, ok := tokenColorMap[token.Type]; ok {
			result.WriteString(color.New(colorAttr).Sprint(value))
		} else {
			// For any unmapped token types, output as-is
			result.WriteString(value)
		}
	}

	return result.String()
}
