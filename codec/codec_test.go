package codec

import (
	"testing"
)

func TestGetEncodingType(t *testing.T) {
	tests := []struct {
		input    string
		expected EncodingType
	}{
		{"json", JSON},
		{"yaml", YAML},
		{"yml", YML},
		{"toml", TOML},
		{"hcl", HCL},
		{"tf", TF},
		{"csv", CSV},
		{"xml", XML},
		{"ini", INI},
		{"gron", GRON},
		{"html", HTML},
	}

	for _, tt := range tests {
		result, err := GetEncodingType(tt.input)
		if err != nil {
			t.Errorf("unexpected error for type %s: %v", tt.input, err)
		} else if result != tt.expected {
			t.Errorf("expected %v, got %v", tt.expected, result)
		}
	}

	unsupportedResult, err := GetEncodingType("unsupported")
	if err == nil {
		t.Errorf("expected error for unsupported type, got result: %v", unsupportedResult)
	}
}

func TestMarshal(t *testing.T) {
	data := map[string]interface{}{"key": "value"}
	tests := []struct {
		encodingType EncodingType
	}{
		{JSON}, {YAML}, {YML}, {TOML}, {HCL}, {TF}, {CSV}, {XML}, {INI}, {GRON}, {HTML},
	}

	for _, tt := range tests {
		_, err := Marshal(data, tt.encodingType)
		if err != nil {
			t.Errorf("marshal failed for %v: %v", tt.encodingType, err)
		}
	}
}

func TestUnmarshal(t *testing.T) {
	jsonData := `{"key": "value"}`
	xmlData := `<root><key>value</key></root>`
	yamlData := "key: value"
	tomlData := "key = \"value\""

	tests := []struct {
		input        []byte
		encodingType EncodingType
	}{
		{[]byte(jsonData), JSON},
		{[]byte(xmlData), XML},
		{[]byte(yamlData), YAML},
		{[]byte(tomlData), TOML},
	}

	for _, tt := range tests {
		var data interface{} 
		err := Unmarshal(tt.input, tt.encodingType, &data)
		if err != nil {
			t.Errorf("unmarshal failed for %v: %v", tt.encodingType, err)
		}
		
		m, ok := data.(map[string]interface{})
		if !ok {
			t.Errorf("expected map[string]interface{}, got %T", data)
		}
		if value, ok := m["key"]; !ok || value != "value" {
			t.Errorf("expected key 'key' with value 'value', got %v", data)
		}
	}
}

func TestUnsupportedTypes(t *testing.T) {
	data := map[string]interface{}{"key": "value"}

	
	unsupportedType := EncodingType(len(SupportedFileTypes) + 1)
	_, err := Marshal(data, unsupportedType)
	if err == nil {
		t.Error("expected error for unsupported marshal type, got nil")
	}

	
	unsupportedData := []byte(`{"key": "value"}`)
	err = Unmarshal(unsupportedData, unsupportedType, &data)
	if err == nil {
		t.Error("expected error for unsupported unmarshal type, got nil")
	}
}

