package cli

import "testing"

func TestResolveVersionPrefersInjectedVersion(t *testing.T) {
	original := Version
	t.Cleanup(func() { Version = original })

	Version = "v9.9.9"
	if got := resolveVersion(); got != "v9.9.9" {
		t.Fatalf("resolveVersion() = %q, want %q", got, "v9.9.9")
	}
}

func TestResolveVersionFallsBackWhenNotInjected(t *testing.T) {
	original := Version
	t.Cleanup(func() { Version = original })

	Version = ""
	if got := resolveVersion(); got == "" {
		t.Fatal("resolveVersion() returned an empty string")
	}
}
