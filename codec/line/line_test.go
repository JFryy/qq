package line

import (
	"reflect"
	"strings"
	"testing"
)

func TestCodecUnmarshalValues(t *testing.T) {
	codec := &Codec{}
	tests := []struct {
		name  string
		input string
		want  []any
	}{
		{
			name:  "plain lines",
			input: "alpha\nbeta\n",
			want:  []any{"alpha", "beta"},
		},
		{
			name:  "whitespace around lines",
			input: "  alpha  \n\t beta \t",
			want:  []any{"alpha", "beta"},
		},
		{
			name:  "numbers and booleans",
			input: "42\n-7\n3.5\ntrue\nfalse",
			want:  []any{float64(42), float64(-7), 3.5, true, false},
		},
		{
			name:  "text that resembles structured values",
			input: "null\n[1,2]\n{\"a\":1}",
			want:  []any{"null", "[1,2]", "{\"a\":1}"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []any
			if err := codec.Unmarshal([]byte(tt.input), &got); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unmarshal = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestCodecUnmarshalTypedSlice(t *testing.T) {
	codec := &Codec{}
	var got []int
	if err := codec.Unmarshal([]byte("1\n2\n3"), &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if want := []int{1, 2, 3}; !reflect.DeepEqual(got, want) {
		t.Errorf("Unmarshal = %v, want %v", got, want)
	}
}

func TestCodecUnmarshalInvalidDestination(t *testing.T) {
	codec := &Codec{}
	tests := []struct {
		name string
		dest any
		want string
	}{
		{name: "nil destination", want: "non-nil pointer"},
		{name: "non-pointer", dest: []string{}, want: "non-nil pointer"},
		{name: "nil pointer", dest: (*[]string)(nil), want: "non-nil pointer"},
		{name: "incompatible element", dest: new([]int), want: "unmarshaling JSON"},
		{name: "incompatible container", dest: new(map[string]any), want: "unmarshaling JSON"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := codec.Unmarshal([]byte("hello"), tt.dest)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Unmarshal error = %v, want error containing %q", err, tt.want)
			}
		})
	}
}
