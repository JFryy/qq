package xml

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/JFryy/qq/codec/keyorder"
	"github.com/JFryy/qq/codec/util"
	"github.com/clbanning/mxj/v2"
)

type Codec struct{}

const (
	indent = "  "
	// attrPrefix and textKey are mxj's default names for attributes and for
	// the character data of an element that also has attributes or children.
	attrPrefix = "-"
	textKey    = "#text"
)

func (c *Codec) Marshal(v any) ([]byte, error) {
	var m map[string]any
	switch v := v.(type) {
	case map[string]any:
		m = v
	case []any:
		m = map[string]any{"root": v}
	default:
		m = map[string]any{"value": v}
	}
	if keyorder.Enabled() {
		return marshalOrdered(m)
	}
	return mxj.Map(m).XmlIndent("", indent)
}

func (c *Codec) Unmarshal(input []byte, v any) error {
	mv, err := mxj.NewMapXml(input)
	if err != nil {
		return fmt.Errorf("error unmarshaling XML: %v", err)
	}

	parsedData := c.parseXMLValues(mv.Old())
	if keyorder.Recording() {
		if err := recordKeyOrder(input, parsedData); err != nil {
			return fmt.Errorf("error recording XML key order: %v", err)
		}
	}

	// reflection of values required for type assertions on interface
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("provided value must be a non-nil pointer")
	}
	rv.Elem().Set(reflect.ValueOf(parsedData))

	return nil
}

// infer the type of the value and parse it accordingly
func (c *Codec) parseXMLValues(v any) any {
	switch v := v.(type) {
	case map[string]any:
		for key, val := range v {
			v[key] = c.parseXMLValues(val)
		}
		return v
	case []any:
		for i, val := range v {
			v[i] = c.parseXMLValues(val)
		}
		return v
	case string:
		return util.ParseValue(v)
	default:
		return v
	}
}

// recordKeyOrder scans the XML document that was decoded into value and
// records, for each element map, its attributes in document order followed by
// its text and child elements in order of first appearance. Element and
// attribute names follow mxj: the local name, with attrPrefix for attributes.
func recordKeyOrder(input []byte, value any) error {
	decoder := xml.NewDecoder(bytes.NewReader(input))
	decoder.CharsetReader = mxj.XmlCharsetReader
	root, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if start, ok := token.(xml.StartElement); ok {
			return scanElement(decoder, start, root[start.Name.Local])
		}
	}
}

// scanElement reads the content of the element opened by start and records
// its key order on value, the element's decoded map.
func scanElement(decoder *xml.Decoder, start xml.StartElement, value any) error {
	m, _ := value.(map[string]any)
	var keys []string
	for _, attr := range start.Attr {
		keys = append(keys, attrPrefix+attr.Name.Local)
	}
	seen := make(map[string]int)
	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		switch t := token.(type) {
		case xml.StartElement:
			name := t.Name.Local
			keys = append(keys, name)
			if err := scanElement(decoder, t, occurrence(m[name], seen[name])); err != nil {
				return err
			}
			seen[name]++
		case xml.CharData:
			if strings.TrimSpace(string(t)) != "" {
				keys = append(keys, textKey)
			}
		case xml.EndElement:
			keyorder.Record(m, keys)
			return nil
		}
	}
}

// occurrence returns the value decoded for the i-th child element with a given
// name. mxj collects repeated child elements into a list.
func occurrence(value any, i int) any {
	if list, ok := value.([]any); ok {
		if i < len(list) {
			return list[i]
		}
		return nil
	}
	if i == 0 {
		return value
	}
	return nil
}

// pretty tracks indentation like mxj's encoder: padding is the current line
// prefix and depth is the number of indents applied.
type pretty struct {
	padding string
	depth   int
}

func (p pretty) indented() pretty {
	return pretty{padding: p.padding + indent, depth: p.depth + 1}
}

