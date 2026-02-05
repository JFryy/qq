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

func TestExecuteStreamingQuery_BasicObject(t *testing.T) {
	input := `{"name":"Alice","age":30}`
	reader := strings.NewReader(input)
	query, err := gojq.Parse(".")
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	executeStreamingQuery(query, reader, codec.JSON, codec.JSON, false, true)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should contain path-value pairs
	if !strings.Contains(output, `"name"`) {
		t.Errorf("expected output to contain 'name', got: %s", output)
	}
	if !strings.Contains(output, `"Alice"`) {
		t.Errorf("expected output to contain 'Alice', got: %s", output)
	}
}

func TestExecuteStreamingQuery_Array(t *testing.T) {
	input := `[1,2,3]`
	reader := strings.NewReader(input)
	query, err := gojq.Parse(".")
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	executeStreamingQuery(query, reader, codec.JSON, codec.JSON, false, true)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should contain array indices and values
	if !strings.Contains(output, "0") || !strings.Contains(output, "1") || !strings.Contains(output, "2") {
		t.Errorf("expected output to contain array indices")
	}
}

func TestExecuteStreamingQuery_WithFilter(t *testing.T) {
	input := `{"a":1,"b":2,"c":3}`
	reader := strings.NewReader(input)
	// Select only path-value pairs (length == 2), not the closing marker
	query, err := gojq.Parse("select(length == 2)")
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	executeStreamingQuery(query, reader, codec.JSON, codec.JSON, false, true)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should contain path-value pairs but not the closing marker
	lines := strings.Split(strings.TrimSpace(output), "\n")
	nonEmptyLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
		}
	}

	// Each path-value pair spans multiple lines due to pretty printing
	// We should have at least some output
	if nonEmptyLines == 0 {
		t.Errorf("expected some output from filtered stream")
	}
}

func TestExecuteStreamingQuery_NestedStructure(t *testing.T) {
	input := `{"user":{"name":"Bob","id":123}}`
	reader := strings.NewReader(input)
	query, err := gojq.Parse(".")
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	executeStreamingQuery(query, reader, codec.JSON, codec.JSON, false, true)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should contain nested paths
	if !strings.Contains(output, `"user"`) || !strings.Contains(output, `"name"`) {
		t.Errorf("expected output to contain nested path elements")
	}
}

func TestStreamParser(t *testing.T) {
	input := `{"test":123}`
	reader := strings.NewReader(input)

	stream, err := codec.StreamParserCollect(reader, codec.JSON)
	if err != nil {
		t.Fatalf("StreamParser failed: %v", err)
	}

	if len(stream) == 0 {
		t.Errorf("expected non-empty stream")
	}

	// Each element should be a path-value pair or closing marker
	for _, element := range stream {
		arr, ok := element.([]any)
		if !ok {
			t.Errorf("expected stream element to be an array, got %T", element)
		}
		if len(arr) != 1 && len(arr) != 2 {
			t.Errorf("expected array of length 1 or 2, got %d", len(arr))
		}
	}
}
