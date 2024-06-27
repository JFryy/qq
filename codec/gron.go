package codec

import (
	"bytes"
	"fmt"
	"github.com/goccy/go-json"
	"reflect"
	"strconv"
	"strings"
)

func gronUnmarshal(data []byte, v interface{}) error {
	d := make(map[string]interface{})

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		parts := strings.SplitN(line, " = ", 2)
		if len(parts) != 2 {
			return nil
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(parts[1], `";`)

		setValueJSON(d, key, value)
	}

	*v.(*interface{}) = d
	return nil

}

func gronMarshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	traverseJSON("", v, &buf)
	return buf.Bytes(), nil
}

func traverseJSON(prefix string, v interface{}, buf *bytes.Buffer) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Map:
		for _, key := range rv.MapKeys() {
			strKey := fmt.Sprintf("%v", key)
			traverseJSON(addPrefix(prefix, strKey), rv.MapIndex(key).Interface(), buf)
		}
	case reflect.Slice:
		for i := 0; i < rv.Len(); i++ {
			traverseJSON(fmt.Sprintf("%s[%d]", prefix, i), rv.Index(i).Interface(), buf)
		}
	default:
		buf.WriteString(fmt.Sprintf("%s = %s;\n", prefix, formatJSONValue(v)))
	}
}

func addPrefix(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

func formatJSONValue(v interface{}) string {
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

func setValueJSON(data map[string]interface{}, key, value string) {
	parts := strings.Split(key, ".")
	var m = data
	for i, part := range parts {
		if i == len(parts)-1 {
			if strings.Contains(part, "[") && strings.Contains(part, "]") {
				k := strings.Split(part, "[")[0]
				if _, ok := m[k]; !ok {
					m[k] = make([]interface{}, 0)
				}

				m[k] = append(m[k].([]interface{}), value)
			} else {
				m[part] = value
			}
		} else {
			if _, ok := m[part]; !ok {
				m[part] = make(map[string]interface{})
			}
			m = m[part].(map[string]interface{})
		}
	}
}
