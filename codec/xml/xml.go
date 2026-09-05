package xml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
	"sort"

	"github.com/JFryy/qq/codec/util"
	"github.com/clbanning/mxj/v2"
)

type Codec struct{}

type xmlNode struct {
	name     string
	children []string
	nested   map[string]*xmlNode
}

func (c *Codec) Marshal(v any) ([]byte, error) {
	if !util.PreserveKeyOrder {
		switch v := v.(type) {
		case map[string]any:
			mv := mxj.Map(v)
			return mv.XmlIndent("", "  ")
		case []any:
			mv := mxj.Map(map[string]any{"root": v})
			return mv.XmlIndent("", "  ")
		default:
			mv := mxj.Map(map[string]any{"value": v})
			return mv.XmlIndent("", "  ")
		}
	}

	switch val := v.(type) {
	case map[string]any:
		if len(val) == 1 {
			for k, innerVal := range val {
				return marshalXMLOrdered(innerVal, k, "  ", "")
			}
		}
		return marshalXMLOrdered(val, "root", "  ", "")
	default:
		return marshalXMLOrdered(val, "root", "  ", "")
	}
}

func (c *Codec) Unmarshal(input []byte, v any) error {
	mv, err := mxj.NewMapXml(input)
	if err != nil {
		return fmt.Errorf("error unmarshaling XML: %v", err)
	}

	parsedData := c.parseXMLValues(mv.Old())

	if util.PreserveKeyOrder {
		structNode, err := parseXMLStructure(input)
		if err == nil {
			keys := make(map[uintptr][]string)
			matchAndRegister(parsedData, structNode, keys)
			for ptr, list := range keys {
				util.SetKeyOrder(ptr, list)
			}
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

func parseXMLStructure(data []byte) (*xmlNode, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	root := &xmlNode{name: "root", nested: make(map[string]*xmlNode)}
	stack := []*xmlNode{root}

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			parent := stack[len(stack)-1]
			tagName := t.Name.Local

			has := false
			for _, child := range parent.children {
				if child == tagName {
					has = true
					break
				}
			}
			if !has {
				parent.children = append(parent.children, tagName)
			}

			childNode := parent.nested[tagName]
			if childNode == nil {
				childNode = &xmlNode{name: tagName, nested: make(map[string]*xmlNode)}
				parent.nested[tagName] = childNode
			}
			stack = append(stack, childNode)

		case xml.EndElement:
			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
			}
		}
	}

	return root, nil
}

func matchAndRegister(val any, node *xmlNode, keys map[uintptr][]string) {
	switch v := val.(type) {
	case map[string]any:
		if len(node.children) > 0 {
			var orderedKeys []string
			for _, child := range node.children {
				if _, has := v[child]; has {
					orderedKeys = append(orderedKeys, child)
				}
			}
			ptr := uintptr(reflect.ValueOf(v).UnsafePointer())
			keys[ptr] = orderedKeys
		}

		for key, childVal := range v {
			if childNode, ok := node.nested[key]; ok {
				matchAndRegister(childVal, childNode, keys)
			}
		}

	default:
		rv := reflect.ValueOf(val)
		if rv.Kind() == reflect.Slice {
			for i := 0; i < rv.Len(); i++ {
				matchAndRegister(rv.Index(i).Interface(), node, keys)
			}
		}
	}
}

func marshalXMLOrdered(v any, tagName string, indent string, currentIndent string) ([]byte, error) {
	if v == nil {
		return []byte(fmt.Sprintf("%s<%s></%s>", currentIndent, tagName, tagName)), nil
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice {
		var buf bytes.Buffer
		for i := 0; i < rv.Len(); i++ {
			if i > 0 && indent != "" {
				buf.WriteByte('\n')
			}
			elemBytes, err := marshalXMLOrdered(rv.Index(i).Interface(), tagName, indent, currentIndent)
			if err != nil {
				return nil, err
			}
			buf.Write(elemBytes)
		}
		return buf.Bytes(), nil
	}

	switch val := v.(type) {
	case map[string]any:
		ptr := uintptr(reflect.ValueOf(val).UnsafePointer())
		orderedKeys, ok := util.GetKeyOrder(ptr)
		if !ok || len(orderedKeys) != len(val) {
			orderedKeys = make([]string, 0, len(val))
			for k := range val {
				orderedKeys = append(orderedKeys, k)
			}
			sort.Strings(orderedKeys)
		} else {
			for _, k := range orderedKeys {
				if _, has := val[k]; !has {
					orderedKeys = make([]string, 0, len(val))
					for k2 := range val {
						orderedKeys = append(orderedKeys, k2)
					}
					sort.Strings(orderedKeys)
					break
				}
			}
		}

		var buf bytes.Buffer
		fmt.Fprintf(&buf, "%s<%s>", currentIndent, tagName)
		nextIndent := currentIndent + indent
		for _, k := range orderedKeys {
			if indent != "" {
				buf.WriteByte('\n')
			}
			elemBytes, err := marshalXMLOrdered(val[k], k, indent, nextIndent)
			if err != nil {
				return nil, err
			}
			buf.Write(elemBytes)
		}
		if indent != "" {
			buf.WriteByte('\n')
			buf.WriteString(currentIndent)
		}
		fmt.Fprintf(&buf, "</%s>", tagName)
		return buf.Bytes(), nil

	default:
		text := fmt.Sprintf("%v", val)
		var buf bytes.Buffer
		if err := xml.EscapeText(&buf, []byte(text)); err != nil {
			return nil, err
		}
		return []byte(fmt.Sprintf("%s<%s>%s</%s>", currentIndent, tagName, buf.String(), tagName)), nil
	}
}
