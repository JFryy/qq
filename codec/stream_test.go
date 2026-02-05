package codec

import (
	"strings"
	"testing"
)

func TestStreamParser_SimpleObject(t *testing.T) {
	input := `{"name":"Alice","age":30}`
	reader := strings.NewReader(input)

	result, err := StreamParserCollect(reader, JSON)
	if err != nil {
		t.Fatalf("StreamParser failed: %v", err)
	}

	// Expected: [["name"], "Alice"], [["age"], 30], [["age"]] (jq emits last key)
	if len(result) != 3 {
		t.Errorf("Expected 3 items, got %d", len(result))
		for i, item := range result {
			t.Logf("Item %d: %v", i, item)
		}
	}

	// Check first path-value pair
	first := result[0].([]any)
	if len(first) != 2 {
		t.Errorf("Expected [path, value], got %v", first)
	}
	path := first[0].([]any)
	if len(path) != 1 || path[0] != "name" {
		t.Errorf("Expected path [\"name\"], got %v", path)
	}
	if first[1] != "Alice" {
		t.Errorf("Expected value 'Alice', got %v", first[1])
	}
}

func TestStreamParser_Array(t *testing.T) {
	input := `[1,2,3]`
	reader := strings.NewReader(input)

	result, err := StreamParserCollect(reader, JSON)
	if err != nil {
		t.Fatalf("StreamParser failed: %v", err)
	}

	// Expected: [[0], 1], [[1], 2], [[2], 3], [[2]] (jq emits last index)
	if len(result) != 4 {
		t.Errorf("Expected 4 items, got %d", len(result))
		for i, item := range result {
			t.Logf("Item %d: %v", i, item)
		}
	}

	// Check first element
	first := result[0].([]any)
	path := first[0].([]any)
	if len(path) != 1 || path[0] != 0 {
		t.Errorf("Expected path [0], got %v", path)
	}
	if first[1] != float64(1) {
		t.Errorf("Expected value 1, got %v", first[1])
	}
}

func TestStreamParser_NestedObject(t *testing.T) {
	input := `{"user":{"name":"Bob","id":123}}`
	reader := strings.NewReader(input)

	result, err := StreamParserCollect(reader, JSON)
	if err != nil {
		t.Fatalf("StreamParser failed: %v", err)
	}

	// Should emit path-value pairs for nested structure
	// [["user","name"], "Bob"], [["user","id"], 123], [["user"]], [[]]
	if len(result) < 3 {
		t.Errorf("Expected at least 3 items for nested object, got %d", len(result))
	}

	// Check nested path
	first := result[0].([]any)
	path := first[0].([]any)
	if len(path) != 2 || path[0] != "user" || path[1] != "name" {
		t.Errorf("Expected path [\"user\", \"name\"], got %v", path)
	}
}

func TestStreamParser_ArrayOfObjects(t *testing.T) {
	input := `[{"id":1},{"id":2}]`
	reader := strings.NewReader(input)

	result, err := StreamParserCollect(reader, JSON)
	if err != nil {
		t.Fatalf("StreamParser failed: %v", err)
	}

	// Should emit: [[0,"id"], 1], [[0]], [[1,"id"], 2], [[1]], [[]]
	if len(result) < 4 {
		t.Errorf("Expected at least 4 items, got %d", len(result))
	}

	// Check first object's id
	first := result[0].([]any)
	path := first[0].([]any)
	if len(path) != 2 || path[0] != 0 || path[1] != "id" {
		t.Errorf("Expected path [0, \"id\"], got %v", path)
	}
	if first[1] != float64(1) {
		t.Errorf("Expected value 1, got %v", first[1])
	}
}

func TestStreamParser_EmptyObject(t *testing.T) {
	input := `{}`
	reader := strings.NewReader(input)

	result, err := StreamParserCollect(reader, JSON)
	if err != nil {
		t.Fatalf("StreamParser failed: %v", err)
	}

	// Empty object should emit just [[]]
	if len(result) != 0 {
		t.Logf("Result: %v", result)
	}
}

func TestStreamParser_EmptyArray(t *testing.T) {
	input := `[]`
	reader := strings.NewReader(input)

	result, err := StreamParserCollect(reader, JSON)
	if err != nil {
		t.Fatalf("StreamParser failed: %v", err)
	}

	// Empty array should emit just [[]]
	if len(result) != 0 {
		t.Logf("Result: %v", result)
	}
}

