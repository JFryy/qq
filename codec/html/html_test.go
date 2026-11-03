package html

import (
	"reflect"
	"testing"
)

func TestCodecUnmarshal(t *testing.T) {
	codec := &Codec{}
	tests := []struct {
		name  string
		input string
		want  any
	}{
		{
			name:  "nested elements and repeated siblings",
			input: `<html><body><section><item>first</item><item>second</item></section><footer>End</footer></body></html>`,
			want: map[string]any{"html": map[string]any{"body": map[string]any{
				"section": map[string]any{"item": []any{"first", "second"}},
				"footer":  "End",
			}}},
		},
		{
			name:  "attributes, entities and text",
			input: `<html><body><p title="A &amp; B">Hello &lt;world&gt;</p><aside>Other</aside></body></html>`,
			want: map[string]any{"html": map[string]any{"body": map[string]any{
				"p":     map[string]any{"@title": "A & B", "#text": "Hello <world>"},
				"aside": "Other",
			}}},
		},
		{
			name:  "numbers and booleans remain text",
			input: `<html><body><count>42</count><enabled>true</enabled><other>false</other></body></html>`,
			want: map[string]any{"html": map[string]any{"body": map[string]any{
				"count": "42", "enabled": "true", "other": "false",
			}}},
		},
		{
			name:  "comments",
			input: `<html><body><p>one</p><!-- note &amp; more --><p>two</p></body></html>`,
			want: map[string]any{"html": map[string]any{"body": map[string]any{
				"p": []any{"one", "two"}, "#comment": "note & more",
			}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got any
			if err := codec.Unmarshal([]byte(tt.input), &got); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unmarshal() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestCodecHTMLToMap(t *testing.T) {
	codec := &Codec{}
	tests := []struct {
		name  string
		input string
		want  map[string]any
	}{
		{
			name:  "unicode escape in text and attribute",
			input: `<html><body><p title="\u0041">\u263A</p><aside>ok</aside></body></html>`,
			want: map[string]any{"html": map[string]any{"body": map[string]any{
				"p": map[string]any{"@title": "A", "#text": "☺"}, "aside": "ok",
			}}},
		},
		{
			name:  "multiple comments",
			input: `<html><body><!-- first --><!-- second --><p>ok</p></body></html>`,
			want: map[string]any{"html": map[string]any{"body": map[string]any{
				"#comment": []string{"first", "second"}, "p": "ok",
			}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := codec.HTMLToMap([]byte(tt.input))
			if err != nil {
				t.Fatalf("HTMLToMap() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTMLToMap() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestCodecUnmarshalInvalidDestination(t *testing.T) {
	codec := &Codec{}
	input := []byte(`<html><body><p>value</p><aside>other</aside></body></html>`)
	for _, tt := range []struct {
		name        string
		destination any
	}{
		{name: "nil destination", destination: nil},
		{name: "non-pointer destination", destination: map[string]any{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := codec.Unmarshal(input, tt.destination); err == nil {
				t.Error("Unmarshal() error = nil, want invalid destination error")
			}
		})
	}
}
