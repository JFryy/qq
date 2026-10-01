package codec

import "testing"

func TestDetect(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected EncodingType
	}{
		{"json object", `{"name":"Alice","age":30}`, JSON},
		{"json array", `[1,2,3]`, JSON},
		{"json with leading space", "  \n{\"a\":1}", JSON},
		{"yaml map", "name: Alice\nage: 30\n", YAML},
		{"yaml doc marker", "---\nname: Bob\n", YAML},
		{"yaml list", "- one\n- two\n", YAML},
		{"toml", "title = \"hi\"\n[owner]\nname = \"x\"\n", TOML},
		{"xml", "<root><a>1</a></root>", XML},
		{"xml with declaration", "<?xml version=\"1.0\"?><note><to>x</to></note>", XML},
		{"html doctype", "<!DOCTYPE html><html><body><p>hi</p></body></html>", HTML},
		{"html tag", "<html><head></head></html>", HTML},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Detect([]byte(tt.input))
			if err != nil {
				t.Fatalf("Detect(%q) returned error: %v", tt.input, err)
			}
			if got != tt.expected {
				t.Errorf("Detect(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDetectUndetectable(t *testing.T) {
	// Inputs that Detect should refuse rather than guess at.
	cases := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"whitespace only", "   \n\t"},
		{"plain scalar", "just some words"},
		{"bare number", "42"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := Detect([]byte(tt.input)); err == nil {
				t.Errorf("Detect(%q) = %v, expected an error", tt.input, got)
			}
		})
	}
}
