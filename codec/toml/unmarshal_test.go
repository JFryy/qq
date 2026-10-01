package toml

import (
	"reflect"
	"testing"

	"github.com/itchyny/gojq"
)

func TestUnmarshalArraysOfTablesQueries(t *testing.T) {
	input := []byte(`
count = 9007199254740993
empty = []
values = [1, 2]
[[plugins]]
id = "tab-was-taken"
priority = 7
[[plugins.dependencies]]
id = "nested"
[[plugins]]
id = "freedomchat"
inline = [{ id = "inline" }]
`)
	for _, target := range []string{"interface", "map"} {
		t.Run(target, func(t *testing.T) {
			var data any
			if target == "map" {
				var decoded map[string]any
				if err := (Codec{}).Unmarshal(input, &decoded); err != nil {
					t.Fatal(err)
				}
				data = decoded
			} else if err := (Codec{}).Unmarshal(input, &data); err != nil {
				t.Fatal(err)
			}

			for _, test := range []struct {
				query string
				want  any
			}{
				{".plugins[0].id", "tab-was-taken"},
				{"[.plugins[].id]", []any{"tab-was-taken", "freedomchat"}},
				{".plugins[0].dependencies[0].id", "nested"},
				{".plugins[1].inline[0].id", "inline"},
				{".plugins | length", 2},
				{".plugins[0].priority + 1", 8},
				{".empty | length", 0},
				{".values | add", 3},
				{".count | tostring", "9007199254740993"},
			} {
				t.Run(test.query, func(t *testing.T) {
					query, err := gojq.Parse(test.query)
					if err != nil {
						t.Fatal(err)
					}
					iter := query.Run(data)
					got, ok := iter.Next()
					if err, isError := got.(error); isError {
						t.Fatalf("query returned error of type %T", err)
					}
					if !ok || !reflect.DeepEqual(got, test.want) {
						t.Fatalf("got %#v, want %#v", got, test.want)
					}
					if _, ok := iter.Next(); ok {
						t.Fatal("unexpected extra query result")
					}
				})
			}
		})
	}
}

func TestUnmarshalTypedTables(t *testing.T) {
	var decoded struct {
		Plugins []struct{ ID string }
	}
	if err := (Codec{}).Unmarshal([]byte("[[plugins]]\nid = 'example'\n"), &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Plugins) != 1 || decoded.Plugins[0].ID != "example" {
		t.Fatalf("unexpected typed result: %#v", decoded)
	}
}

func TestUnmarshalInvalidInput(t *testing.T) {
	var decoded any
	if err := (Codec{}).Unmarshal([]byte("[[plugins]\n"), &decoded); err == nil {
		t.Fatal("expected invalid TOML to return an error")
	}
}
