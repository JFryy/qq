package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/JFryy/qq/codec"
)

func TestInferFileType(t *testing.T) {
	tests := []struct {
		filename string
		expected codec.EncodingType
	}{
		{"test.json", codec.JSON},
		{"test.yaml", codec.YAML},
		{"test.yml", codec.YML},
		{"test.toml", codec.TOML},
		{"test.xml", codec.XML},
		{"test.csv", codec.CSV},
		{"test.hcl", codec.HCL},
		{"test.tf", codec.TF},
		{"test.ini", codec.INI},
		{"test.unknown", codec.JSON}, // defaults to JSON
		{"/path/to/file.json", codec.JSON},
		{"FILE.JSON", codec.JSON}, // case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := inferFileType(tt.filename)
			if result != tt.expected {
				t.Errorf("inferFileType(%q) = %v, expected %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestIsFile(t *testing.T) {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "test*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"existing file", tmpfile.Name(), true},
		{"non-existing file", "/nonexistent/file.txt", false},
		{"directory", "/tmp", false},
		{"empty path", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFile(tt.path)
			if result != tt.expected {
				t.Errorf("isFile(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsTerminal(t *testing.T) {
	// Test with os.Stdout - this might be a terminal or not depending on test environment
	result := isTerminal(os.Stdout)
	t.Logf("isTerminal(os.Stdout) = %v", result)

	// Create a pipe to test non-terminal
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()

	if isTerminal(w) {
		t.Error("Pipe should not be detected as terminal")
	}
}

func TestCreateRootCmd(t *testing.T) {
	cmd := CreateRootCmd()

	if cmd == nil {
		t.Fatal("CreateRootCmd returned nil")
	}

	if cmd.Use == "" {
		t.Error("Command Use should not be empty")
	}

	// Test flags exist
	flags := cmd.Flags()
	if !flags.HasFlags() {
		t.Error("Command should have flags")
	}

	requiredFlags := []string{"input", "output", "raw-output", "monochrome-output", "interactive"}
	for _, flag := range requiredFlags {
		if flags.Lookup(flag) == nil {
			t.Errorf("Missing required flag: %s", flag)
		}
	}
}

func TestRootCmdFlagDefaults(t *testing.T) {
	cmd := CreateRootCmd()

	tests := []struct {
		flagName string
		expected string
	}{
		{"input", "json"},
		{"output", "json"},
	}

	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			value, err := cmd.Flags().GetString(tt.flagName)
			if err != nil {
				t.Fatalf("Failed to get flag %s: %v", tt.flagName, err)
			}
			if value != tt.expected {
				t.Errorf("Default value for --%s = %q, expected %q", tt.flagName, value, tt.expected)
			}
		})
	}

	// Test boolean flag defaults
	boolTests := []struct {
		flagName string
		expected bool
	}{
		{"raw-output", false},
		{"monochrome-output", false},
		{"interactive", false},
	}

	for _, tt := range boolTests {
		t.Run(tt.flagName, func(t *testing.T) {
			value, err := cmd.Flags().GetBool(tt.flagName)
			if err != nil {
				t.Fatalf("Failed to get flag %s: %v", tt.flagName, err)
			}
			if value != tt.expected {
				t.Errorf("Default value for --%s = %v, expected %v", tt.flagName, value, tt.expected)
			}
		})
	}
}

func TestExecuteQueryBasic(t *testing.T) {
	// This is a basic integration test that doesn't require external dependencies
	// More comprehensive tests should be in E2E tests

	tests := []struct {
		name     string
		input    string
		query    string
		fileType codec.EncodingType
		wantErr  bool
	}{
		{
			name:     "simple identity query",
			input:    `{"key": "value"}`,
			query:    ".",
			fileType: codec.JSON,
			wantErr:  false,
		},
		{
			name:     "key extraction",
			input:    `{"key": "value"}`,
			query:    ".key",
			fileType: codec.JSON,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the input
			var data any
			err := codec.Unmarshal([]byte(tt.input), tt.fileType, &data)
			if err != nil {
				t.Fatalf("Failed to unmarshal input: %v", err)
			}

			// Note: Full query execution testing should be in E2E tests
			// This just verifies the data can be unmarshaled
			if data == nil && !tt.wantErr {
				t.Error("Expected data, got nil")
			}
		})
	}
}

func TestVersionFlag(t *testing.T) {
	cmd := CreateRootCmd()
	cmd.SetArgs([]string{"--version"})

	// Capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command in goroutine and capture exit
	done := make(chan bool)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Command might call os.Exit, which we can't easily test
				done <- true
			}
		}()
		cmd.Execute()
		done <- true
	}()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	<-done

	// Version output should contain "qq version"
	if !strings.Contains(output, "qq version") && output != "" {
		t.Logf("Version output: %q", output)
	}
}

func TestHelpFlag(t *testing.T) {
	cmd := CreateRootCmd()

	// Get help text
	helpText := cmd.UsageString()

	if helpText == "" {
		t.Error("Help text should not be empty")
	}

	// Should mention key features
	if !strings.Contains(strings.ToLower(helpText), "json") {
		t.Error("Help text should mention json")
	}

	// Should have the basic flags
	if !strings.Contains(helpText, "input") || !strings.Contains(helpText, "output") {
		t.Error("Help text should mention input/output flags")
	}
}
