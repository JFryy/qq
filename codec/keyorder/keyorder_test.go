package keyorder

import (
	"slices"
	"testing"
)

func TestKeys(t *testing.T) {
	Enable()
	t.Cleanup(Disable)

	m := map[string]any{"z": 1, "a": 2, "m": 3, "b": 4}
	Record(m, []string{"z", "missing", "a", "z"})

	// Recorded keys first, skipping duplicates and missing keys, then the
	// rest sorted.
	want := []string{"z", "a", "b", "m"}
	if got := Keys(m); !slices.Equal(got, want) {
		t.Errorf("Keys() = %v, want %v", got, want)
	}

	unrecorded := map[string]any{"z": 1, "a": 2}
	if got := Keys(unrecorded); !slices.Equal(got, []string{"a", "z"}) {
		t.Errorf("Keys() of unrecorded map = %v, want sorted", got)
	}
}

func TestRecordIgnoredUnlessRecording(t *testing.T) {
	m := map[string]any{"z": 1, "a": 2}

	Disable()
	Record(m, []string{"z", "a"})
	if Enabled() || Recording() {
		t.Fatal("expected registry to be disabled")
	}

	Enable()
	t.Cleanup(Disable)
	StopRecording()
	Record(m, []string{"z", "a"})
	if !Enabled() || Recording() {
		t.Fatal("expected registry enabled and not recording")
	}
	if got := Keys(m); got[0] != "a" {
		t.Errorf("Record after StopRecording took effect: %v", got)
	}
}

func TestEnableResetsRegistry(t *testing.T) {
	m := map[string]any{"z": 1, "a": 2}
	Enable()
	Record(m, []string{"z", "a"})
	Enable()
	t.Cleanup(Disable)
	if got := Keys(m); got[0] != "a" {
		t.Errorf("order survived Enable: %v", got)
	}
}
