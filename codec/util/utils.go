package util

import (
	"strconv"
	"strings"
	"time"
)

// DisableAutoConvert determines if string values in parser codecs are converted automatically.
var DisableAutoConvert bool

func ParseValue(value string) any {
	if DisableAutoConvert {
		return value
	}
	value = strings.TrimSpace(value)

	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}
	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		return floatValue
	}
	if boolValue, err := strconv.ParseBool(value); err == nil {
		return boolValue
	}
	if dateValue, err := time.Parse(time.RFC3339, value); err == nil {
		return dateValue
	}
	if dateValue, err := time.Parse("2006-01-02", value); err == nil {
		return dateValue
	}
	return value
}
