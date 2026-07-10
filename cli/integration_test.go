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
func runPipeline(t *testing.T, input []byte, inputFormat codec.EncodingType, outputFormat codec.EncodingType, queryStr string, preserveKeyOrder bool, noAutoConvert bool, rawOut bool) (string, error) {
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

	exitCode := executeQuery(q, data, outputFormat, rawOut, true, false)

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	if exitCode != 0 {
		return "", fmt.Errorf("exit code: %d, output: %s", exitCode, buf.String())
	}

	return buf.String(), nil
}

func TestE2E_Formats_JQExpressions(t *testing.T) {
	// Test suite verifying each input format with various JQ expressions
	tests := []struct {
		name        string
		input       []byte
		inputFormat codec.EncodingType
		query       string
		expected    string
	}{
		// CSV Input
		{
			name:        "CSV identity",
			input:       []byte("a,b,c\n1,2,3\n"),
			inputFormat: codec.CSV,
			query:       ".",
			expected:    `[{"a":1,"b":2,"c":3}]`,
		},
		{
			name:        "CSV array iteration",
			input:       []byte("a,b,c\n1,2,3\n"),
			inputFormat: codec.CSV,
			query:       ".[]",
			expected:    `{"a":1,"b":2,"c":3}`,
		},
		{
			name:        "CSV length",
			input:       []byte("a,b,c\n1,2,3\n"),
			inputFormat: codec.CSV,
			query:       "length",
			expected:    `1`,
		},
		// TSV Input
		{
			name:        "TSV array index",
			input:       []byte("a\tb\n1\t2\n"),
			inputFormat: codec.TSV,
			query:       ".[0]",
			expected:    `{"a":1,"b":2}`,
		},
		// JSON Input
		{
			name:        "JSON array iteration",
			input:       []byte(`[{"x": 10}, {"x": 20}]`),
			inputFormat: codec.JSON,
			query:       ".[] | .x",
			expected:    "10\n20",
		},
		{
			name:        "JSON length",
			input:       []byte(`[{"x": 10}, {"x": 20}]`),
			inputFormat: codec.JSON,
			query:       "length",
			expected:    "2",
		},
		// YAML Input
		{
			name:        "YAML array iteration",
			input:       []byte("- 100\n- 200\n"),
			inputFormat: codec.YAML,
			query:       ".[]",
			expected:    "100\n200",
		},
		// XML Input
		{
			name:        "XML field selection",
			input:       []byte("<root><item>10</item><item>20</item></root>"),
			inputFormat: codec.XML,
			query:       ".root.item[]",
			expected:    "10\n20", // Note:mxj unmarshals elements as string/numbers
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := runPipeline(t, tt.input, tt.inputFormat, codec.JSON, tt.query, false, false, false)
			if err != nil {
				t.Fatalf("pipeline failed: %v", err)
			}
			trimmedGot := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(got), "\n", ""), " ", "")
			trimmedExp := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(tt.expected), "\n", ""), " ", "")
			// For non-JSON output assertions like plain numbers (e.g. 10\n20) we compare directly
			if !strings.Contains(tt.expected, "{") && !strings.Contains(tt.expected, "[") {
				trimmedGot = strings.TrimSpace(got)
				trimmedExp = strings.TrimSpace(tt.expected)
			}
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
			got, err := runPipeline(t, csvInput, codec.CSV, tt.outputFormat, ".", true, false, false)
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
	// 1. CSV Auto-Coercion
	t.Run("CSV default vs no-auto-convert", func(t *testing.T) {
		csvInput := []byte("status,count\nT,1\nF,0\ntrue,1.5\n")
		
		gotDefault, _ := runPipeline(t, csvInput, codec.CSV, codec.JSON, ".[]", false, false, false)
		expDefault := `{"count":1,"status":true}{"count":0,"status":false}{"count":1.5,"status":true}`
		if strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(gotDefault), "\n", ""), " ", "") != expDefault {
			t.Errorf("default got: %q, want: %q", gotDefault, expDefault)
		}

		gotDisabled, _ := runPipeline(t, csvInput, codec.CSV, codec.JSON, ".[]", false, true, false)
		expDisabled := `{"count":"1","status":"T"}{"count":"0","status":"F"}{"count":"1.5","status":"true"}`
		if strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(gotDisabled), "\n", ""), " ", "") != expDisabled {
			t.Errorf("disabled got: %q, want: %q", gotDisabled, expDisabled)
		}
	})

	// 2. XML Auto-Coercion
	t.Run("XML default vs no-auto-convert", func(t *testing.T) {
		xmlInput := []byte("<root><status>F</status><count>1</count></root>")

		gotDefault, _ := runPipeline(t, xmlInput, codec.XML, codec.JSON, ".", false, false, false)
		expDefault := `{"root":{"count":1,"status":false}}`
		if strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(gotDefault), "\n", ""), " ", "") != expDefault {
			t.Errorf("default got: %q, want: %q", gotDefault, expDefault)
		}

		gotDisabled, _ := runPipeline(t, xmlInput, codec.XML, codec.JSON, ".", false, true, false)
		expDisabled := `{"root":{"count":"1","status":"F"}}`
		if strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(gotDisabled), "\n", ""), " ", "") != expDisabled {
			t.Errorf("disabled got: %q, want: %q", gotDisabled, expDisabled)
		}
	})

	// 3. INI Auto-Coercion
	t.Run("INI default vs no-auto-convert", func(t *testing.T) {
		iniInput := []byte("[section]\nstatus=F\ncount=1\n")

		gotDefault, _ := runPipeline(t, iniInput, codec.INI, codec.JSON, ".", false, false, false)
		expDefault := `{"section":{"count":1,"status":false}}`
		if strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(gotDefault), "\n", ""), " ", "") != expDefault {
			t.Errorf("default got: %q, want: %q", gotDefault, expDefault)
		}

		gotDisabled, _ := runPipeline(t, iniInput, codec.INI, codec.JSON, ".", false, true, false)
		expDisabled := `{"section":{"count":"1","status":"F"}}`
		if strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(gotDisabled), "\n", ""), " ", "") != expDisabled {
			t.Errorf("disabled got: %q, want: %q", gotDisabled, expDisabled)
		}
	})

	// 4. Gron Auto-Coercion
	t.Run("Gron default vs no-auto-convert", func(t *testing.T) {
		gronInput := []byte("json.status = \"F\";\njson.count = 1;\n")

		gotDefault, _ := runPipeline(t, gronInput, codec.GRON, codec.JSON, ".json", false, false, false)
		expDefault := `{"count":1,"status":false}`
		if strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(gotDefault), "\n", ""), " ", "") != expDefault {
			t.Errorf("default got: %q, want: %q", gotDefault, expDefault)
		}

		gotDisabled, _ := runPipeline(t, gronInput, codec.GRON, codec.JSON, ".json", false, true, false)
		expDisabled := `{"count":"1","status":"F"}`
		if strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(gotDisabled), "\n", ""), " ", "") != expDisabled {
			t.Errorf("disabled got: %q, want: %q", gotDisabled, expDisabled)
		}
	})
}

