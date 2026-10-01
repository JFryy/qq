package hcl

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/zclconf/go-cty/cty"
)

func TestConvertDecoderNumberTypes(t *testing.T) {
	for _, test := range []struct {
		input any
		want  cty.Value
	}{
		{int(-1), cty.NumberIntVal(-1)},
		{int8(-128), cty.NumberIntVal(-128)},
		{int16(-32768), cty.NumberIntVal(-32768)},
		{int32(math.MinInt32), cty.NumberIntVal(math.MinInt32)},
		{int64(math.MinInt64), cty.NumberIntVal(math.MinInt64)},
		{uint(1), cty.NumberUIntVal(1)},
		{uint8(255), cty.NumberUIntVal(255)},
		{uint16(65535), cty.NumberUIntVal(65535)},
		{uint32(math.MaxUint32), cty.NumberUIntVal(math.MaxUint32)},
		{uint64(math.MaxUint64), cty.NumberUIntVal(math.MaxUint64)},
		{float32(9.5), cty.NumberFloatVal(9.5)},
		{float64(9.5), cty.NumberFloatVal(9.5)},
	} {
		t.Run(fmt.Sprintf("%T", test.input), func(t *testing.T) {
			actual, err := (&Codec{}).convertToCtyValue(test.input)
			if err != nil {
				t.Fatal(err)
			}
			if !actual.RawEquals(test.want) {
				t.Fatalf("got %s, want %s", actual.GoString(), test.want.GoString())
			}
		})
	}
}

func TestConvertTimestamp(t *testing.T) {
	text := "1979-05-27T07:32:00.123456789+02:00"
	value, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := (&Codec{}).convertToCtyValue(value)
	if err != nil {
		t.Fatal(err)
	}
	if !actual.RawEquals(cty.StringVal(text)) {
		t.Fatalf("timestamp lost precision or timezone: %s", actual.GoString())
	}
	invalid := time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, err := (&Codec{}).convertToCtyValue(invalid); err == nil {
		t.Fatal("expected error for timestamp outside RFC3339 range")
	}
}
