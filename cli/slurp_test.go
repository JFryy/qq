package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/JFryy/qq/codec"
	"github.com/itchyny/gojq"
)

func TestSlurpInputs_MultipleJSON(t *testing.T) {
	input := []byte(`{"id":1}
{"id":2}
{"id":3}`)

	result, err := slurpInputs(input, codec.JSON)
	if err != nil {
		t.Fatalf("slurpInputs failed: %v", err)
	}

	arr, ok := result.([]any)
	if !ok {
		t.Fatalf("expected array, got %T", result)
	}

	if len(arr) != 3 {
		t.Errorf("expected 3 items, got %d", len(arr))
	}
}

func TestSlurpInputs_JSONL(t *testing.T) {
	input := []byte(`{"id":1}
{"id":2}`)

	result, err := slurpInputs(input, codec.JSONL)
	if err != nil {
		t.Fatalf("slurpInputs failed: %v", err)
	}

	arr, ok := result.([]any)
	if !ok {
		t.Fatalf("expected array, got %T", result)
	}

	if len(arr) != 2 {
		t.Errorf("expected 2 items, got %d", len(arr))
	}
}

func TestSlurpInputs_YAML_MultiDoc(t *testing.T) {
	input := []byte(`name: Alice
---
name: Bob`)

	result, err := slurpInputs(input, codec.YAML)
	if err != nil {
		t.Fatalf("slurpInputs failed: %v", err)
	}

	arr, ok := result.([]any)
	if !ok {
		t.Fatalf("expected array, got %T", result)
	}

	if len(arr) != 2 {
		t.Errorf("expected 2 items, got %d", len(arr))
	}
}

func TestExecuteQuery_WithSlurp(t *testing.T) {
	input := []byte(`{"id":1}
{"id":2}
{"id":3}`)

	data, err := slurpInputs(input, codec.JSON)
	if err != nil {
		t.Fatalf("slurpInputs failed: %v", err)
	}

	query, err := gojq.Parse("length")
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exitCode := executeQuery(query, data, codec.JSON, false, true, false)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := strings.TrimSpace(buf.String())

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(output, "3") {
		t.Errorf("expected output to contain '3', got: %s", output)
	}
}

func TestExecuteQuery_ExitStatus_True(t *testing.T) {
	query, err := gojq.Parse(".")
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exitCode := executeQuery(query, true, codec.JSON, false, true, true)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if exitCode != 0 {
		t.Errorf("expected exit code 0 for true value, got %d", exitCode)
	}
}

func TestExecuteQuery_ExitStatus_False(t *testing.T) {
	query, err := gojq.Parse(".")
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exitCode := executeQuery(query, false, codec.JSON, false, true, true)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if exitCode != 1 {
		t.Errorf("expected exit code 1 for false value, got %d", exitCode)
	}
}

func TestExecuteQuery_ExitStatus_Null(t *testing.T) {
	query, err := gojq.Parse(".")
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	// Capture stdout - need to handle nil value specially
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Use a query that produces null without error
	query, _ = gojq.Parse(".nonexistent")
	data := map[string]any{}
	exitCode := executeQuery(query, data, codec.JSON, false, true, true)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if exitCode != 1 {
		t.Errorf("expected exit code 1 for null value, got %d", exitCode)
	}
}

func TestExecuteQuery_ExitStatus_NoOutput(t *testing.T) {
	query, err := gojq.Parse("select(. > 10)")
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exitCode := executeQuery(query, 5, codec.JSON, false, true, true)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if exitCode != 4 {
		t.Errorf("expected exit code 4 for no output, got %d", exitCode)
	}
}
