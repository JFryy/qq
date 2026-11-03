package cbor

import (
	"reflect"
	"testing"
)

func TestCodecUnmarshal(t *testing.T) {
	codec := &Codec{}
	tests := []struct {
		name string
		data []byte
		want any
	}{
		{"positive integer", []byte{0x18, 0x2a}, int(42)},
		{"negative integer", []byte{0x24}, int(-5)},
		{"float", []byte{0xfa, 0x3f, 0xc0, 0x00, 0x00}, float64(1.5)},
		{"true", []byte{0xf5}, true},
		{"false", []byte{0xf4}, false},
		{"null", []byte{0xf6}, nil},
		{"text", []byte{0x62, 'h', 'i'}, "hi"},
		{
			"nested map and array",
			[]byte{0xa1, 0x65, 'i', 't', 'e', 'm', 's', 0x82, 0x01, 0xa1, 0x62, 'o', 'k', 0xf5},
			map[string]any{"items": []any{int(1), map[string]any{"ok": true}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got any
			if err := codec.Unmarshal(tt.data, &got); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unmarshal() = %#v (%T), want %#v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}

func TestCodecMarshalRoundTrip(t *testing.T) {
	codec := &Codec{}
	tests := []struct {
		name  string
		input any
		want  any
	}{
		{"map", map[string]any{"name": "alice", "active": true, "count": 7}, map[string]any{"name": "alice", "active": true, "count": int(7)}},
		{"nested values", []any{map[string]any{"values": []any{int(-2), float64(1.5), nil}}, false}, []any{map[string]any{"values": []any{int(-2), float64(1.5), nil}}, false}},
		{"nil", nil, nil},
		{"empty array", []any{}, []any{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := codec.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if len(encoded) == 0 {
				t.Fatal("Marshal() returned empty data")
			}
			var got any
			if err := codec.Unmarshal(encoded, &got); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("round trip = %#v (%T), want %#v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}

func TestCodecUnmarshalTypedDestination(t *testing.T) {
	var got struct {
		Count int `cbor:"count"`
	}
	if err := (&Codec{}).Unmarshal([]byte{0xa1, 0x65, 'c', 'o', 'u', 'n', 't', 0x18, 0x2a}, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got.Count != 42 {
		t.Errorf("Count = %d, want 42", got.Count)
	}
}

func TestCodecUnmarshalInvalidInput(t *testing.T) {
	codec := &Codec{}
	tests := []struct {
		name string
		data []byte
	}{
		{"nil", nil},
		{"empty", []byte{}},
		{"truncated integer", []byte{0x18}},
		{"truncated map", []byte{0xa1, 0x61, 'a'}},
		{"unexpected break", []byte{0xff}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got any
			if err := codec.Unmarshal(tt.data, &got); err == nil {
				t.Errorf("Unmarshal(%x) succeeded, want error", tt.data)
			}
		})
	}
}
