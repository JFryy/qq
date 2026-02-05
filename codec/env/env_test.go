package env

import (
	"testing"

	"github.com/goccy/go-json"
)

func TestBasicEnvParsing(t *testing.T) {
	envContent := `# Database configuration
DATABASE_URL=postgresql://user:pass@localhost/db
DATABASE_PORT=5432
DEBUG=true

# API Configuration  
API_KEY="secret-key-with-spaces"
API_TIMEOUT=30.5
FEATURE_ENABLED=false

# Export example
export PATH="/usr/local/bin:$PATH"`

	codec := Codec{}
	var result map[string]string

	err := codec.Unmarshal([]byte(envContent), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal env: %v", err)
	}

	// Test DATABASE_URL
	if result["DATABASE_URL"] != "postgresql://user:pass@localhost/db" {
		t.Errorf("Expected DATABASE_URL value, got %v", result["DATABASE_URL"])
	}

	// Test DATABASE_PORT (now returned as string)
	if result["DATABASE_PORT"] != "5432" {
		t.Errorf("Expected DATABASE_PORT value '5432', got %v", result["DATABASE_PORT"])
	}

	// Test DEBUG (now returned as string)
	if result["DEBUG"] != "true" {
		t.Errorf("Expected DEBUG value 'true', got %v", result["DEBUG"])
	}

	// Test API_KEY (quoted string)
	if result["API_KEY"] != "secret-key-with-spaces" {
		t.Errorf("Expected API_KEY value, got %v", result["API_KEY"])
	}

	// Test API_TIMEOUT (now returned as string)
	if result["API_TIMEOUT"] != "30.5" {
		t.Errorf("Expected API_TIMEOUT value '30.5', got %v", result["API_TIMEOUT"])
	}

	// Test export (simplified - we just get the value)
	if result["PATH"] != "/usr/local/bin:$PATH" {
		t.Errorf("Expected PATH value, got %v", result["PATH"])
	}
}

func TestCommentsAndInlineComments(t *testing.T) {
	envContent := `# This is a comment
API_URL=https://api.example.com # Production API
SECRET_KEY="my-secret" # Keep this safe
PORT=8080`

	codec := Codec{}
	var result map[string]string

	err := codec.Unmarshal([]byte(envContent), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal env: %v", err)
	}

	// Test that values are extracted correctly (comments are ignored in simplified version)
	if result["API_URL"] != "https://api.example.com" {
		t.Errorf("Expected API_URL value, got %v", result["API_URL"])
	}

	// Test quoted value (comment is ignored)
	if result["SECRET_KEY"] != "my-secret" {
		t.Errorf("Expected SECRET_KEY value 'my-secret', got %v", result["SECRET_KEY"])
	}

	if result["PORT"] != "8080" {
		t.Errorf("Expected PORT value '8080', got %v", result["PORT"])
	}
}

func TestSpecialValues(t *testing.T) {
	envContent := `EMPTY_VALUE=""
ZERO_VALUE=0
FALSE_VALUE=false
TRUE_VALUE=yes
NULL_VALUE=
SPACES_VALUE="  spaces  "`

	codec := Codec{}
	var result map[string]string

	err := codec.Unmarshal([]byte(envContent), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal env: %v", err)
	}

	// Test empty quoted string
	if result["EMPTY_VALUE"] != "" {
		t.Errorf("Expected empty string, got %v", result["EMPTY_VALUE"])
	}

	// Test zero (now returned as string)
	if result["ZERO_VALUE"] != "0" {
		t.Errorf("Expected '0', got %v", result["ZERO_VALUE"])
	}

	// Test boolean variations (now returned as strings)
	if result["TRUE_VALUE"] != "yes" {
		t.Errorf("Expected 'yes', got %v", result["TRUE_VALUE"])
	}

	if result["FALSE_VALUE"] != "false" {
		t.Errorf("Expected 'false', got %v", result["FALSE_VALUE"])
	}

	// Test string with spaces
	if result["SPACES_VALUE"] != "  spaces  " {
		t.Errorf("Expected '  spaces  ', got %v", result["SPACES_VALUE"])
	}

	// Test empty unquoted value
	if result["NULL_VALUE"] != "" {
		t.Errorf("Expected empty string for NULL_VALUE, got %v", result["NULL_VALUE"])
	}
}

func TestMarshaling(t *testing.T) {
	original := map[string]string{
		"DATABASE_URL": "postgresql://localhost/db",
		"DEBUG":        "true",
		"PORT":         "8080",
		"API_KEY":      "secret with spaces",
	}

	codec := Codec{}
	data, err := codec.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal env: %v", err)
	}

	// Should be able to parse it back
	var result map[string]string
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal marshaled env: %v", err)
	}

	// Check round-trip values
	if result["DATABASE_URL"] != "postgresql://localhost/db" {
		t.Errorf("Marshaling round-trip failed for DATABASE_URL")
	}
	if result["DEBUG"] != "true" {
		t.Errorf("Marshaling round-trip failed for DEBUG")
	}
	if result["PORT"] != "8080" {
		t.Errorf("Marshaling round-trip failed for PORT")
	}
	if result["API_KEY"] != "secret with spaces" {
		t.Errorf("Marshaling round-trip failed for API_KEY")
	}
}

func TestEscapeSequences(t *testing.T) {
	envContent := `MULTILINE="Line 1\nLine 2\tTabbed"
ESCAPED_QUOTES="He said \"Hello\""
BACKSLASHES="Path\\to\\file"`

	codec := Codec{}
	var result map[string]string

	err := codec.Unmarshal([]byte(envContent), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal env: %v", err)
	}

	// Test multiline with escape sequences
	expected := "Line 1\nLine 2\tTabbed"
	if result["MULTILINE"] != expected {
		t.Errorf("Expected %q, got %q", expected, result["MULTILINE"])
	}

	// Test escaped quotes
	expectedQuotes := `He said "Hello"`
	if result["ESCAPED_QUOTES"] != expectedQuotes {
		t.Errorf("Expected %q, got %q", expectedQuotes, result["ESCAPED_QUOTES"])
	}

	// Test backslashes
	expectedBackslashes := `Path\to\file`
	if result["BACKSLASHES"] != expectedBackslashes {
		t.Errorf("Expected %q, got %q", expectedBackslashes, result["BACKSLASHES"])
	}
}

func TestJSONCompatibility(t *testing.T) {
	envContent := `API_KEY=secret123
DEBUG=true
PORT=8080
TIMEOUT=30.5`

	codec := Codec{}
	result, err := codec.Parse(envContent)
	if err != nil {
		t.Fatalf("Failed to parse env: %v", err)
	}

	// Should be JSON-serializable
	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	// Should be JSON-deserializable
	var jsonResult map[string]string
	err = json.Unmarshal(jsonData, &jsonResult)
	if err != nil {
		t.Fatalf("Failed to unmarshal from JSON: %v", err)
	}

	// Verify structure is preserved (simplified format)
	if jsonResult["API_KEY"] != "secret123" {
		t.Error("JSON round-trip failed")
	}
	if jsonResult["DEBUG"] != "true" {
		t.Error("JSON round-trip failed for DEBUG")
	}
	if jsonResult["PORT"] != "8080" {
		t.Error("JSON round-trip failed for PORT")
	}
}
