package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/JFryy/qq/codec"
	"github.com/itchyny/gojq"
)

// runPipeline is an E2E helper that parses input, executes a JQ expression,
// and captures the serialized output printed by executeQuery.
func runPipeline(t *testing.T, input []byte, inputFormat codec.EncodingType, outputFormat codec.EncodingType, queryStr string, preserveKeyOrder bool, noAutoConvert bool) (string, error) {
	codec.SetPreserveKeyOrder(preserveKeyOrder)
	codec.SetDisableAutoConvert(noAutoConvert)
	defer func() {
		codec.SetPreserveKeyOrder(false)
		codec.SetDisableAutoConvert(false)
	}()

	var data any
	if err := codec.Unmarshal(input, inputFormat, &data); err != nil {
		return "", fmt.Errorf("unmarshal error: %w", err)
	}

	q, err := gojq.Parse(queryStr)
	if err != nil {
		return "", fmt.Errorf("query parse error: %w", err)
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exitCode := executeQuery(q, data, outputFormat, false, true, false)

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	if exitCode != 0 {
		return "", fmt.Errorf("exit code: %d, output: %s", exitCode, buf.String())
	}

	return buf.String(), nil
}

func TestE2E_CSV_JQExpressions(t *testing.T) {
	csvInput := []byte("a,b,c\n1,2,3\n")

	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "identity query",
			query:    ".",
			expected: `[{"a":1,"b":2,"c":3}]`,
		},
		{
			name:     "array iteration",
			query:    ".[]",
			expected: `{"a":1,"b":2,"c":3}`,
		},
		{
			name:     "length calculation",
			query:    "length",
			expected: `1`,
		},
		{
			name:     "first element",
			query:    "first",
			expected: `{"a":1,"b":2,"c":3}`,
		},
		{
			name:     "array index",
			query:    ".[0]",
			expected: `{"a":1,"b":2,"c":3}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := runPipeline(t, csvInput, codec.CSV, codec.JSON, tt.query, false, false)
			if err != nil {
				t.Fatalf("pipeline failed: %v", err)
			}
			trimmedGot := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(got), "\n", ""), " ", "")
			trimmedExp := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(tt.expected), "\n", ""), " ", "")
			if trimmedGot != trimmedExp {
				t.Errorf("got: %q, want: %q", trimmedGot, trimmedExp)
			}
		})
	}
}

func TestE2E_KeyPreservation_Formats(t *testing.T) {
	csvInput := []byte("zebra,apple,mango\n1,2,3\n")

	tests := []struct {
		name         string
		outputFormat codec.EncodingType
		expected     string
	}{
		{
			name:         "JSON output",
			outputFormat: codec.JSON,
			expected: `[
  {
    "zebra": 1,
    "apple": 2,
    "mango": 3
  }
]`,
		},
		{
			name:         "JSONL output",
			outputFormat: codec.JSONL,
			expected:     `{"zebra":1,"apple":2,"mango":3}`,
		},
		{
			name:         "CSV output",
			outputFormat: codec.CSV,
			expected:     "zebra,apple,mango\n1,2,3\n",
		},
		{
			name:         "TSV output",
			outputFormat: codec.TSV,
			expected:     "zebra\tapple\tmango\n1\t2\t3\n",
		},
		{
			name:         "YAML output",
			outputFormat: codec.YAML,
			expected: `zebra: 1
apple: 2
mango: 3
`,
		},
		{
			name:         "XML output",
			outputFormat: codec.XML,
			expected: `<root>
  <zebra>1</zebra>
  <apple>2</apple>
  <mango>3</mango>
</root>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := runPipeline(t, csvInput, codec.CSV, tt.outputFormat, ".", true, false)
			if err != nil {
				t.Fatalf("pipeline failed: %v", err)
			}
			got = strings.TrimPrefix(got, "---\n")
			got = strings.TrimPrefix(got, "---\r\n")
			trimmedGot := strings.TrimSpace(got)
			trimmedExp := strings.TrimSpace(tt.expected)
			if trimmedGot != trimmedExp {
				t.Errorf("got:\n%s\nwant:\n%s", trimmedGot, trimmedExp)
			}
		})
	}
}

func TestE2E_DisableAutoConvert(t *testing.T) {
	csvInput := []byte("status,count\nT,1\nF,0\ntrue,1.5\n")

	// 1. With auto-convert enabled (default)
	t.Run("default conversion", func(t *testing.T) {
		got, err := runPipeline(t, csvInput, codec.CSV, codec.JSON, ".[]", false, false)
		if err != nil {
			t.Fatalf("pipeline failed: %v", err)
		}
		expected := `{"count":1,"status":true}
{"count":0,"status":false}
{"count":1.5,"status":true}`
		trimmedGot := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(got), "\n", ""), " ", "")
		trimmedExp := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(expected), "\n", ""), " ", "")
		if trimmedGot != trimmedExp {
			t.Errorf("got:\n%s\nwant:\n%s", trimmedGot, trimmedExp)
		}
	})

	// 2. With --no-auto-convert (coercion disabled)
	t.Run("no auto convert", func(t *testing.T) {
		got, err := runPipeline(t, csvInput, codec.CSV, codec.JSON, ".[]", false, true)
		if err != nil {
			t.Fatalf("pipeline failed: %v", err)
		}
		expected := `{"count":"1","status":"T"}
{"count":"0","status":"F"}
{"count":"1.5","status":"true"}`
		trimmedGot := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(got), "\n", ""), " ", "")
		trimmedExp := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(expected), "\n", ""), " ", "")
		if trimmedGot != trimmedExp {
			t.Errorf("got:\n%s\nwant:\n%s", trimmedGot, trimmedExp)
		}
	})
}

func TestE2E_EdgeCases(t *testing.T) {
	// Unicode and special characters
	unicodeInput := []byte("name,cél\nJosé,value\n")
	t.Run("unicode support", func(t *testing.T) {
		got, err := runPipeline(t, unicodeInput, codec.CSV, codec.JSON, ".", false, false)
		if err != nil {
			t.Fatalf("pipeline failed: %v", err)
		}
		expected := `[{"cél":"value","name":"José"}]`
		trimmedGot := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(got), "\n", ""), " ", "")
		if trimmedGot != expected {
			t.Errorf("got: %q, want: %q", trimmedGot, expected)
		}
	})

	// Header only (no rows)
	headerOnlyInput := []byte("a,b,c\n")
	t.Run("header only", func(t *testing.T) {
		got, err := runPipeline(t, headerOnlyInput, codec.CSV, codec.JSON, ".", false, false)
		if err != nil {
			t.Fatalf("pipeline failed: %v", err)
		}
		trimmedGot := strings.TrimSpace(got)
		if trimmedGot != "null" && trimmedGot != "[]" && trimmedGot != "" {
			t.Errorf("expected empty/null result, got %q", trimmedGot)
		}
	})
}
