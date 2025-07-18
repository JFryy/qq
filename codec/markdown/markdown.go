package markdown

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Node represents the base interface for all markdown elements
type Node interface {
	Type() string
	Children() []Node
	SetChildren([]Node)
}

// NodeWrapper wraps a Node for JSON serialization
type NodeWrapper struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// MarshalJSON implements custom JSON marshaling for nodes
func (nw NodeWrapper) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: nw.Type,
		Data: nw.Data,
	})
}

// Document represents the root markdown document
type Document struct {
	Title    string                 `json:"title,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Content  []interface{}          `json:"content"`
}

func (d *Document) Type() string { return "document" }
func (d *Document) Children() []Node {
	nodes := make([]Node, 0, len(d.Content))
	for _, item := range d.Content {
		if node, ok := item.(Node); ok {
			nodes = append(nodes, node)
		}
	}
	return nodes
}
func (d *Document) SetChildren(children []Node) {
	d.Content = make([]interface{}, len(children))
	for i, child := range children {
		d.Content[i] = child
	}
}

// Heading represents markdown headers (# ## ### etc)
type Heading struct {
	Level   int           `json:"level"`
	Text    string        `json:"text"`
	ID      string        `json:"id,omitempty"`
	Content []interface{} `json:"content,omitempty"`
}

func (h *Heading) Type() string { return "heading" }
func (h *Heading) Children() []Node {
	nodes := make([]Node, 0, len(h.Content))
	for _, item := range h.Content {
		if node, ok := item.(Node); ok {
			nodes = append(nodes, node)
		}
	}
	return nodes
}
func (h *Heading) SetChildren(children []Node) {
	h.Content = make([]interface{}, len(children))
	for i, child := range children {
		h.Content[i] = child
	}
}

// Paragraph represents regular text paragraphs
type Paragraph struct {
	Text string `json:"text"`
}

func (p *Paragraph) Type() string       { return "paragraph" }
func (p *Paragraph) Children() []Node   { return nil }
func (p *Paragraph) SetChildren([]Node) {}

// CodeBlock represents fenced code blocks
type CodeBlock struct {
	Language string `json:"language,omitempty"`
	Code     string `json:"code"`
	Info     string `json:"info,omitempty"`
}

func (cb *CodeBlock) Type() string       { return "code_block" }
func (cb *CodeBlock) Children() []Node   { return nil }
func (cb *CodeBlock) SetChildren([]Node) {}

// List represents both ordered and unordered lists
type List struct {
	Ordered bool       `json:"ordered"`
	Items   []ListItem `json:"items"`
}

func (l *List) Type() string { return "list" }
func (l *List) Children() []Node {
	nodes := make([]Node, len(l.Items))
	for i, item := range l.Items {
		nodes[i] = &item
	}
	return nodes
}
func (l *List) SetChildren(children []Node) {
	l.Items = make([]ListItem, len(children))
	for i, child := range children {
		if item, ok := child.(*ListItem); ok {
			l.Items[i] = *item
		}
	}
}

// ListItem represents individual list items
type ListItem struct {
	Text     string        `json:"text"`
	Checkbox *bool         `json:"checkbox,omitempty"` // nil=no checkbox, true=checked, false=unchecked
	Content  []interface{} `json:"content,omitempty"`
}

func (li *ListItem) Type() string { return "list_item" }
func (li *ListItem) Children() []Node {
	nodes := make([]Node, 0, len(li.Content))
	for _, item := range li.Content {
		if node, ok := item.(Node); ok {
			nodes = append(nodes, node)
		}
	}
	return nodes
}
func (li *ListItem) SetChildren(children []Node) {
	li.Content = make([]interface{}, len(children))
	for i, child := range children {
		li.Content[i] = child
	}
}

// Table represents markdown tables
type Table struct {
	Headers []string            `json:"headers"`
	Rows    []map[string]string `json:"rows"`
	Align   []string            `json:"align,omitempty"` // left, right, center
}

func (t *Table) Type() string       { return "table" }
func (t *Table) Children() []Node   { return nil }
func (t *Table) SetChildren([]Node) {}

// Link represents hyperlinks
type Link struct {
	Text  string `json:"text"`
	URL   string `json:"url"`
	Title string `json:"title,omitempty"`
}

func (l *Link) Type() string       { return "link" }
func (l *Link) Children() []Node   { return nil }
func (l *Link) SetChildren([]Node) {}

// Image represents images
type Image struct {
	Alt   string `json:"alt"`
	URL   string `json:"url"`
	Title string `json:"title,omitempty"`
}

func (i *Image) Type() string       { return "image" }
func (i *Image) Children() []Node   { return nil }
func (i *Image) SetChildren([]Node) {}

// BlockQuote represents block quotes
type BlockQuote struct {
	Content []interface{} `json:"content"`
}

func (bq *BlockQuote) Type() string { return "blockquote" }
func (bq *BlockQuote) Children() []Node {
	nodes := make([]Node, 0, len(bq.Content))
	for _, item := range bq.Content {
		if node, ok := item.(Node); ok {
			nodes = append(nodes, node)
		}
	}
	return nodes
}
func (bq *BlockQuote) SetChildren(children []Node) {
	bq.Content = make([]interface{}, len(children))
	for i, child := range children {
		bq.Content[i] = child
	}
}

// HorizontalRule represents horizontal rules (---, ***, ___)
type HorizontalRule struct{}

func (hr *HorizontalRule) Type() string       { return "horizontal_rule" }
func (hr *HorizontalRule) Children() []Node   { return nil }
func (hr *HorizontalRule) SetChildren([]Node) {}

// Codec represents the markdown parser
type Codec struct {
	strict bool // Whether to strictly validate markdown syntax
}

// NewCodec creates a new markdown codec instance
func NewCodec() *Codec {
	return &Codec{strict: false}
}

// NewStrictCodec creates a new markdown codec with strict validation
func NewStrictCodec() *Codec {
	return &Codec{strict: true}
}

// Unmarshal parses markdown data into the provided interface
func (c *Codec) Unmarshal(data []byte, v any) error {
	if v == nil {
		return errors.New("v cannot be nil")
	}

	result, err := c.Parse(string(data))
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, v)
}

// Parse parses markdown content into a hierarchical structure
func (c *Codec) Parse(content string) (map[string]interface{}, error) {
	parser := &hierarchicalParser{
		lines:  strings.Split(content, "\n"),
		pos:    0,
		strict: c.strict,
	}

	return parser.parse()
}

// hierarchicalParser creates a hierarchical structure based on headings
type hierarchicalParser struct {
	lines  []string
	pos    int
	strict bool
}

// Section represents a hierarchical section with semantic structure
type Section struct {
	ID       string              `json:"id"`
	Title    string              `json:"title"`
	Content  []ContentItem       `json:"content,omitempty"`
	Sections map[string]*Section `json:"sections,omitempty"`
}

// ContentItem represents any piece of content with its type preserved
type ContentItem struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// parse processes the entire document into hierarchical structure
func (p *hierarchicalParser) parse() (map[string]interface{}, error) {
	// Track heading hierarchy
	var sectionStack []*Section
	var levelStack []int
	var rootSections = make(map[string]*Section)
	var currentSection *Section
	idCounter := make(map[string]int)

	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])

		// Skip empty lines
		if line == "" {
			p.pos++
			continue
		}

		// Check if this is a heading
		if strings.HasPrefix(line, "#") {
			level, text := p.parseHeadingLine(line)

			// Generate stable ID for this section
			id := p.generateSectionID(text, idCounter)

			// Create new section
			section := &Section{
				ID:       id,
				Title:    text,
				Content:  make([]ContentItem, 0),
				Sections: make(map[string]*Section),
			}

			// Pop stack until we find appropriate parent level
			for len(levelStack) > 0 && levelStack[len(levelStack)-1] >= level {
				sectionStack = sectionStack[:len(sectionStack)-1]
				levelStack = levelStack[:len(levelStack)-1]
			}

			// Determine where to place this section
			if len(sectionStack) == 0 {
				// Top-level section
				rootSections[id] = section
			} else {
				// Child of the current parent section
				parent := sectionStack[len(sectionStack)-1]
				parent.Sections[id] = section
			}

			// Add to stack and update current
			sectionStack = append(sectionStack, section)
			levelStack = append(levelStack, level)
			currentSection = section

			p.pos++
		} else {
			// Parse content block
			contentItem, err := p.parseContentItem()
			if err != nil {
				if p.strict {
					return nil, err
				}
				p.pos++
				continue
			}

			if contentItem != nil {
				// Add content to the current section
				if currentSection != nil {
					currentSection.Content = append(currentSection.Content, *contentItem)
				} else {
					// No section yet, we need to handle orphaned content
					// For now, skip it or could create a default section
					p.pos++
				}
			}
		}
	}

	// Build final result - convert sections to the expected interface format
	result := make(map[string]interface{})
	for id, section := range rootSections {
		result[id] = section
	}

	return result, nil
}

// parseHeadingLine extracts heading level and text
func (p *hierarchicalParser) parseHeadingLine(line string) (int, string) {
	level := 0
	for i, char := range line {
		if char == '#' {
			level++
		} else {
			text := strings.TrimSpace(line[i:])
			return level, text
		}
	}
	return level, ""
}

// generateSectionID creates a stable ID from section title
func (p *hierarchicalParser) generateSectionID(title string, counter map[string]int) string {
	// Convert to lowercase and replace non-alphanumeric with hyphens
	id := strings.ToLower(title)
	id = regexp.MustCompile(`[^a-z0-9\s-]`).ReplaceAllString(id, "")
	id = regexp.MustCompile(`\s+`).ReplaceAllString(id, "-")
	id = strings.Trim(id, "-")

	// Handle duplicates
	if count, exists := counter[id]; exists {
		counter[id]++
		return fmt.Sprintf("%s-%d", id, count)
	}
	counter[id] = 1
	return id
}

// parseContentItem parses a content block and returns a typed ContentItem
func (p *hierarchicalParser) parseContentItem() (*ContentItem, error) {
	if p.pos >= len(p.lines) {
		return nil, nil
	}

	line := p.lines[p.pos]
	trimmed := strings.TrimSpace(line)

	// Parse different block types and wrap them in ContentItem
	switch {
	case strings.HasPrefix(trimmed, "```"):
		codeBlock, err := p.parseCodeBlock()
		if err != nil {
			return nil, err
		}
		return &ContentItem{Type: "code_block", Data: codeBlock}, nil

	case strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* "):
		list, err := p.parseList(false)
		if err != nil {
			return nil, err
		}
		return &ContentItem{Type: "list", Data: list}, nil

	case p.isOrderedListStart(trimmed):
		list, err := p.parseList(true)
		if err != nil {
			return nil, err
		}
		return &ContentItem{Type: "list", Data: list}, nil

	case strings.Contains(trimmed, "|"):
		table, err := p.parseTable()
		if err != nil {
			return nil, err
		}
		return &ContentItem{Type: "table", Data: table}, nil

	case strings.HasPrefix(trimmed, ">"):
		blockquote, err := p.parseBlockQuote()
		if err != nil {
			return nil, err
		}
		return &ContentItem{Type: "blockquote", Data: blockquote}, nil

	case p.isHorizontalRule(trimmed):
		hr, err := p.parseHorizontalRule()
		if err != nil {
			return nil, err
		}
		return &ContentItem{Type: "horizontal_rule", Data: hr}, nil

	default:
		paragraph, err := p.parseParagraph()
		if err != nil {
			return nil, err
		}
		return &ContentItem{Type: "paragraph", Data: paragraph}, nil
	}
}

