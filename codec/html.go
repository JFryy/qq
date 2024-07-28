package codec

import (
	"bytes"
	"github.com/goccy/go-json"
	"golang.org/x/net/html"
	"strings"
)

/*
HTML to Map Converter. These functions do not yet cover conversion to HTML, only from HTML to other arbitrary output formats at this time.
This implementation may have some limitations and may not cover all edge cases.
*/

func htmlUnmarshal(data []byte, v interface{}) error {
	htmlMap, err := HTMLToMap(data)
	if err != nil {
		return err
	}
	b, err := json.Marshal(htmlMap)
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

	// Always handle presence of root html node
	var root *html.Node
	for node := doc.FirstChild; node != nil; node = node.NextSibling {
		if node.Type == html.ElementNode && node.Data == "html" {
			root = node
			break
		}
	}

	if root == nil {
		return nil, nil
	}

	result := nodeToMap(root)
	if m, ok := result.(map[string]interface{}); ok {
		return map[string]interface{}{"html": m}, nil
	}
	return nil, nil
}

func nodeToMap(node *html.Node) interface{} {
	m := make(map[string]interface{})

	// Process attributes if present for node
	if node.Attr != nil {
		for _, attr := range node.Attr {
			m["@"+attr.Key] = attr.Val
		}
	}

	// Recursively process all the children
	var childTexts []string
	var comments []string
	children := make(map[string][]interface{})
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		switch child.Type {
		case html.TextNode:
			text := strings.TrimSpace(child.Data)
			if text != "" && !(strings.TrimSpace(text) == "" && strings.ContainsAny(text, "\n\r")) {
				text, _ = strings.CutSuffix(text, "\n\r")
				text, _ = strings.CutPrefix(text, "\n")
				childTexts = append(childTexts, text)
			}
		case html.CommentNode:
			text := strings.TrimSpace(child.Data)
			if text != "" && !(strings.TrimSpace(text) == "" && strings.ContainsAny(text, "\n\r")) {
				text, _ = strings.CutSuffix(text, "\n\r")
				text, _ = strings.CutPrefix(text, "\n")
				comments = append(comments, text)
			}
		case html.ElementNode:
			childMap := nodeToMap(child)
			if childMap != nil {
				children[child.Data] = append(children[child.Data], childMap)
			}
		}
	}

	// Merge children into one
	for key, value := range children {
		if len(value) == 1 {
			m[key] = value[0]
		} else {
			m[key] = value
		}
	}

	// Handle the children's text
	if len(childTexts) > 0 {
		if len(childTexts) == 1 {
			if len(m) == 0 {
				return childTexts[0]
			}
			m["#text"] = childTexts[0]
		} else {
			m["#text"] = strings.Join(childTexts, " ")
		}
	}

	// Handle comments
	if len(comments) > 0 {
		if len(comments) == 1 {
			if len(m) == 0 {
				return map[string]interface{}{"#comment": comments[0]}
			} else {
				m["#comment"] = comments[0]
			}
		} else {
			m["#comment"] = comments
		}
	}

	if len(m) == 0 {
		return nil
	} else if len(m) == 1 {
		if text, ok := m["#text"]; ok {
			return text
		}
		if len(node.Attr) == 0 {
			for key, val := range m {
				if childMap, ok := val.(map[string]interface{}); ok && len(childMap) == 1 {
					return val
				}
				return map[string]interface{}{key: val}
			}
		}
	}

	return m
}