// marshalOrdered produces the same output as mxj.Map.XmlIndent("", indent),
// except that attributes and child elements are written in recorded key
// order. Values other than maps and arrays are encoded by mxj itself.
func marshalOrdered(m map[string]any) ([]byte, error) {
	var b bytes.Buffer
	var err error
	if len(m) == 1 {
		for key, value := range m {
			if _, ok := value.([]any); ok {
				err = writeElement(&b, mxj.DefaultRootTag, m, pretty{})
			} else {
				err = writeElement(&b, key, value, pretty{})
			}
		}
	} else {
		err = writeElement(&b, mxj.DefaultRootTag, m, pretty{})
	}
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func writeElement(b *bytes.Buffer, key string, value any, p pretty) error {
	switch v := value.(type) {
	case map[string]any:
		return writeMap(b, key, v, p)
	case []any:
		return writeList(b, key, v, p)
	}

	out, err := mxj.Map{key: value}.XmlIndent(p.padding, indent)
	if err != nil {
		return err
	}
	b.Write(out)
	// mxj ends every nested element except a non-empty []string list with a
	// newline; a top-level call has no nesting, so add it here.
	if list, ok := value.([]string); !ok || len(list) == 0 {
		endLine(b, p)
	}
	return nil
}

func writeMap(b *bytes.Buffer, key string, m map[string]any, p pretty) error {
	b.WriteString(p.padding + "<" + key)

	var attrs, children []string
	for _, k := range keyorder.Keys(m) {
		switch {
		case isAttr(k):
			attrs = append(attrs, k)
		case k != textKey:
			children = append(children, k)
		}
	}
	for _, k := range attrs {
		value, err := attrValue(k, m[k])
		if err != nil {
			return err
		}
		b.WriteString(" " + k[len(attrPrefix):] + `="` + value + `"`)
	}

	if len(attrs) == len(m) {
		b.WriteString("/>")
		endLine(b, p)
		return nil
	}

	text, hasText := m[textKey]
	if hasText && len(children) == 0 {
		b.WriteString(">" + textValue(text) + "</" + key + ">")
		endLine(b, p)
		return nil
	}

	b.WriteString(">")
	if hasText {
		b.WriteString(textValue(text))
	}
	b.WriteString("\n")
	for _, k := range children {
		child := m[k]
		childPretty := p.indented()
		// Lists indent each of their elements themselves.
		if _, ok := child.([]any); ok {
			childPretty = p
		}
		if err := writeElement(b, k, child, childPretty); err != nil {
			return err
		}
	}
	b.WriteString(p.padding + "</" + key + ">")
	endLine(b, p)
	return nil
}

func writeList(b *bytes.Buffer, key string, list []any, p pretty) error {
	if len(list) == 0 {
		// Matches mxj's output for an empty list.
		b.WriteString(p.padding + indent + "<" + key + p.padding + "/>")
		endLine(b, p)
		return nil
	}
	for _, item := range list {
		if err := writeElement(b, key, item, p.indented()); err != nil {
			return err
		}
	}
	return nil
}

func endLine(b *bytes.Buffer, p pretty) {
	if p.depth > 0 {
		b.WriteString("\n")
	}
}

func isAttr(key string) bool {
	return len(key) > len(attrPrefix) && strings.HasPrefix(key, attrPrefix)
}

// attrValue formats an attribute value as mxj does with its default of no
// character escaping. Only scalar attribute values are accepted.
func attrValue(key string, value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case float64, bool, int, int32, int64, float32, json.Number:
		return fmt.Sprintf("%v", v), nil
	default:
		return "", fmt.Errorf("invalid attribute value for: %s:<%T>", key, value)
	}
}

// textValue formats element text as mxj does with its default of no character
// escaping.
func textValue(value any) string {
	if v, ok := value.([]byte); ok {
		return string(v)
	}
	return fmt.Sprintf("%v", value)
}
