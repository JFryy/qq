package env

import (
	"encoding/json"
	"testing"
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
	var result map[string]interface{}

	err := codec.Unmarshal([]byte(envContent), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal env: %v", err)
	}

	// Test DATABASE_URL
	dbURL, exists := result["DATABASE_URL"]
	if !exists {
		t.Error("DATABASE_URL not found")
	}

	if envVar, ok := dbURL.(map[string]interface{}); ok {
		if envVar["value"] != "postgresql://user:pass@localhost/db" {
			t.Errorf("Expected DATABASE_URL value, got %v", envVar["value"])
		}
		if envVar["type"] != "string" {
			t.Errorf("Expected string type, got %v", envVar["type"])
		}
	}

	// Test DATABASE_PORT (integer)
	dbPort, exists := result["DATABASE_PORT"]
	if !exists {
		t.Error("DATABASE_PORT not found")
	}

	if envVar, ok := dbPort.(map[string]interface{}); ok {
		if envVar["value"] != float64(5432) { // JSON numbers are float64
			t.Errorf("Expected DATABASE_PORT value 5432, got %v", envVar["value"])
		}
		if envVar["type"] != "integer" {
			t.Errorf("Expected integer type, got %v", envVar["type"])
		}
	}

	// Test DEBUG (boolean)
	debug, exists := result["DEBUG"]
	if !exists {
		t.Error("DEBUG not found")
	}

	if envVar, ok := debug.(map[string]interface{}); ok {
		if envVar["value"] != true {
			t.Errorf("Expected DEBUG value true, got %v", envVar["value"])
		}
		if envVar["type"] != "boolean" {
			t.Errorf("Expected boolean type, got %v", envVar["type"])
		}
	}

	// Test API_KEY (quoted string)
	apiKey, exists := result["API_KEY"]
	if !exists {
		t.Error("API_KEY not found")
	}

	if envVar, ok := apiKey.(map[string]interface{}); ok {
		if envVar["value"] != "secret-key-with-spaces" {
			t.Errorf("Expected API_KEY value, got %v", envVar["value"])
		}
	}

	// Test API_TIMEOUT (float)
	apiTimeout, exists := result["API_TIMEOUT"]
	if !exists {
		t.Error("API_TIMEOUT not found")
	}

	if envVar, ok := apiTimeout.(map[string]interface{}); ok {
		if envVar["value"] != 30.5 {
			t.Errorf("Expected API_TIMEOUT value 30.5, got %v", envVar["value"])
		}
		if envVar["type"] != "number" {
			t.Errorf("Expected number type, got %v", envVar["type"])
		}
	}

	// Test export
	path, exists := result["PATH"]
	if !exists {
		t.Error("PATH not found")
	}

	if envVar, ok := path.(map[string]interface{}); ok {
		if exported, ok := envVar["exported"].(bool); !ok || !exported {
			t.Error("Expected PATH to be exported")
		}
	}
}

