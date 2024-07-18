package codec

import (
	"bytes"
	"github.com/goccy/go-json"
	"golang.org/x/net/html"
	"strings"
)

/*
HTML to Map Converter. These functions do not yet cover conversion to HTML, only from HTML to other arbitary output formats at this time.
This implementation may have some limitations and may not cover all edge cases.
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

func HTMLToMap(htmlBytes []byte) (map[string]interface{}, error) {
	doc, err := html.Parse(bytes.NewReader(htmlBytes))
	if err != nil {
		return nil, err
	}
	return nodeToMap(doc), nil
}

func nodeToMap(node *html.Node) map[string]interface{} {
	// handle text node edge cases
	if node.Type == html.TextNode {
		text := strings.TrimSpace(node.Data)
		if text != "" {
			if strings.TrimSpace(text) == "" && strings.ContainsAny(text, "\n\r") {
				return nil
			}
			return map[string]interface{}{"data": node.Data}
		}
	}

	// map initialization
	m := make(map[string]interface{})

	// Process attributes if present for node
	if node.Attr != nil {
		attrs := make(map[string]string)
		for _, attr := range node.Attr {
			attrs[attr.Key] = attr.Val
		}
		m["attr"] = attrs
	}

	// Recursively process children
	children := make(map[string][]interface{})
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		childMap := nodeToMap(child)
		if childMap != nil {
			if child.Type == html.ElementNode {
				children[child.Data] = append(children[child.Data], childMap)
			} else if data, ok := childMap["data"]; ok {
				m["data"] = data
			}
		}
	}

	// merge
	for key, value := range children {
		if len(value) == 1 {
			m[key] = value[0]
		} else {
			m[key] = value
		}
	}

	if len(m) == 0 {
		return nil
	}
	return m
}