// parseCodeBlock parses fenced code blocks
func (p *hierarchicalParser) parseCodeBlock() (*CodeBlock, error) {
	if p.pos >= len(p.lines) {
		return nil, errors.New("unexpected end of file in code block")
	}

	firstLine := strings.TrimSpace(p.lines[p.pos])
	language := strings.TrimSpace(firstLine[3:]) // Remove ```
	p.pos++

	var codeLines []string
	for p.pos < len(p.lines) {
		line := p.lines[p.pos]
		if strings.TrimSpace(line) == "```" {
			p.pos++
			return &CodeBlock{
				Language: language,
				Code:     strings.Join(codeLines, "\n"),
			}, nil
		}
		codeLines = append(codeLines, line)
		p.pos++
	}

	if p.strict {
		return nil, errors.New("unclosed code block")
	}

	return &CodeBlock{
		Language: language,
		Code:     strings.Join(codeLines, "\n"),
	}, nil
}

// parseList parses both ordered and unordered lists
func (p *hierarchicalParser) parseList(ordered bool) (*List, error) {
	list := &List{
		Ordered: ordered,
		Items:   make([]ListItem, 0),
	}

	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])

		if line == "" {
			p.pos++
			continue
		}

		var text string
		var checkbox *bool

		if ordered {
			if !p.isOrderedListStart(line) {
				break
			}
			// Find the first space after the number and dot
			spaceIdx := strings.Index(line, " ")
			if spaceIdx > 0 {
				text = strings.TrimSpace(line[spaceIdx:])
			}
		} else {
			if !strings.HasPrefix(line, "- ") && !strings.HasPrefix(line, "* ") {
				break
			}
			text = strings.TrimSpace(line[2:])
		}

		// Check for task list checkbox
		if strings.HasPrefix(text, "[ ]") {
			unchecked := false
			checkbox = &unchecked
			text = strings.TrimSpace(text[3:])
		} else if strings.HasPrefix(text, "[x]") || strings.HasPrefix(text, "[X]") {
			checked := true
			checkbox = &checked
			text = strings.TrimSpace(text[3:])
		}

		list.Items = append(list.Items, ListItem{
			Text:     text,
			Checkbox: checkbox,
		})
		p.pos++
	}

	return list, nil
}

