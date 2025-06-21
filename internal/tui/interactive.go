package tui

import (
	"fmt"
	"github.com/goccy/go-json"
	"os"
	"strings"

	"github.com/JFryy/qq/codec"
	"github.com/charmbracelet/bubbles/textarea"
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
	headerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	legendStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	borderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
)

type model struct {
	textArea       textarea.Model
	jsonInput      string
	jqOutput       string
	lastOutput     string
	currentIndex   int
	showingPreview bool
	jqOptions      []string
	suggestedValue string
	jsonObj        any
	viewport       viewport.Model
	gracefulExit   bool
}

func newModel(data string) model {
	m := model{
		viewport: viewport.New(0, 0),
	}

	t := textarea.New()
	t.Cursor.Style = cursorStyle
	t.Placeholder = "Enter jq filter"
	t.SetValue(".")
	t.Focus()
	t.SetWidth(80)
	t.SetHeight(3)
	t.CharLimit = 0
	t.ShowLineNumbers = false
	t.KeyMap.InsertNewline.SetEnabled(true)
	m.textArea = t
	m.jsonInput = string(data)
	m.jqOptions = generateJqOptions(m.jsonInput)

	m.runJqFilter()
	m.jsonObj, _ = jsonStrToInterface(m.jsonInput)

	return m
}

func generateJqOptions(jsonStr string) []string {
	var jsonData any
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

func extractPaths(data any, prefix string, options map[string]struct{}) {
	switch v := data.(type) {
	case map[string]any:
		for key, value := range v {
			newPrefix := prefix + "." + key
			options[newPrefix] = struct{}{}
			extractPaths(value, newPrefix, options)
		}
	case []any:
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
		headerHeight := 6 // Space for header, textarea, and legend
		footerHeight := 1
		availableHeight := msg.Height - headerHeight - footerHeight
		m.viewport.Width = msg.Width
		m.viewport.Height = availableHeight
		m.textArea.SetWidth(msg.Width - 4) // Set width with padding
		m.updateViewportContent()
		return m, nil

	case tea.KeyMsg:
		switch {
		case msg.String() == "ctrl+c" || msg.String() == "esc":
			// Try to run current query if it's valid, otherwise just exit
			if m.isValidQuery() {
				m.gracefulExit = true
				m.jsonObj, _ = jsonStrToInterface(m.jsonInput)
			}
			return m, tea.Quit

		// Suggest next jq option
		case msg.String() == "tab":
			if !m.showingPreview {
				m.showingPreview = true
				m.currentIndex = 0
			} else {
				m.currentIndex = (m.currentIndex + 1) % len(m.jqOptions)
			}
			m.suggestedValue = m.jqOptions[m.currentIndex]
			return m, nil

		case msg.String() == "enter":
			if m.showingPreview {
				m.textArea.SetValue(m.suggestedValue)
				m.showingPreview = false
				m.suggestedValue = ""
				m.runJqFilter()
				return m, nil
			}
			// Let the textarea handle the enter key for newlines - don't intercept it
			break

		case msg.String() == "up":
			m.viewport.LineUp(1)
			return m, nil

		case msg.String() == "down":
			m.viewport.LineDown(1)
			return m, nil

		case msg.String() == "pageup":
			m.viewport.ViewUp()
			return m, nil
		case msg.String() == "pagedown":
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
	var cmd tea.Cmd
	m.textArea, cmd = m.textArea.Update(msg)
	return cmd
}

func (m model) isValidQuery() bool {
	query := strings.TrimSpace(m.textArea.Value())
	if query == "" {
		return false
	}
	_, err := gojq.Parse(query)
	return err == nil
}

func jsonStrToInterface(jsonStr string) (any, error) {
	var jsonData any
	err := json.Unmarshal([]byte(jsonStr), &jsonData)
	if err != nil {
		return nil, fmt.Errorf("Invalid JSON input: %s", err)
	}
	return jsonData, nil
}

func (m *model) runJqFilter() {
	query, err := gojq.Parse(m.textArea.Value())
	if err != nil {
		m.jqOutput = fmt.Sprintf("Invalid jq query: %s\n\nLast valid output:\n%s", err, m.lastOutput)
		m.updateViewportContent()
		return
	}

	var jsonData any
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
	prettyOutput, err := codec.PrettyFormat(m.jqOutput, codec.JSON, false, false)
	if err != nil {
		m.viewport.SetContent(fmt.Sprintf("Error formatting output: %s", err))
		return
	}
	m.viewport.SetContent(outputStyle.Render(prettyOutput))
}

func (m model) View() string {
	var b strings.Builder

	// Header
	b.WriteString(headerStyle.Render("🔍 qq Interactive Mode"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", 50)))
	b.WriteString("\n")

	// Text area
	b.WriteString(m.textArea.View())

	// Preview suggestion
	if m.showingPreview && m.suggestedValue != "" {
		b.WriteString("\n")
		b.WriteString(previewStyle.Render("💡 Suggestion: " + m.suggestedValue))
	}

	// Legend
	b.WriteString("\n")
	b.WriteString(legendStyle.Render("Tab: autocomplete | Enter: accept/newline | Ctrl+C/Esc: execute & exit | ↑↓: scroll"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", 50)))
	b.WriteString("\n")

	// Output viewport
	b.WriteString(m.viewport.View())

	return b.String()
}

func printOutput(m model) {
	if m.gracefulExit {
		// Graceful exit with formatted output
		s := m.textArea.Value()
		fmt.Printf("\033[36m# Query: %s\033[0m\n", s)
		o, err := codec.PrettyFormat(m.jqOutput, codec.JSON, false, false)
		if err != nil {
			fmt.Printf("\033[31mError formatting output: %s\033[0m\n", err)
			os.Exit(1)
		}
		fmt.Println(o)
		os.Exit(0)
	} else {
		// Abrupt exit
		fmt.Println("\033[33mExited without executing query\033[0m")
		os.Exit(0)
	}
}

func Interact(s string) {
	m, err := tea.NewProgram(newModel(s), tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	printOutput(m.(model))
}
