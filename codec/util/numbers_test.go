package util

import (
	"math/big"
	"reflect"
	"testing"
)

func TestReplaceBigIntsCopiesChangedContainers(t *testing.T) {
	n, _ := new(big.Int).SetString("123456789012345678901234567890", 10)
	untouched := map[string]any{"k": 1}
	input := map[string]any{"list": []any{1, n}, "other": untouched}
	got, err := ReplaceBigInts(input, func(n *big.Int) (any, error) {
		return n.String(), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{"list": []any{1, "123456789012345678901234567890"}, "other": untouched}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
	if input["list"].([]any)[1] != n {
		t.Fatal("input was modified")
	}
	if reflect.ValueOf(got.(map[string]any)["other"]).Pointer() != reflect.ValueOf(untouched).Pointer() {
		t.Fatal("unchanged container was copied")
	}
}

func TestNormalizeNumbersIntegersBeyondInt(t *testing.T) {
	got := NormalizeNumbers([]any{uint64(1 << 63), *big.NewInt(5), int64(-1)})
	items := got.([]any)
	if n, ok := items[0].(*big.Int); !ok || n.String() != "9223372036854775808" {
		t.Errorf("got %#v, want *big.Int 9223372036854775808", items[0])
	}
	if items[1] != 5 || items[2] != -1 {
		t.Errorf("got %#v, want 5 and -1 as int", items[1:])
	}
}
