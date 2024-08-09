package markdown

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

type CodeBlock struct {
	Lang string `json:"lang"`
	Text string `json:"text"`
}

type Hyperlink struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

type Table []map[string]string

type Codec struct {
	Section    map[string]interface{}
	Subsection map[string]interface{}
	InCodeBlock bool
	InTable     bool
}

func (m *Codec) Unmarshal(data []byte, v interface{}) error {
	if v == nil {
		return errors.New("v cannot be nil")
	}

	content := m.parseReadme(string(data))

	jsonData, err := json.Marshal(content)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, v)
}

func (m *Codec) parseHyperlink(line string) *Hyperlink {
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

func (m *Codec) parseReadme(content string) interface{} {
	lines := strings.Split(content, "\n")
	sections := make(map[string]interface{})
	var title string
	var table Table
	var list []string
	var orderedList []string
	inCodeBlock := false
	inTable := false
	codeLanguage := ""
	codeContent := []string{}
	headers := []string{}
	inList := false
	inOrderedList := false
	var currentHeading string
	re := regexp.MustCompile("^[1-9]+. ")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(trimmedLine, "```"):
			// Toggle code block state
			inCodeBlock = !inCodeBlock
			if inCodeBlock {
				codeLanguage = strings.TrimSpace(trimmedLine[3:])
				codeContent = []string{}
			} else {
				codeBlock := CodeBlock{
					Lang: codeLanguage,
					Text: strings.Join(codeContent, "\n"),
				}
				if m.Subsection != nil {
					m.addToSubsection(&m.Subsection, "code", codeBlock)
				} else if m.Section != nil {
					m.addToSubsection(&m.Section, "code", codeBlock)
				}
			}
			continue

		case inCodeBlock:
			codeContent = append(codeContent, line)
			continue

		case strings.HasPrefix(trimmedLine, "# "):
			if title == "" {
				title = strings.TrimSpace(trimmedLine[2:])
			} else {
				// Finalize the current section before starting a new one
				if m.Subsection != nil {
					if len(list) > 0 {
						m.addToSubsection(&m.Subsection, "lists", list)
						list = []string{}
					}
					if m.Section != nil {
						heading := (m.Subsection)["heading"].(string)
						m.addToSubsection(&m.Section, heading, m.Subsection)
					}
					m.Subsection = nil
				}
				if m.Section != nil {
					if len(list) > 0 {
						m.addToSubsection(&m.Section, "lists", list)
						list = []string{}
					}
					sections[currentHeading] = m.Section
				}
			}
			currentHeading = strings.TrimSpace(trimmedLine[2:])
			newSection := make(map[string]interface{})
			m.Section = newSection
			inList = false
			inOrderedList = false

		case strings.HasPrefix(trimmedLine, "##"):
			// New subsection heading
			if m.Section != nil {
				if len(list) > 0 {
					m.addToSubsection(&m.Section, "lists", list)
					list = []string{}
				}
				if m.Subsection != nil {
					if len(list) > 0 {
						m.addToSubsection(&m.Subsection, "lists", list)
						list = []string{}
					}
					heading := (m.Subsection)["heading"].(string)
					m.addToSubsection(&m.Section, heading, m.Subsection)
				}
				newSubsection := make(map[string]interface{})
				m.Subsection = newSubsection
				(m.Subsection)["heading"] = strings.TrimSpace(trimmedLine[3:])
			}
			inList = false

		case strings.HasPrefix(trimmedLine, "- ") || strings.HasPrefix(trimmedLine, "* "):
			if !inList && (m.Section != nil || m.Subsection != nil) {
				if len(list) > 0 {
					if m.Subsection != nil {
						m.addToSubsection(&m.Subsection, "lists", list)
					} else {
						m.addToSubsection(&m.Section, "lists", list)
					}
					list = []string{}
				}
				inList = true
			}
			if inList {
				list = append(list, strings.TrimSpace(trimmedLine[2:]))
			}
			continue

		case re.MatchString(trimmedLine):
			if !inOrderedList && (m.Section != nil || m.Subsection != nil) {
				if len(orderedList) > 0 {
					if m.Subsection != nil {
						m.addToSubsection(&m.Subsection, "ol", orderedList)
					} else {
						m.addToSubsection(&m.Section, "ol", orderedList)

					}
					orderedList = []string{}
				}
				inOrderedList = true
			}
			if inOrderedList {
				orderedList = append(orderedList, strings.TrimSpace(trimmedLine[3:]))
			}
			continue

		case strings.Contains(trimmedLine, "|") && !inCodeBlock:
			// skip below table header
			if strings.HasPrefix(trimmedLine, "|-") {
				continue
			}
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
			inOrderedList = false

		case m.parseHyperlink(trimmedLine) != nil:
			// Hyperlink
			hyperlink := m.parseHyperlink(trimmedLine)
			if m.Subsection != nil {
				m.addToSubsection(&m.Subsection, "links", hyperlink)
			} else if m.Section != nil {
				m.addToSubsection(&m.Section, "links", hyperlink)
			}
			inList = false
			inOrderedList = false

		case trimmedLine != "":
			// Paragraph (non-empty)
			if m.Section != nil && !inCodeBlock && !inTable {
				if len(list) > 0 {
					if m.Subsection != nil {
						m.addToSubsection(&m.Subsection, "li", list)
					} else {
						m.addToSubsection(&m.Section, "li", list)
					}
					list = []string{}
					inList = false
					inOrderedList = false
				}
				if m.Subsection != nil {
					m.addToSubsection(&m.Subsection, "p", trimmedLine)
				} else {
					m.addToSubsection(&m.Section, "p", trimmedLine)
				}
			}

		case len(trimmedLine) == 0 || strings.HasPrefix(trimmedLine, "# ") || strings.HasPrefix(trimmedLine, "## "):
			if len(list) > 0 {
				if m.Subsection != nil {
					m.addToSubsection(&m.Subsection, "li", list)
				} else if m.Section != nil {
					m.addToSubsection(&m.Section, "li", list)
				}
				list = []string{}
				inList = false
			} else if len(orderedList) > 0 {
				if m.Subsection != nil {
					m.addToSubsection(&m.Subsection, "ol", orderedList)
				} else if m.Section != nil {
					m.addToSubsection(&m.Section, "ol", orderedList)
				}
				orderedList = []string{}
				inOrderedList = false
			}

		case inTable && len(trimmedLine) == 0:
			if len(table) > 0 {
				if m.Subsection != nil {
					m.addToSubsection(&m.Subsection, "table", table)
				} else if m.Section != nil {
					m.addToSubsection(&m.Section, "table", table)
				}
				table = nil
				headers = []string{}
				inTable = false
			}
			continue
		}
	}

	if m.Subsection != nil && m.Section != nil {
		m.addToSubsection(&m.Section, m.Subsection["heading"].(string), m.Subsection)
	}
	if m.Section != nil {
		sections[currentHeading] = m.Section
	}

	sections["title"] = title
	return sections
}

func (m *Codec) addToSubsection(subsection *map[string]interface{}, key string, value interface{}) {
	if subsection == nil || *subsection == nil {
		return
	}
	if existing, ok := (*subsection)[key].([]interface{}); ok {
		(*subsection)[key] = append(existing, value)
	} else {
		(*subsection)[key] = []interface{}{value}
	}
}

