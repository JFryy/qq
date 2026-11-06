package indent

import (
	"strings"
	"testing"
)

func TestSet(t *testing.T) {
	t.Cleanup(Reset)
	tests := []struct {
		n       int
		wantErr bool
	}{
		{-1, true},
		{0, false},
		{2, false},
		{7, false},
		{8, true},
	}
	for _, tt := range tests {
		Reset()
		err := Set(tt.n)
		if (err != nil) != tt.wantErr {
			t.Errorf("Set(%d) error = %v, wantErr %v", tt.n, err, tt.wantErr)
		}
		want := tt.n
		if tt.wantErr {
			want = Default
		}
		if got := Width(); got != want {
			t.Errorf("after Set(%d): Width() = %d, want %d", tt.n, got, want)
		}
		if got := String(); got != strings.Repeat(" ", want) {
			t.Errorf("after Set(%d): String() = %q, want %d spaces", tt.n, got, want)
		}
	}
}

func TestReset(t *testing.T) {
	if err := Set(5); err != nil {
		t.Fatal(err)
	}
	Reset()
	if got := Width(); got != Default {
		t.Errorf("Width() after Reset = %d, want %d", got, Default)
	}
}
