package tui

import (
	"fmt"
	"github.com/goccy/go-json"
	"os"
	"strings"

	"github.com/JFryy/qq/codec"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/itchyny/gojq"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	cursorStyle  = focusedStyle
	previewStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("178")).Italic(true)
	outputStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("36"))
)

type model struct {
	inputs         []textinput.Model
	jsonInput      string
	jqOutput       string
	lastOutput     string
	currentIndex   int
	showingPreview bool
	jqOptions      []string
	suggestedValue string
	jsonObj        interface{}
	viewport       viewport.Model
}

func newModel(data string) model {
	m := model{
		inputs:   make([]textinput.Model, 1),
		viewport: viewport.New(0, 0),
	}

	t := textinput.New()
	t.Cursor.Style = cursorStyle
	t.Placeholder = "Enter jq filter"
	t.SetValue(".")
	t.Focus()
	t.PromptStyle = focusedStyle
	t.TextStyle = focusedStyle
	m.inputs[0] = t
	m.jsonInput = string(data)
	m.jqOptions = generateJqOptions(m.jsonInput)

	m.runJqFilter()
	m.jsonObj, _ = jsonStrToInterface(m.jsonInput)

	return m
}

func generateJqOptions(jsonStr string) []string {
	var jsonData interface{}
	err := json.Unmarshal([]byte(jsonStr), &jsonData)
	if err != nil {
		return []string{"."}
	}

	options := make(map[string]struct{})
	extractPaths(jsonData, "", options)

	// Convert map to slice
	var result []string
	for option := range options {
		result = append(result, option)
	}
	return result
}

func extractPaths(data interface{}, prefix string, options map[string]struct{}) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			newPrefix := prefix + "." + key
			options[newPrefix] = struct{}{}
			extractPaths(value, newPrefix, options)
		}
	case []interface{}:
		for i, item := range v {
			newPrefix := fmt.Sprintf("%s[%d]", prefix, i)
			options[newPrefix] = struct{}{}
			extractPaths(item, newPrefix, options)
		}
	}
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := 2
		footerHeight := 1
		availableHeight := msg.Height - headerHeight - footerHeight
		m.viewport.Width = msg.Width
		m.viewport.Height = availableHeight
		m.updateViewportContent()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

			// Suggest next jq option
		case "tab":
			if !m.showingPreview {
				m.showingPreview = true
				m.currentIndex = 0
			} else {
				m.currentIndex = (m.currentIndex + 1) % len(m.jqOptions)
			}
			m.suggestedValue = m.jqOptions[m.currentIndex]
			return m, nil

		case "enter":
			if m.showingPreview {
				m.inputs[0].SetValue(m.suggestedValue)
				m.showingPreview = false
				m.suggestedValue = ""
				m.runJqFilter()
				return m, nil
			}
			m.jsonObj, _ = jsonStrToInterface(m.jsonInput)
			return m, tea.Quit

		case "up":
			m.viewport.LineUp(1)
			return m, nil

		case "down":
			m.viewport.LineDown(1)
			return m, nil

		case "pageup":
			m.viewport.ViewUp()
			return m, nil
		case "pagedown":
			m.viewport.ViewDown()
			return m, nil

		default:
			if m.showingPreview {
				m.showingPreview = false
				m.suggestedValue = ""
				return m, nil
			}
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	// Evaluate jq filter on input change
	m.runJqFilter()

	return m, cmd
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func jsonStrToInterface(jsonStr string) (interface{}, error) {
	var jsonData interface{}
	err := json.Unmarshal([]byte(jsonStr), &jsonData)
	if err != nil {
		return nil, fmt.Errorf("Invalid JSON input: %s", err)
	}
	return jsonData, nil
}

func (m *model) runJqFilter() {
	query, err := gojq.Parse(m.inputs[0].Value())
	if err != nil {
		m.jqOutput = fmt.Sprintf("Invalid jq query: %s\n\nLast valid output:\n%s", err, m.lastOutput)
		m.updateViewportContent()
		return
	}

	var jsonData interface{}
	err = json.Unmarshal([]byte(m.jsonInput), &jsonData)
	if err != nil {
		m.jqOutput = fmt.Sprintf("Invalid JSON input: %s\n\nLast valid output:\n%s", err, m.lastOutput)
		m.updateViewportContent()
		return
	}

	iter := query.Run(jsonData)
	var result []string
	isNull := true
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			m.jqOutput = fmt.Sprintf("Error executing jq query: %s\n\nLast valid output:\n%s", err, m.lastOutput)
			m.updateViewportContent()
			return
		}
		output, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			m.jqOutput = fmt.Sprintf("Error formatting output: %s\n\nLast valid output:\n%s", err, m.lastOutput)
			m.updateViewportContent()
			return
		}
		if string(output) != "null" {
			isNull = false
			result = append(result, string(output))
		}
	}

	if isNull {
		m.jqOutput = fmt.Sprintf("Query result is null\n\nLast valid output:\n%s", m.lastOutput)
		m.updateViewportContent()
		return
	}

	m.jqOutput = strings.Join(result, "\n")
	m.lastOutput = m.jqOutput
	m.updateViewportContent()
}

func (m *model) updateViewportContent() {
	prettyOutput, err := codec.PrettyFormat(m.jqOutput, codec.JSON, false)
	if err != nil {
		m.viewport.SetContent(fmt.Sprintf("Error formatting output: %s", err))
		return
	}
	m.viewport.SetContent(outputStyle.Render(prettyOutput))
}

func (m model) View() string {
	var b strings.Builder

	for i := range m.inputs {
		if m.showingPreview && m.suggestedValue != "" {
			b.WriteString(m.inputs[i].View() + previewStyle.Render(m.suggestedValue))
		} else {
			b.WriteString(m.inputs[i].View())
		}
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	b.WriteString("\n")
	b.WriteString(m.viewport.View())

	return b.String()
}

func printOutput(m model) {
	s := m.inputs[0].Value()
	fmt.Printf("\033[32m%s\033[0m\n", s)
	o, err := codec.PrettyFormat(m.jqOutput, codec.JSON, false)
	if err != nil {
		fmt.Println("Error formatting output:", err)
		os.Exit(1)
	}
	fmt.Println(o)
	os.Exit(0)
}

func Interact(s string) {
	m, err := tea.NewProgram(newModel(s), tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	printOutput(m.(model))
}