func TestCommentsAndInlineComments(t *testing.T) {
	envContent := `# This is a comment
API_URL=https://api.example.com # Production API
SECRET_KEY="my-secret" # Keep this safe
PORT=8080`

	codec := Codec{}
	var result map[string]interface{}

	err := codec.Unmarshal([]byte(envContent), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal env: %v", err)
	}

	// Test inline comment
	apiURL, exists := result["API_URL"]
	if !exists {
		t.Error("API_URL not found")
	}

	if envVar, ok := apiURL.(map[string]interface{}); ok {
		if comment, ok := envVar["comment"].(string); !ok || comment != "Production API" {
			t.Errorf("Expected comment 'Production API', got %v", comment)
		}
	}

	// Test quoted value with comment
	secretKey, exists := result["SECRET_KEY"]
	if !exists {
		t.Error("SECRET_KEY not found")
	}

	if envVar, ok := secretKey.(map[string]interface{}); ok {
		if envVar["value"] != "my-secret" {
			t.Errorf("Expected SECRET_KEY value 'my-secret', got %v", envVar["value"])
		}
		if comment, ok := envVar["comment"].(string); !ok || comment != "Keep this safe" {
			t.Errorf("Expected comment 'Keep this safe', got %v", comment)
		}
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
	var result map[string]interface{}

	err := codec.Unmarshal([]byte(envContent), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal env: %v", err)
	}

	// Test empty quoted string
	emptyValue, exists := result["EMPTY_VALUE"]
	if !exists {
		t.Error("EMPTY_VALUE not found")
	}

	if envVar, ok := emptyValue.(map[string]interface{}); ok {
		if envVar["value"] != "" {
			t.Errorf("Expected empty string, got %v", envVar["value"])
		}
		if envVar["type"] != "string" {
			t.Errorf("Expected string type, got %v", envVar["type"])
		}
	}

	// Test zero as integer
	zeroValue, exists := result["ZERO_VALUE"]
	if !exists {
		t.Error("ZERO_VALUE not found")
	}

	if envVar, ok := zeroValue.(map[string]interface{}); ok {
		if envVar["value"] != float64(0) {
			t.Errorf("Expected 0, got %v", envVar["value"])
		}
		if envVar["type"] != "integer" {
			t.Errorf("Expected integer type, got %v", envVar["type"])
		}
	}

	// Test boolean variations
	trueValue, exists := result["TRUE_VALUE"]
	if !exists {
		t.Error("TRUE_VALUE not found")
	}

	if envVar, ok := trueValue.(map[string]interface{}); ok {
		if envVar["value"] != true {
			t.Errorf("Expected true, got %v", envVar["value"])
		}
	}

	// Test string with spaces
	spacesValue, exists := result["SPACES_VALUE"]
	if !exists {
		t.Error("SPACES_VALUE not found")
	}

	if envVar, ok := spacesValue.(map[string]interface{}); ok {
		if envVar["value"] != "  spaces  " {
			t.Errorf("Expected '  spaces  ', got %v", envVar["value"])
		}
	}
}

func TestMarshaling(t *testing.T) {
	original := map[string]interface{}{
		"DATABASE_URL": map[string]interface{}{
			"value":    "postgresql://localhost/db",
			"type":     "string",
			"comment":  "Database connection",
			"exported": false,
		},
		"DEBUG": map[string]interface{}{
			"value":    true,
			"type":     "boolean",
			"exported": false,
		},
		"PORT": map[string]interface{}{
			"value":    8080,
			"type":     "integer",
			"exported": true,
		},
	}

	codec := Codec{}
	data, err := codec.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal env: %v", err)
	}

	// Should be able to parse it back
	var result map[string]interface{}
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal marshaled env: %v", err)
	}

	// Check DATABASE_URL
	if dbURL, exists := result["DATABASE_URL"]; exists {
		if envVar, ok := dbURL.(map[string]interface{}); ok {
			if envVar["value"] != "postgresql://localhost/db" {
				t.Errorf("Marshaling round-trip failed for DATABASE_URL")
			}
		}
	}
}

func TestEscapeSequences(t *testing.T) {
	envContent := `MULTILINE="Line 1\nLine 2\tTabbed"
ESCAPED_QUOTES="He said \"Hello\""
BACKSLASHES="Path\\to\\file"`

	codec := Codec{}
	var result map[string]interface{}

	err := codec.Unmarshal([]byte(envContent), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal env: %v", err)
	}

	// Test multiline with escape sequences
	multiline, exists := result["MULTILINE"]
	if !exists {
		t.Error("MULTILINE not found")
	}

	if envVar, ok := multiline.(map[string]interface{}); ok {
		expected := "Line 1\nLine 2\tTabbed"
		if envVar["value"] != expected {
			t.Errorf("Expected %q, got %q", expected, envVar["value"])
		}
	}

	// Test escaped quotes
	escapedQuotes, exists := result["ESCAPED_QUOTES"]
	if !exists {
		t.Error("ESCAPED_QUOTES not found")
	}

	if envVar, ok := escapedQuotes.(map[string]interface{}); ok {
		expected := `He said "Hello"`
		if envVar["value"] != expected {
			t.Errorf("Expected %q, got %q", expected, envVar["value"])
		}
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
	var jsonResult map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonResult)
	if err != nil {
		t.Fatalf("Failed to unmarshal from JSON: %v", err)
	}

	// Verify structure is preserved
	if apiKey, exists := jsonResult["API_KEY"]; exists {
		if envVar, ok := apiKey.(map[string]interface{}); ok {
			if envVar["value"] != "secret123" {
				t.Error("JSON round-trip failed")
			}
		}
	}
}