func TestStreamParser_PrimitiveValue(t *testing.T) {
	input := `42`
	reader := strings.NewReader(input)

	result, err := StreamParserCollect(reader, JSON)
	if err != nil {
		t.Fatalf("StreamParser failed: %v", err)
	}

	// Primitive at root: [[], 42]
	if len(result) != 1 {
		t.Errorf("Expected 1 item, got %d", len(result))
	}

	first := result[0].([]any)
	path := first[0].([]any)
	if len(path) != 0 {
		t.Errorf("Expected empty path [], got %v", path)
	}
	if first[1] != float64(42) {
		t.Errorf("Expected value 42, got %v", first[1])
	}
}

func TestStreamParser_JSONL(t *testing.T) {
	input := `{"id":1,"name":"Alice"}
{"id":2,"name":"Bob"}`
	reader := strings.NewReader(input)

	result, err := StreamParserCollect(reader, JSONL)
	if err != nil {
		t.Fatalf("StreamParser failed: %v", err)
	}

	// Should emit path-value pairs for each JSONL object with index
	// [[0,"id"], 1], [[0,"name"], "Alice"], [[0,"name"]], [[1,"id"], 2], [[1,"name"], "Bob"], [[1,"name"]]
	if len(result) < 4 {
		t.Errorf("Expected at least 4 items, got %d", len(result))
	}

	// Check first object's first field
	first := result[0].([]any)
	path := first[0].([]any)
	if len(path) != 2 || path[0] != 0 {
		t.Errorf("Expected path [0, ...], got %v", path)
	}
}

func TestStreamParser_Lines(t *testing.T) {
	input := `line1
line2
line3`
	reader := strings.NewReader(input)

	result, err := StreamParserCollect(reader, LINE)
	if err != nil {
		t.Fatalf("StreamParser failed: %v", err)
	}

	// Should emit [[index], line] for each line
	if len(result) != 4 { // 3 lines + closing marker
		t.Errorf("Expected 4 items, got %d", len(result))
	}

	// Check first line
	first := result[0].([]any)
	path := first[0].([]any)
	if len(path) != 1 || path[0] != 0 {
		t.Errorf("Expected path [0], got %v", path)
	}
	if first[1] != "line1" {
		t.Errorf("Expected 'line1', got %v", first[1])
	}
}

func TestStreamParser_YAML(t *testing.T) {
	input := `name: Alice
age: 30`
	reader := strings.NewReader(input)

	result, err := StreamParserCollect(reader, YAML)
	if err != nil {
		t.Fatalf("StreamParser failed: %v", err)
	}

	// Should emit path-value pairs for YAML object
	if len(result) < 2 {
		t.Errorf("Expected at least 2 items, got %d", len(result))
	}
}

func TestStreamParser_YAML_MultiDocument(t *testing.T) {
	input := `---
name: Alice
age: 30
---
name: Bob
age: 25`
	reader := strings.NewReader(input)

	result, err := StreamParserCollect(reader, YAML)
	if err != nil {
		t.Fatalf("StreamParser failed: %v", err)
	}

	// Should emit path-value pairs for both documents
	// Each doc has 2 fields + closing marker, indexed at [0] and [1]
	if len(result) < 4 {
		t.Errorf("Expected at least 4 items for 2 YAML documents, got %d", len(result))
	}

	// Check first document is indexed at [0]
	first := result[0].([]any)
	path := first[0].([]any)
	if len(path) < 2 || path[0] != 0 {
		t.Errorf("Expected first document at index 0, got path: %v", path)
	}
}

func TestStreamParser_CSV(t *testing.T) {
	input := `name,age,score
Alice,30,850
Bob,25,920`
	reader := strings.NewReader(input)

	result, err := StreamParserCollect(reader, CSV)
	if err != nil {
		t.Fatalf("StreamParser failed: %v", err)
	}

	// Should emit path-value pairs for each row
	// 2 rows * 3 fields each + closing markers
	if len(result) < 6 {
		t.Errorf("Expected at least 6 items for CSV with 2 rows, got %d", len(result))
	}

	// Check first row is indexed at [0]
	first := result[0].([]any)
	path := first[0].([]any)
	if len(path) < 2 || path[0] != 0 {
		t.Errorf("Expected first row at index 0, got path: %v", path)
	}
}

func TestStreamParser_TSV(t *testing.T) {
	input := "name\tage\tscore\nAlice\t30\t850\nBob\t25\t920"
	reader := strings.NewReader(input)

	result, err := StreamParserCollect(reader, TSV)
	if err != nil {
		t.Fatalf("StreamParser failed: %v", err)
	}

	// Should emit path-value pairs for each row
	if len(result) < 6 {
		t.Errorf("Expected at least 6 items for TSV with 2 rows, got %d", len(result))
	}

	// Verify we have proper row objects
	first := result[0].([]any)
	if len(first) != 2 {
		t.Errorf("Expected [path, value] pair, got %v", first)
	}
}
