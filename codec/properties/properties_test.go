package properties

import (
	"strings"
	"testing"
)

func TestBasicPropertiesMarshalUnmarshal(t *testing.T) {
	testData := map[string]string{
		"app.name":    "MyApp",
		"app.version": "1.0.0",
		"server.port": "8080",
		"debug.mode":  "true",
	}

	codec := &Codec{}

	// Test marshaling
	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal properties data: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Marshaled data is empty")
	}

	// Check that it contains expected keys
	propsStr := string(data)
	if !strings.Contains(propsStr, "app.name") || !strings.Contains(propsStr, "MyApp") {
		t.Error("Properties output missing expected content")
	}

	// Test unmarshaling
	var result map[string]string
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal properties data: %v", err)
	}

	// Verify length
	if len(result) != 4 {
		t.Fatalf("Expected 4 properties, got %d", len(result))
	}

	// Verify values
	if result["app.name"] != "MyApp" {
		t.Errorf("Expected app.name 'MyApp', got %v", result["app.name"])
	}
	if result["server.port"] != "8080" {
		t.Errorf("Expected server.port '8080', got %v", result["server.port"])
	}
}

func TestPropertiesWithComments(t *testing.T) {
	propsData := `# This is a comment
app.name=MyApp
! Another comment style
server.port=8080
# More comments
debug.mode=true`

	codec := &Codec{}
	var result map[string]string
	err := codec.Unmarshal([]byte(propsData), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal properties with comments: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("Expected 3 properties, got %d", len(result))
	}

	if result["app.name"] != "MyApp" {
		t.Errorf("Expected app.name 'MyApp', got %v", result["app.name"])
	}
}

func TestPropertiesWithEscapeSequences(t *testing.T) {
	propsData := `path=C\:\\Users\\Test
message=Hello\nWorld\tTab
special=equals\=colon\:space\ here`

	codec := &Codec{}
	var result map[string]string
	err := codec.Unmarshal([]byte(propsData), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal properties with escapes: %v", err)
	}

	if result["path"] != "C:\\Users\\Test" {
		t.Errorf("Expected escaped path, got %v", result["path"])
	}
	if result["message"] != "Hello\nWorld\tTab" {
		t.Errorf("Expected escaped newline and tab, got %v", result["message"])
	}
	if result["special"] != "equals=colon:space here" {
		t.Errorf("Expected unescaped special chars, got %v", result["special"])
	}
}

func TestPropertiesWithColonSeparator(t *testing.T) {
	propsData := `key1:value1
key2: value2
key3 : value3`

	codec := &Codec{}
	var result map[string]string
	err := codec.Unmarshal([]byte(propsData), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal properties with colon: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("Expected 3 properties, got %d", len(result))
	}

	if result["key1"] != "value1" {
		t.Errorf("Expected key1='value1', got %v", result["key1"])
	}
	if result["key2"] != "value2" {
		t.Errorf("Expected key2='value2', got %v", result["key2"])
	}
	if result["key3"] != "value3" {
		t.Errorf("Expected key3='value3', got %v", result["key3"])
	}
}

func TestPropertiesWithSpaces(t *testing.T) {
	propsData := `message=Hello World
path=/home/user/my folder/file.txt
empty=`

	codec := &Codec{}
	var result map[string]string
	err := codec.Unmarshal([]byte(propsData), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal properties with spaces: %v", err)
	}

	if result["message"] != "Hello World" {
		t.Errorf("Expected 'Hello World', got %v", result["message"])
	}
	if result["path"] != "/home/user/my folder/file.txt" {
		t.Errorf("Expected path with spaces, got %v", result["path"])
	}
	if result["empty"] != "" {
		t.Errorf("Expected empty value, got %v", result["empty"])
	}
}

func TestPropertiesMarshalEscaping(t *testing.T) {
	testData := map[string]string{
		"path":    "C:\\Users\\Test",
		"message": "Hello\nWorld\tTab",
		"key=2":   "value with = sign",
		"key:3":   "value with : sign",
		"key 4":   "value with space in key",
	}

	codec := &Codec{}

	// Test marshaling
	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal properties: %v", err)
	}

	propsStr := string(data)

	// Check that special characters are escaped in keys
	if !strings.Contains(propsStr, `key\=2`) {
		t.Error("Expected escaped equals in key")
	}
	if !strings.Contains(propsStr, `key\:3`) {
		t.Error("Expected escaped colon in key")
	}
	if !strings.Contains(propsStr, `key\ 4`) {
		t.Error("Expected escaped space in key")
	}

	// Unmarshal and verify
	var result map[string]string
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal marshaled properties: %v", err)
	}

	if result["path"] != "C:\\Users\\Test" {
		t.Errorf("Path mismatch: got %v", result["path"])
	}
	if result["message"] != "Hello\nWorld\tTab" {
		t.Errorf("Message mismatch: got %v", result["message"])
	}
}

func TestPropertiesRoundTrip(t *testing.T) {
	originalData := map[string]string{
		"database.url":      "jdbc:postgresql://localhost:5432/mydb",
		"database.username": "admin",
		"database.password": "secret123",
		"app.name":          "Test Application",
		"app.version":       "2.0.0",
	}

	codec := &Codec{}

	// Marshal original data
	data, err := codec.Marshal(originalData)
	if err != nil {
		t.Fatalf("Failed to marshal data: %v", err)
	}

	// Unmarshal to get result
	var result map[string]string
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal data: %v", err)
	}

	// Verify all keys and values are preserved
	if len(result) != len(originalData) {
		t.Fatalf("Expected %d properties, got %d", len(originalData), len(result))
	}

	for key, expectedValue := range originalData {
		if result[key] != expectedValue {
			t.Errorf("Mismatch for key %s: expected %s, got %s", key, expectedValue, result[key])
		}
	}
}

func TestPropertiesEmptyLines(t *testing.T) {
	propsData := `
# Comment

key1=value1


key2=value2

`

	codec := &Codec{}
	var result map[string]string
	err := codec.Unmarshal([]byte(propsData), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal properties with empty lines: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 properties, got %d", len(result))
	}
}
