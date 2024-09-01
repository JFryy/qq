package util

import (
	"strconv"
	"strings"
	"time"
)

func ParseValue(value string) interface{} {
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