func TestE2E_FlagCombinations(t *testing.T) {
	// Combination: --no-auto-convert + --preserve-key-order + -r (raw output)
	csvInput := []byte("zebra,apple,mango\nT,2,3\n")
	t.Run("no-auto-convert + preserve-key-order + raw-output", func(t *testing.T) {
		got, err := runPipeline(t, csvInput, codec.CSV, codec.JSON, ".[0].zebra", true, true, true)
		if err != nil {
			t.Fatalf("pipeline failed: %v", err)
		}
		expected := "T\n"
		if got != expected {
			t.Errorf("got: %q, want: %q", got, expected)
		}
	})
}

func TestE2E_EdgeCases(t *testing.T) {
	// Unicode and special characters
	unicodeInput := []byte("name,cél\nJosé,value\n")
	t.Run("unicode support", func(t *testing.T) {
		got, err := runPipeline(t, unicodeInput, codec.CSV, codec.JSON, ".", false, false, false)
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
		got, err := runPipeline(t, headerOnlyInput, codec.CSV, codec.JSON, ".", false, false, false)
		if err != nil {
			t.Fatalf("pipeline failed: %v", err)
		}
		trimmedGot := strings.TrimSpace(got)
		if trimmedGot != "null" && trimmedGot != "[]" && trimmedGot != "" {
			t.Errorf("expected empty/null result, got %q", trimmedGot)
		}
	})

	// Empty input for JSON
	t.Run("empty JSON input", func(t *testing.T) {
		_, err := runPipeline(t, []byte(""), codec.JSON, codec.JSON, ".", false, false, false)
		if err == nil {
			t.Error("expected error for empty JSON input, got nil")
		}
	})
}
