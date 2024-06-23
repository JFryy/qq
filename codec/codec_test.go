package codec

import (
    "testing"
    "reflect"
    "github.com/goccy/go-yaml"
    "github.com/goccy/go-json"
    "github.com/stretchr/testify/assert"
    "github.com/JFryy/qq/codec"
)

func TestUnmarshalJSON(t *testing.T) {
    jsonData := []byte(`{"key": "value"}`)
    var data map[string]interface{}

    if err := json.Unmarshal(jsonData, &data); err != nil {
        t.Errorf("JSON unmarshal failed: %v", err)
    }

    // Assert that the unmarshaled data is as expected
    expectedData := map[string]interface{}{"key": "value"}
    if !reflect.DeepEqual(data, expectedData) {
        t.Errorf("JSON unmarshal result incorrect, got: %v, want: %v", data, expectedData)
    }
}

func TestMarshalJSON(t *testing.T) {
    data := map[string]interface{}{
        "key": "value",
    }

    jsonData, err := json.Marshal(data)
    if err != nil {
        t.Errorf("JSON marshal failed: %v", err)
    }

    // Assert that the marshaled JSON is as expected
    expectedJSON := `{"key":"value"}`
    if string(jsonData) != expectedJSON {
        t.Errorf("JSON marshal result incorrect, got: %s, want: %s", jsonData, expectedJSON)
    }
}

func TestUnmarshalYAML(t *testing.T) {
    yamlData := []byte("key: value\n")
    var data map[string]interface{}

    if err := yaml.Unmarshal(yamlData, &data); err != nil {
        t.Errorf("YAML unmarshal failed: %v", err)
    }

    // Assert that the unmarshaled data is as expected
    expectedData := map[string]interface{}{"key": "value"}
    if !reflect.DeepEqual(data, expectedData) {
        t.Errorf("YAML unmarshal result incorrect, got: %v, want: %v", data, expectedData)
    }
}

func TestMarshalYAML(t *testing.T) {
    data := map[string]interface{}{
        "key": "value",
    }

    yamlData, err := yaml.Marshal(data)
    if err != nil {
        t.Errorf("YAML marshal failed: %v", err)
    }

    // Assert that the marshaled YAML is as expected
    expectedYAML := "key: value\n"
    if string(yamlData) != expectedYAML {
        t.Errorf("YAML marshal result incorrect, got: %s, want: %s", yamlData, expectedYAML)
    }
}

func TestUnmarshalXML(t *testing.T) {
    xmlData := []byte("<root><key>value</key></root>")
    var data map[string]interface{}

    err := xmlUnmarshal(xmlData, &data)
    assert.NoError(t, err, "XML unmarshal should not return an error")

    expectedData := map[string]interface{}{"root": map[string]interface{}{"key": "value"}}
    assert.Equal(t, expectedData, data, "XML unmarshal result incorrect")
}

func TestMarshalXML(t *testing.T) {
    data := map[string]interface{}{
        "root": map[string]interface{}{
            "key": "value",
        },
    }

    xmlData, err := xmlMarshal(data)
    assert.NoError(t, err, "XML marshal should not return an error")

    expectedXML := "<map><entry key=\"root\"><map><entry key=\"key\">value</entry></map></entry></map>"
    assert.Equal(t, expectedXML, string(xmlData), "XML marshal result incorrect")
}

func TestUnmarshalHCL(t *testing.T) {
    hclData := []byte("key = \"value\"")
    var data map[string]interface{}

    err := hclUnmarshal(hclData, &data)
    assert.NoError(t, err, "HCL unmarshal should not return an error")

    expectedData := map[string]interface{}{"key": "value"}
    assert.Equal(t, expectedData, data, "HCL unmarshal result incorrect")
}

func TestMarshalHCL(t *testing.T) {
    data := map[string]interface{}{
        "key": "value",
    }

    hclData, err := hclMarshal(data)
    assert.NoError(t, err, "HCL marshal should not return an error")

    expectedHCL := "key = \"value\"\n"
    assert.Equal(t, expectedHCL, string(hclData), "HCL marshal result incorrect")
}
