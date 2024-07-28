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

	// Handle text proceeding and following whitespace, newlines, etc.
	if node.Type == html.TextNode {
		text := strings.TrimSpace(node.Data)
		if text == "" {
			return nil
		}
		if strings.TrimSpace(text) == "" && strings.ContainsAny(text, "\n\r") {
			return nil
		}
		text, _ = strings.CutSuffix(text, "\n\r")
		text, _ = strings.CutPrefix(text, "\n")
		return text
	}

	if node.Type == html.CommentNode {
		text := strings.TrimSpace(node.Data)
		if text == "" {
			return nil
		}
		if strings.TrimSpace(text) == "" && strings.ContainsAny(text, "\n\r") {
			return nil
		}
		text, _ = strings.CutSuffix(text, "\n\r")
		text, _ = strings.CutPrefix(text, "\n")
		return map[string]interface{}{"#comment": text}

	}

	m := make(map[string]interface{})

	// Process attributes if present for node
	if node.Attr != nil {
		for _, attr := range node.Attr {
			m["@"+attr.Key] = attr.Val
		}
	}

	// Recursively process children
	var childTexts []string
	var comments []string
	children := make(map[string][]interface{})
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		childMap := nodeToMap(child)
		if childMap != nil {
			if child.Type == html.ElementNode {
				children[child.Data] = append(children[child.Data], childMap)
			} else if text, ok := childMap.(string); ok {
				childTexts = append(childTexts, text)
			} else if comment, ok := childMap.(map[string]interface{}); ok && comment["#comment"] != nil {
				if commentText, ok := comment["#comment"].(string); ok {
					comments = append(comments, commentText)
				}
			}
		}
	}

	// Merge children into map
	for key, value := range children {
		if len(value) == 1 {
			m[key] = value[0]
		} else {
			m[key] = value
		}
	}

	// Handle text content if present
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

	// Handle comments if present
	if len(comments) > 0 {
		if len(comments) == 1 {
			if len(m) == 0 {
				return map[string]interface{}{"#comment": comments[0]}
			} else {
				m["#comment"] = comments[0]
			}
		}
	}

	// Simplify map if only contains text or single child element
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

