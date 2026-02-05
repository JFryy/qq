package jsonc

import (
	"testing"
)

func TestBasicJSONCUnmarshal(t *testing.T) {
	jsoncData := `{
		// This is a single-line comment
		"name": "Alice",
		"age": 30, // inline comment
		/* Multi-line
		   comment here */
		"city": "NYC"
	}`

	codec := &Codec{}
	var result map[string]any
	err := codec.Unmarshal([]byte(jsoncData), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSONC: %v", err)
	}

	if result["name"] != "Alice" {
		t.Errorf("Expected name 'Alice', got %v", result["name"])
	}
	if result["age"].(float64) != 30 {
		t.Errorf("Expected age 30, got %v", result["age"])
	}
	if result["city"] != "NYC" {
		t.Errorf("Expected city 'NYC', got %v", result["city"])
	}
}

func TestJSONCWithCommentsOnly(t *testing.T) {
	jsoncData := `{
		"name": "Bob",
		// Comment at the end
		"age": 25
	}`

	codec := &Codec{}
	var result map[string]any
	err := codec.Unmarshal([]byte(jsoncData), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSONC: %v", err)
	}

	if result["name"] != "Bob" {
		t.Errorf("Expected name 'Bob', got %v", result["name"])
	}
	if result["age"].(float64) != 25 {
		t.Errorf("Expected age 25, got %v", result["age"])
	}
}

func TestJSONCWithCommentsInStrings(t *testing.T) {
	jsoncData := `{
		// Real comment
		"url": "https://example.com/path?query=1", // Comment after URL
		"code": "if (x > 0) { /* not a comment */ }",
		"comment": "This has // slashes in it"
	}`

	codec := &Codec{}
	var result map[string]any
	err := codec.Unmarshal([]byte(jsoncData), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSONC: %v", err)
	}

	// Verify that "comment-like" sequences in strings are preserved
	if result["url"] != "https://example.com/path?query=1" {
		t.Errorf("URL mismatch: %v", result["url"])
	}
	if result["code"] != "if (x > 0) { /* not a comment */ }" {
		t.Errorf("Code mismatch: %v", result["code"])
	}
	if result["comment"] != "This has // slashes in it" {
		t.Errorf("Comment mismatch: %v", result["comment"])
	}
}

func TestJSONCMultiLineComment(t *testing.T) {
	jsoncData := `{
		/* This is a
		   multi-line comment
		   spanning several lines */
		"key1": "value1",
		/* Another /* nested comment attempt */
		"key2": "value2"
	}`

	codec := &Codec{}
	var result map[string]any
	err := codec.Unmarshal([]byte(jsoncData), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSONC with multi-line comments: %v", err)
	}

	if result["key1"] != "value1" {
		t.Errorf("Expected key1='value1', got %v", result["key1"])
	}
	if result["key2"] != "value2" {
		t.Errorf("Expected key2='value2', got %v", result["key2"])
	}
}

func TestJSONCMarshal(t *testing.T) {
	testData := map[string]any{
		"name": "Alice",
		"age":  30,
		"tags": []string{"developer", "golang"},
	}

	codec := &Codec{}

	// Marshal to JSONC (which is just pretty JSON)
	data, err := codec.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal JSONC: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Marshaled data is empty")
	}

	// Should be valid JSON
	var result map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal marshaled data: %v", err)
	}

	if result["name"] != "Alice" {
		t.Errorf("Name mismatch: %v", result["name"])
	}
}

func TestJSONCRoundTrip(t *testing.T) {
	jsoncInput := `{
		// Configuration file
		"database": {
			/* Connection settings */
			"host": "localhost",
			"port": 5432
		},
		"debug": true // Enable debug mode
	}`

	codec := &Codec{}

	// Unmarshal JSONC
	var intermediate map[string]any
	err := codec.Unmarshal([]byte(jsoncInput), &intermediate)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSONC: %v", err)
	}

	// Marshal back (without comments)
	data, err := codec.Marshal(intermediate)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal again to verify
	var result map[string]any
	err = codec.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed final unmarshal: %v", err)
	}

	// Verify structure preserved
	db := result["database"].(map[string]any)
	if db["host"] != "localhost" {
		t.Errorf("Host mismatch: %v", db["host"])
	}
	if result["debug"] != true {
		t.Errorf("Debug mismatch: %v", result["debug"])
	}
}

func TestJSONCEmptyComments(t *testing.T) {
	jsoncData := `{
		//
		"key": "value"
		/**/
	}`

	codec := &Codec{}
	var result map[string]any
	err := codec.Unmarshal([]byte(jsoncData), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSONC with empty comments: %v", err)
	}

	if result["key"] != "value" {
		t.Errorf("Expected key='value', got %v", result["key"])
	}
}

func TestJSONCEscapedQuotes(t *testing.T) {
	jsoncData := `{
		// Test escaped quotes
		"message": "She said \"hello\" to me",
		"path": "C:\\Users\\test"
	}`

	codec := &Codec{}
	var result map[string]any
	err := codec.Unmarshal([]byte(jsoncData), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSONC with escaped quotes: %v", err)
	}

	if result["message"] != `She said "hello" to me` {
		t.Errorf("Message mismatch: %v", result["message"])
	}
	if result["path"] != `C:\Users\test` {
		t.Errorf("Path mismatch: %v", result["path"])
	}
}
