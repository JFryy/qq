package codec

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

/*
TODO:
* add support for nested headings
* get hyperlink paragraph, list content but also append to hyperlink field
* remove header key from sections, it is redundant
*/

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

func parseHyperlink(line string) *Hyperlink {
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

func parseReadme(content string) interface{} {
	lines := strings.Split(content, "\n")
	sections := make(map[string]interface{})
	var currentSection *map[string]interface{}
	var currentSubsection *map[string]interface{}
	var title string
	var table Table
	var list []string
	inCodeBlock := false
	inTable := false
	codeLanguage := ""
	codeContent := []string{}
	headers := []string{}
	inList := false
	var currentHeading string

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
				if currentSubsection != nil {
					addToCurrentSubsection(currentSubsection, "blocks", codeBlock)
				} else if currentSection != nil {
					addToCurrentSubsection(currentSection, "blocks", codeBlock)
				}
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
				// Finalize the current section before starting a new one
				if currentSubsection != nil {
					if len(list) > 0 {
						addToCurrentSubsection(currentSubsection, "lists", list)
						list = []string{}
					}
					if currentSection != nil {
                        heading := (*currentSubsection)["heading"].(string)
						addToCurrentSubsection(currentSection, heading, *currentSubsection)
					}
					currentSubsection = nil
				}
				if currentSection != nil {
					if len(list) > 0 {
						addToCurrentSubsection(currentSection, "lists", list)
						list = []string{}
					}
					sections[currentHeading] = *currentSection
				}
			}
			currentHeading = strings.TrimSpace(trimmedLine[2:])
			newSection := make(map[string]interface{})
			currentSection = &newSection
			inList = false
		} else if strings.HasPrefix(trimmedLine, "##") {
			// New subsection heading
			if currentSection != nil {
				if len(list) > 0 {
					addToCurrentSubsection(currentSection, "lists", list)
					list = []string{}
				}
				if currentSubsection != nil {
					if len(list) > 0 {
						addToCurrentSubsection(currentSubsection, "lists", list)
						list = []string{}
					}
                    heading := (*currentSubsection)["heading"].(string)
					addToCurrentSubsection(currentSection, heading, *currentSubsection)
				}
				newSubsection := make(map[string]interface{})
				currentSubsection = &newSubsection
				(*currentSubsection)["heading"] = strings.TrimSpace(trimmedLine[3:])
			}
			inList = false
		} else if strings.HasPrefix(trimmedLine, "- ") || strings.HasPrefix(trimmedLine, "* ") {
			if !inList && (currentSection != nil || currentSubsection != nil) {
				if len(list) > 0 {
					if currentSubsection != nil {
						addToCurrentSubsection(currentSubsection, "lists", list)
					} else {
						addToCurrentSubsection(currentSection, "lists", list)
					}
					list = []string{}
				}
				inList = true
			}
			if inList {
				list = append(list, strings.TrimSpace(trimmedLine[2:]))
			}
			continue
		} else if strings.Contains(trimmedLine, "|") && !inCodeBlock {
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
		} else if hyperlink := parseHyperlink(trimmedLine); hyperlink != nil {
			// Hyperlink
			if currentSubsection != nil {
				addToCurrentSubsection(currentSubsection, "hyperlinks", hyperlink)
			} else if currentSection != nil {
				addToCurrentSubsection(currentSection, "hyperlinks", hyperlink)
			}
			inList = false
		} else if trimmedLine != "" {
			// Paragraph (non-empty)
			if currentSection != nil && !inCodeBlock && !inTable {
				if len(list) > 0 {
					if currentSubsection != nil {
						addToCurrentSubsection(currentSubsection, "lists", list)
					} else {
						addToCurrentSubsection(currentSection, "lists", list)
					}
					list = []string{}
					inList = false
				}
				if currentSubsection != nil {
					addToCurrentSubsection(currentSubsection, "paragraphs", trimmedLine)
				} else {
					addToCurrentSubsection(currentSection, "paragraphs", trimmedLine)
				}
			}
		}

		if len(trimmedLine) == 0 || strings.HasPrefix(trimmedLine, "# ") || strings.HasPrefix(trimmedLine, "## ") {
			if len(list) > 0 {
				if currentSubsection != nil {
					addToCurrentSubsection(currentSubsection, "lists", list)
				} else if currentSection != nil {
					addToCurrentSubsection(currentSection, "lists", list)
				}
				list = []string{}
				inList = false
			}
		}

		if inTable && len(trimmedLine) == 0 {
			if len(table) > 0 {
				if currentSubsection != nil {
					addToCurrentSubsection(currentSubsection, "tables", table)
				} else if currentSection != nil {
					addToCurrentSubsection(currentSection, "tables", table)
				}
				table = nil
			}
			inTable = false
			headers = nil
		}
	}

	if len(list) > 0 {
		if currentSubsection != nil {
			addToCurrentSubsection(currentSubsection, "lists", list)
		} else if currentSection != nil {
			addToCurrentSubsection(currentSection, "lists", list)
		}
	}
	if currentSubsection != nil {
		if currentSection != nil {
            heading := (*currentSubsection)["heading"].(string)
			addToCurrentSubsection(currentSection, heading, *currentSubsection)
		}
	}
	if currentSection != nil {
		sections[currentHeading] = *currentSection
	}

    return sections
}

func addToCurrentSubsection(subsection *map[string]interface{}, key string, value interface{}) {
	if subsection == nil || *subsection == nil {
		return
	}
	if existing, ok := (*subsection)[key].([]interface{}); ok {
		(*subsection)[key] = append(existing, value)
	} else {
		(*subsection)[key] = []interface{}{value}
	}
}

