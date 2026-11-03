package ini

import (
	"reflect"
	"strings"
	"testing"
)

func TestCodecUnmarshal(t *testing.T) {
	codec := &Codec{}
	tests := []struct {
		name  string
		input string
		want  map[string]any
	}{
		{
			name: "sections and typed values",
			input: "[server]\nname = api\nport = 8080\nratio = 1.5\nenabled = true\n" +
				"[client]\nretries = -2\nenabled = false\n",
			want: map[string]any{
				"server": map[string]any{"name": "api", "port": 8080, "ratio": 1.5, "enabled": true},
				"client": map[string]any{"retries": -2, "enabled": false},
			},
		},
		{
			name:  "empty input",
			input: "",
			want:  map[string]any{},
		},
		{
			name:  "empty section",
			input: "[empty]\n",
			want:  map[string]any{"empty": map[string]any{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got map[string]any
			if err := codec.Unmarshal([]byte(tt.input), &got); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unmarshal() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestCodecUnmarshalMalformed(t *testing.T) {
	codec := &Codec{}
	for _, input := range []string{"[broken", "[section]\n= value\n"} {
		t.Run(input, func(t *testing.T) {
			var got map[string]any
			if err := codec.Unmarshal([]byte(input), &got); err == nil {
				t.Errorf("Unmarshal(%q) expected an error", input)
			}
		})
	}
}

func TestCodecMarshalRoundTrip(t *testing.T) {
	codec := &Codec{}
	tests := []struct {
		name  string
		input map[string]any
	}{
		{
			name: "multiple sections and types",
			input: map[string]any{
				"server": map[string]any{"name": "api", "port": 8080, "enabled": true, "ratio": 1.5},
				"client": map[string]any{"retries": -2, "enabled": false},
			},
		},
		{
			name:  "empty section",
			input: map[string]any{"empty": map[string]any{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := codec.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			var got map[string]any
			if err := codec.Unmarshal(encoded, &got); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.input) {
				t.Errorf("round trip = %#v, want %#v (encoded %q)", got, tt.input, encoded)
			}
		})
	}
}

func TestCodecMarshalSortedKeys(t *testing.T) {
	encoded, err := (&Codec{}).Marshal(map[string]any{
		"zeta":  map[string]any{"z": "last", "a": "first"},
		"alpha": map[string]any{"key": "value"},
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	output := string(encoded)
	for _, pair := range [][2]string{{"[alpha]", "[zeta]"}, {"a = first", "z = last"}} {
		first, second := strings.Index(output, pair[0]), strings.Index(output, pair[1])
		if first < 0 || second < 0 || first >= second {
			t.Errorf("Marshal() output %q does not put %q before %q", output, pair[0], pair[1])
		}
	}
}

func TestCodecMarshalRejectsNonMap(t *testing.T) {
	codec := &Codec{}
	for _, input := range []any{nil, "text", 42, []any{"item"}} {
		if _, err := codec.Marshal(input); err == nil {
			t.Errorf("Marshal(%#v) expected an error", input)
		}
	}
}