// parseTable parses markdown tables
func (p *hierarchicalParser) parseTable() (*Table, error) {
	var headers []string
	var rows []map[string]string

	// Parse header row
	if p.pos >= len(p.lines) {
		return nil, errors.New("unexpected end of table")
	}

	headerLine := strings.TrimSpace(p.lines[p.pos])
	headers = p.parseTableRow(headerLine)
	p.pos++

	// Skip separator row (|---|---|)
	if p.pos < len(p.lines) && strings.Contains(p.lines[p.pos], "---") {
		p.pos++
	}

	// Parse data rows
	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])
		if line == "" || !strings.Contains(line, "|") {
			break
		}

		cells := p.parseTableRow(line)
		row := make(map[string]string)
		for i, cell := range cells {
			if i < len(headers) {
				row[headers[i]] = cell
			}
		}
		rows = append(rows, row)
		p.pos++
	}

	return &Table{
		Headers: headers,
		Rows:    rows,
	}, nil
}

// parseTableRow splits a table row into cells
func (p *hierarchicalParser) parseTableRow(line string) []string {
	cells := strings.Split(line, "|")
	result := make([]string, 0)

	for _, cell := range cells {
		trimmed := strings.TrimSpace(cell)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// parseBlockQuote parses block quotes
func (p *hierarchicalParser) parseBlockQuote() (*BlockQuote, error) {
	var lines []string

	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])
		if !strings.HasPrefix(line, ">") {
			break
		}

		content := strings.TrimSpace(line[1:])
		lines = append(lines, content)
		p.pos++
	}

	// For now, treat blockquote content as simple text paragraphs
	// In a more sophisticated implementation, we could parse the content recursively
	content := strings.Join(lines, " ")

	return &BlockQuote{
		Content: []interface{}{
			ContentItem{
				Type: "paragraph",
				Data: &Paragraph{Text: content},
			},
		},
	}, nil
}

