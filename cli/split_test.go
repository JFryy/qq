package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JFryy/qq/codec"
	"github.com/itchyny/gojq"
)

const splitManifest = `apiVersion: v1
kind: Service
metadata:
  name: web
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
`

func runSplit(t *testing.T, dir, input, expr, nameExpr string, overwrite bool) error {
	t.Helper()
	var data any
	if err := codec.Unmarshal([]byte(input), codec.YAML, &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	query, err := gojq.Parse(expr)
	if err != nil {
		t.Fatalf("parse query: %v", err)
	}
	nameQuery, err := gojq.Parse(nameExpr)
	if err != nil {
		t.Fatalf("parse split expression: %v", err)
	}
	return writeSplitResults(query, nameQuery, data, codec.YAML, dir, overwrite)
}

func TestSplit_MultiDocYAMLByName(t *testing.T) {
	dir := t.TempDir()
	if err := runSplit(t, dir, splitManifest, ".[]", ".metadata.name", false); err != nil {
		t.Fatalf("split failed: %v", err)
	}

	for name, kind := range map[string]string{"web.yaml": "Service", "api.yaml": "Deployment"} {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}
		var doc map[string]any
		if err := codec.Unmarshal(b, codec.YAML, &doc); err != nil {
			t.Fatalf("parsing %s: %v", name, err)
		}
		if doc["kind"] != kind {
			t.Errorf("%s: expected kind %s, got %v", name, kind, doc["kind"])
		}
	}

	entries, _ := os.ReadDir(dir)
	if len(entries) != 2 {
		t.Errorf("expected 2 files, got %d", len(entries))
	}
}

func TestSplit_ExtensionFromOutputFormat(t *testing.T) {
	dir := t.TempDir()
	query, _ := gojq.Parse(".[]")
	nameQuery, _ := gojq.Parse(".id")
	data := []any{map[string]any{"id": 1}, map[string]any{"id": 2.5}}
	if err := writeSplitResults(query, nameQuery, data, codec.JSON, dir, false); err != nil {
		t.Fatalf("split failed: %v", err)
	}
	for _, name := range []string{"1.json", "2.5.json"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected %s: %v", name, err)
		}
	}
}

func TestSplit_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "web.yaml")
	if err := os.WriteFile(existing, []byte("keep\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := runSplit(t, dir, splitManifest, ".[]", ".metadata.name", false)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected already exists error, got %v", err)
	}
	b, _ := os.ReadFile(existing)
	if string(b) != "keep\n" {
		t.Errorf("existing file was modified: %q", b)
	}

	if err := runSplit(t, dir, splitManifest, ".[]", ".metadata.name", true); err != nil {
		t.Fatalf("split with overwrite failed: %v", err)
	}
	b, _ = os.ReadFile(existing)
	if !strings.Contains(string(b), "kind: Service") {
		t.Errorf("existing file was not overwritten: %q", b)
	}
}

func TestSplit_DuplicateNames(t *testing.T) {
	dir := t.TempDir()
	err := runSplit(t, dir, splitManifest, ".[]", ".apiVersion | split(\"/\") | last", true)
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate name error, got %v", err)
	}
}

func TestSplit_RejectsUnsafeNames(t *testing.T) {
	for _, name := range []string{"", ".", "..", "../x", "/etc/x", "a/b", `a\b`, "./x"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			input := "metadata:\n  name: '" + name + "'\n"
			err := runSplit(t, dir, input, ".", ".metadata.name", false)
			if err == nil {
				t.Fatalf("expected error for name %q", name)
			}
			entries, _ := os.ReadDir(dir)
			if len(entries) != 0 {
				t.Errorf("expected no files to be written, got %d", len(entries))
			}
		})
	}
}

func TestSplit_RejectsNonStringName(t *testing.T) {
	for _, expr := range []string{".metadata", ".missing", "true", "[1]", "empty", "1, 2"} {
		t.Run(expr, func(t *testing.T) {
			dir := t.TempDir()
			if err := runSplit(t, dir, splitManifest, ".[0]", expr, false); err == nil {
				t.Fatalf("expected error for split expression %q", expr)
			}
		})
	}
}
