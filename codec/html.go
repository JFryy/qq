package codec

import (
	"bytes"
	"github.com/goccy/go-json"
	"golang.org/x/net/html"
	"strings"
)

/* in progress code for html conversions
   has a variety of issues at the moment for unmarshaling,
   marshaling is not implemented yet and may not be implemented
   e.g. decoding only to other formats, not encodiing to html supported
   since the specification of input being formatted is difficult to achieve
   on the fly with jq.
*/

func htmlUnmarshal(data []byte, v interface{}) error {
	htmlMap, err := HTMLToMap(data)
	if err != nil {
		return err
	}
	b, err := json.Marshal(htmlMap) // To use JSON unmarshal into the interface
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func htmlMarshal(v interface{}) ([]byte, error) {
	var htmlMap map[string]interface{}
	b, err := json.Marshal(v) // To use JSON marshal from the interface
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &htmlMap)
	if err != nil {
		return nil, err
	}
	return MapToHTML(htmlMap)
}

func HTMLToMap(htmlBytes []byte) (map[string]interface{}, error) {
	doc, err := html.Parse(bytes.NewReader(htmlBytes))
	if err != nil {
		return nil, err
	}
	return nodeToMap(doc), nil
}

func MapToHTML(data map[string]interface{}) ([]byte, error) {
	node := mapToNode(data)
	var buf bytes.Buffer
	err := html.Render(&buf, node)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func nodeToMap(node *html.Node) map[string]interface{} {
	if node.Type == html.TextNode {
		text := strings.TrimSpace(node.Data)
		if text == "" {
			return nil // Skip whitespace-only text nodes
		}
		if strings.TrimSpace(text) == "" && strings.ContainsAny(text, "\n\r") {
			return nil
		}
		return map[string]interface{}{"text": text}
	}

	if node.Type == html.TextNode {
		return map[string]interface{}{"text": node.Data}
	}

	m := make(map[string]interface{})
	if node.Data != "" {
		m["type"] = node.Data
	}

	if node.Attr != nil {
		attrs := make(map[string]string)
		for _, attr := range node.Attr {
			attrs[attr.Key] = attr.Val
		}
		m["attr"] = attrs
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		childMap := nodeToMap(child)
		if child.Data != "" {
			if _, ok := m[child.Data]; ok {
				m[child.Data] = append(m[child.Data].([]interface{}), childMap)
			} else {
				m[child.Data] = []interface{}{childMap}
			}
		} else if text, ok := childMap["text"]; ok {
			m["text"] = text
		}
	}

	return m
}

func mapToNode(m map[string]interface{}) *html.Node {
	node := &html.Node{}
	if data, ok := m["type"].(string); ok {
		node.Type = html.ElementNode
		node.Data = data
	}

	if attrs, ok := m["attr"].(map[string]string); ok {
		for key, val := range attrs {
			node.Attr = append(node.Attr, html.Attribute{Key: key, Val: val})
		}
	}

	for key, val := range m {
		if key == "data" || key == "attr" || key == "text" {
			continue
		}

		children, ok := val.([]interface{})
		if !ok {
			continue
		}

		for _, child := range children {
			childMap, ok := child.(map[string]interface{})
			if !ok {
				continue
			}
			childNode := mapToNode(childMap)
			childNode.Data = key
			node.AppendChild(childNode)
		}
	}

	if text, ok := m["text"].(string); ok {
		node.Type = html.TextNode
		node.Data = text
	}

	return node
}
