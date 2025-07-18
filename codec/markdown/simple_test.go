package markdown

import (
	"testing"
)

func TestHierarchicalStructure(t *testing.T) {
	markdown := `# My Document

This is intro content.

## Features

- Feature 1
- Feature 2

### Subfeature

Details about subfeature.

## Installation

Installation instructions here.`

	codec := NewCodec()
	result, err := codec.Parse(markdown)
	if err != nil {
		t.Fatalf("Failed to parse markdown: %v", err)
	}

	// Check root structure - "my-document" should be the root key (ID-ified)
	if _, exists := result["my-document"]; !exists {
		t.Errorf("Expected 'my-document' section to exist as root key, got keys: %v", getKeys(result))
	}

	// Get the document section
	docSection, ok := result["my-document"].(*Section)
	if !ok {
		t.Fatalf("Expected 'my-document' to be a Section, got %T", result["my-document"])
	}

	// Check basic section properties
	if docSection.Title != "My Document" {
		t.Errorf("Expected title 'My Document', got '%s'", docSection.Title)
	}

	if docSection.ID != "my-document" {
		t.Errorf("Expected ID 'my-document', got '%s'", docSection.ID)
	}

	// Check Features section exists under the document
	if _, exists := docSection.Sections["features"]; !exists {
		t.Error("Expected 'features' section to exist under 'my-document'")
	}

	// Check Installation section exists under the document
	if _, exists := docSection.Sections["installation"]; !exists {
		t.Error("Expected 'installation' section to exist under 'my-document'")
	}

	// Check that Features has subsections
	featuresSection, exists := docSection.Sections["features"]
	if !exists {
		t.Fatal("Features section not found")
	}

	if _, exists := featuresSection.Sections["subfeature"]; !exists {
		t.Error("Expected 'subfeature' section to exist under 'features'")
	}
}

func TestBasicUnmarshal(t *testing.T) {
	markdown := `# Test

Content here.`

	codec := NewCodec()
	var result map[string]interface{}

	err := codec.Unmarshal([]byte(markdown), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Check that "test" exists as root key (ID-ified)
	if _, exists := result["test"]; !exists {
		t.Errorf("Expected 'test' section to exist as root key, got keys: %v", getKeys(result))
	}
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
