package codec

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

/*
TODO:
* convert title to top level key, convert sections to keys under title or section, so nested maps rather than lists
* table parsing, skip ---- content
*/
type ReadmeContent struct {
	Title    string        `json:"title"`
	Sections []interface{} `json:"sections"`
}

type CodeBlock struct {
	Language string `json:"language"`
	Text     string `json:"text"`
}

type Hyperlink struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

type Table []map[string]string

func markdownUnmarshal(data []byte, v interface{}) error {
	if v == nil {
		return errors.New("v cannot be nil")
	}

	content := parseReadme(string(data))

	jsonData, err := json.Marshal(content)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, v)
}

func parseReadme(content string) ReadmeContent {
	lines := strings.Split(content, "\n")
	var sections []interface{}
	var currentSection map[string]interface{}
	var title string
	var table Table
	var list []string
	inCodeBlock := false
	inTable := false
	codeLanguage := ""
	codeContent := []string{}
	headers := []string{}
	inList := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "```") {
			// Toggle code block state
			inCodeBlock = !inCodeBlock

			if inCodeBlock {
				codeLanguage = strings.TrimSpace(trimmedLine[3:])
				codeContent = []string{}
			} else {
				codeBlock := CodeBlock{
					Language: codeLanguage,
					Text:     strings.Join(codeContent, "\n"),
				}
				addToCurrentSection(&currentSection, "code_blocks", codeBlock)
			}
			continue
		}

		if inCodeBlock {
			codeContent = append(codeContent, line)
			continue
		}

		if strings.HasPrefix(trimmedLine, "# ") {
			// New top-level heading (title)
			if title == "" {
				title = strings.TrimSpace(trimmedLine[2:])
			} else {
				if currentSection != nil {
					if len(list) > 0 {
						addToCurrentSection(&currentSection, "lists", list)
						list = []string{}
					}
					sections = append(sections, currentSection)
				}
				currentSection = map[string]interface{}{
					"heading": strings.TrimSpace(trimmedLine[2:]),
				}
			}
			inList = false
		} else if strings.HasPrefix(trimmedLine, "## ") {
			// New section heading
			if currentSection != nil {
				if len(list) > 0 {
					addToCurrentSection(&currentSection, "lists", list)
					list = []string{}
				}
				sections = append(sections, currentSection)
			}
			currentSection = map[string]interface{}{
				"heading": strings.TrimSpace(trimmedLine[3:]),
			}
			inList = false
		} else if strings.HasPrefix(trimmedLine, "- ") || strings.HasPrefix(trimmedLine, "* ") {
			// List item
			if !inList && currentSection != nil {
				if len(list) > 0 {
					addToCurrentSection(&currentSection, "lists", list)
					list = []string{}
				}
				inList = true
			}
			if inList {
				list = append(list, strings.TrimSpace(trimmedLine[2:]))
			}
			continue
		} else if strings.Contains(trimmedLine, "|") && !inCodeBlock {
			// Table
			inTable = true
			cells := strings.Split(trimmedLine, "|")
			for i := range cells {
				cells[i] = strings.TrimSpace(cells[i])
			}

			if len(headers) == 0 {
				headers = cells[1 : len(cells)-1] // Ignore leading and trailing empty cells from split
			} else {
				if len(headers) > 0 {
					row := map[string]string{}
					for i, header := range headers {
						if i < len(cells) {
							row[header] = cells[i+1] // Skip leading empty cell
						}
					}
					table = append(table, row)
				}
			}
			inList = false
		} else if hyperlink := parseHyperlink(trimmedLine); hyperlink != nil {
			// Hyperlink
			addToCurrentSection(&currentSection, "hyperlinks", hyperlink)
			inList = false
		} else if trimmedLine != "" {
			// Paragraph (non-empty)
			if currentSection != nil && !inCodeBlock && !inTable {
				if len(list) > 0 {
					addToCurrentSection(&currentSection, "lists", list)
					list = []string{}
					inList = false
				}
				addToCurrentSection(&currentSection, "paragraphs", trimmedLine)
			}
		}

		if len(trimmedLine) == 0 || strings.HasPrefix(trimmedLine, "# ") || strings.HasPrefix(trimmedLine, "## ") {
			if len(list) > 0 {
				addToCurrentSection(&currentSection, "lists", list)
				list = []string{}
				inList = false
			}
		}

		if inTable && len(trimmedLine) == 0 {
			if len(table) > 0 {
				addToCurrentSection(&currentSection, "tables", table)
				table = nil
			}
			inTable = false
			headers = nil
		}
	}

	if len(list) > 0 {
		addToCurrentSection(&currentSection, "lists", list)
	}
	if currentSection != nil {
		sections = append(sections, currentSection)
	}

	return ReadmeContent{
		Title:    title,
		Sections: filterEmptyParagraphs(sections),
	}
}

func parseHyperlink(line string) *Hyperlink {
	// Regex to match Markdown hyperlinks: [text](url)
	re := regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) == 3 {
		return &Hyperlink{
			Text: matches[1],
			URL:  matches[2],
		}
	}
	return nil
}

func addToCurrentSection(section *map[string]interface{}, key string, value interface{}) {
	if *section == nil {
		*section = make(map[string]interface{})
	}
	if (*section)[key] == nil {
		(*section)[key] = []interface{}{}
	}
	(*section)[key] = append((*section)[key].([]interface{}), value)
}

func filterEmptyParagraphs(sections []interface{}) []interface{} {
	var filteredSections []interface{}

	for _, section := range sections {
		sec, ok := section.(map[string]interface{})
		if !ok {
			filteredSections = append(filteredSections, section)
			continue
		}

		if paragraphs, ok := sec["paragraphs"].([]interface{}); ok {
			var nonEmptyParagraphs []interface{}
			for _, para := range paragraphs {
				if paraStr, ok := para.(string); ok && strings.TrimSpace(paraStr) != "" {
					nonEmptyParagraphs = append(nonEmptyParagraphs, para)
				}
			}
			if len(nonEmptyParagraphs) > 0 {
				sec["paragraphs"] = nonEmptyParagraphs
				filteredSections = append(filteredSections, sec)
			}
		} else {
			filteredSections = append(filteredSections, sec)
		}
	}

	return filteredSections
}

