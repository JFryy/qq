package gron

import "testing"

func TestUnmarshalRejectsConflictingPaths(t *testing.T) {
	for _, input := range []string{
		"a = 1\na.b = 2\n",
		"a = 1\na[0] = 2\n",
		"a[0] = 1\na[0].b = 2\n",
		"a[x] = 1\n",
		"a[-1] = 1\n",
		"a]0[ = 1\n",
	} {
		var got any
		if err := (&Codec{}).Unmarshal([]byte(input), &got); err == nil {
			t.Errorf("Unmarshal(%q) succeeded, want error", input)
		}
	}
}

func TestMarshalSortsKeys(t *testing.T) {
	got, err := (&Codec{}).Marshal(map[string]any{"z": 1, "a": map[string]any{"y": 2, "b": 3}})
	if err != nil {
		t.Fatal(err)
	}
	want := "a.b = 3;\na.y = 2;\nz = 1;\n"
	if string(got) != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