// parseHorizontalRule parses horizontal rules
func (p *hierarchicalParser) parseHorizontalRule() (*HorizontalRule, error) {
	p.pos++
	return &HorizontalRule{}, nil
}

// parseParagraph parses regular paragraphs
func (p *hierarchicalParser) parseParagraph() (*Paragraph, error) {
	var lines []string

	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])
		if line == "" || p.isBlockStart(line) {
			break
		}
		lines = append(lines, line)
		p.pos++
	}

	text := strings.Join(lines, " ")

	// Process inline elements like links and images
	text = p.processInlineElements(text)

	return &Paragraph{Text: text}, nil
}

// Helper functions

func (p *hierarchicalParser) isOrderedListStart(line string) bool {
	re := regexp.MustCompile(`^\d+\.\s`)
	return re.MatchString(line)
}

func (p *hierarchicalParser) isHorizontalRule(line string) bool {
	line = strings.ReplaceAll(line, " ", "")
	return len(line) >= 3 &&
		(strings.Count(line, "-") == len(line) ||
			strings.Count(line, "*") == len(line) ||
			strings.Count(line, "_") == len(line))
}

func (p *hierarchicalParser) isBlockStart(line string) bool {
	return strings.HasPrefix(line, "#") ||
		strings.HasPrefix(line, "```") ||
		strings.HasPrefix(line, "- ") ||
		strings.HasPrefix(line, "* ") ||
		strings.HasPrefix(line, ">") ||
		strings.Contains(line, "|") ||
		p.isOrderedListStart(line) ||
		p.isHorizontalRule(line)
}

func (p *hierarchicalParser) processInlineElements(text string) string {
	// Process links and images - this could be expanded for more inline elements
	linkRe := regexp.MustCompile(`\[([^\]]*)\]\(([^)]*)\)`)
	imageRe := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]*)\)`)

	// For now, just return the text as-is
	// In a full implementation, you'd want to parse these into separate nodes
	// or return structured data about inline elements

	_ = linkRe
	_ = imageRe

	return text
}

// Marshal method for codec interface compliance
func (c *Codec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}
