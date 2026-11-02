package gron

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/JFryy/qq/codec/util"
	"github.com/goccy/go-json"
)

type Codec struct{}

func (c *Codec) Unmarshal(data []byte, v any) error {
	lines := strings.Split(string(data), "\n")
	var isArray bool
	dataMap := make(map[string]any)
	arrayData := make([]any, 0)

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		parts := strings.SplitN(line, " = ", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid line format: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(parts[1], `";`)
		parsedValue := util.ParseValue(value)

		if strings.HasPrefix(key, "[") && strings.Contains(key, "]") {
			isArray = true
		}

		if err := c.setValueJSON(dataMap, key, parsedValue); err != nil {
			return fmt.Errorf("invalid gron line %q: %w", line, err)
		}
	}

	if isArray && len(dataMap) == 1 {
		for _, val := range dataMap {
			if arrayVal, ok := val.([]any); ok {
				arrayData = arrayVal
			}
		}
	}

	vv := reflect.ValueOf(v)
	if vv.Kind() != reflect.Pointer || vv.IsNil() {
		return fmt.Errorf("provided value must be a non-nil pointer")
	}
	if isArray && len(arrayData) > 0 {
		vv.Elem().Set(reflect.ValueOf(arrayData))
	} else {
		vv.Elem().Set(reflect.ValueOf(dataMap))
	}

	return nil
}

func (c *Codec) Marshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	c.traverseJSON("", v, &buf)
	return buf.Bytes(), nil
}

func (c *Codec) traverseJSON(prefix string, v any, buf *bytes.Buffer) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Map:
		keys := rv.MapKeys()
		sort.Slice(keys, func(i, j int) bool {
			return fmt.Sprint(keys[i]) < fmt.Sprint(keys[j])
		})
		for _, key := range keys {
			strKey := fmt.Sprintf("%v", key)
			c.traverseJSON(addPrefix(prefix, strKey), rv.MapIndex(key).Interface(), buf)
		}
	case reflect.Slice:
		for i := 0; i < rv.Len(); i++ {
			c.traverseJSON(fmt.Sprintf("%s[%d]", prefix, i), rv.Index(i).Interface(), buf)
		}
	default:
		fmt.Fprintf(buf, "%s = %s;\n", prefix, formatJSONValue(v))
	}
}

func addPrefix(prefix, name string) string {
	if prefix == "" {
		return name
	}
	if strings.Contains(name, "[") && strings.Contains(name, "]") {
		return prefix + name
	}
	return prefix + "." + name
}

func formatJSONValue(v any) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case bool:
		return strconv.FormatBool(val)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	default:
		if v == nil {
			return "null"
		}
		data, _ := json.Marshal(v)
		return string(data)
	}
}

func (c *Codec) setValueJSON(data map[string]any, key string, value any) error {
	parts := strings.Split(key, ".")
	m := data
	for i, part := range parts {
		last := i == len(parts)-1
		if !strings.Contains(part, "[") || !strings.Contains(part, "]") {
			if last {
				m[part] = value
				return nil
			}
			if _, ok := m[part]; !ok {
				m[part] = make(map[string]any)
			}
			next, ok := m[part].(map[string]any)
			if !ok {
				return fmt.Errorf("%q is assigned both a value and nested fields", part)
			}
			m = next
			continue
		}

		k := strings.Split(part, "[")[0]
		index, err := parseArrayIndex(part)
		if err != nil {
			return err
		}
		if _, ok := m[k]; !ok {
			m[k] = make([]any, index+1)
		}
		arr, ok := m[k].([]any)
		if !ok {
			return fmt.Errorf("%q is assigned both a value and array elements", k)
		}
		for len(arr) <= index {
			arr = append(arr, nil)
		}
		m[k] = arr
		if last {
			arr[index] = value
			return nil
		}
		if arr[index] == nil {
			arr[index] = make(map[string]any)
		}
		next, ok := arr[index].(map[string]any)
		if !ok {
			return fmt.Errorf("%s is assigned both a value and nested fields", part)
		}
		m = next
	}
	return nil
}

func parseArrayIndex(part string) (int, error) {
	open, end := strings.Index(part, "["), strings.Index(part, "]")
	if end < open {
		return 0, fmt.Errorf("invalid array index in %q", part)
	}
	index, err := strconv.Atoi(strings.TrimSpace(part[open+1 : end]))
	if err != nil || index < 0 {
		return 0, fmt.Errorf("invalid array index in %q", part)
	}
	return index, nil
}
