package codec

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/goccy/go-json"
)

type ProtoFile struct {
	PackageName string
	Messages    map[string]Message
	Enums       map[string]Enum
}

type Message struct {
	Name   string
	Fields map[string]Field
}

type Field struct {
	Name   string
	Type   string
	Number int
}

type Enum struct {
	Name   string
	Values map[string]int
}

type Codec struct{}

func (c *Codec) Unmarshal(input []byte, v interface{}) error {
	protoContent := string(input)

	protoContent = removeComments(protoContent)

	protoFile := &ProtoFile{Messages: make(map[string]Message), Enums: make(map[string]Enum)}

	messagePattern := `message\s+([A-Za-z0-9_]+)\s*\{([^}]*)\}`
	fieldPattern := `([A-Za-z0-9_]+)\s+([A-Za-z0-9_]+)\s*=\s*(\d+);`
	enumPattern := `enum\s+([A-Za-z0-9_]+)\s*\{([^}]*)\}`
	enumValuePattern := `([A-Za-z0-9_]+)\s*=\s*(-?\d+);`

	re := regexp.MustCompile(messagePattern)
	fieldRe := regexp.MustCompile(fieldPattern)
	enumRe := regexp.MustCompile(enumPattern)
	enumValueRe := regexp.MustCompile(enumValuePattern)

	packagePattern := `package\s+([A-Za-z0-9_]+);`
	packageRe := regexp.MustCompile(packagePattern)
	packageMatch := packageRe.FindStringSubmatch(protoContent)
	if len(packageMatch) > 0 {
		protoFile.PackageName = packageMatch[1]
	}

	matches := re.FindAllStringSubmatch(protoContent, -1)
	for _, match := range matches {
		messageName := match[1]
		messageContent := match[2]

		fields := make(map[string]Field)
		fieldMatches := fieldRe.FindAllStringSubmatch(messageContent, -1)
		for _, fieldMatch := range fieldMatches {
			fieldType := fieldMatch[1]
			fieldName := fieldMatch[2]
			fieldNumber, err := strconv.Atoi(fieldMatch[3])
            if err != nil {
                return err
            }
			fields[fieldName] = Field{
				Name:   fieldName,
				Type:   fieldType,
				Number: fieldNumber,
			}
		}

		protoFile.Messages[messageName] = Message{
			Name:   messageName,
			Fields: fields,
		}
	}

	enumMatches := enumRe.FindAllStringSubmatch(protoContent, -1)
	for _, match := range enumMatches {
		enumName := match[1]
		enumContent := match[2]

		enumValues := make(map[string]int)
		enumValueMatches := enumValueRe.FindAllStringSubmatch(enumContent, -1)
		for _, enumValueMatch := range enumValueMatches {
			enumValueName := enumValueMatch[1]
			enumValueNumber := enumValueMatch[2]
			number, err := strconv.Atoi(enumValueNumber)
			if err != nil {
				return nil
			}
			enumValues[enumValueName] = number
		}

		protoFile.Enums[enumName] = Enum{
			Name:   enumName,
			Values: enumValues,
		}
	}
	jsonMap, err := ConvertProtoToJSON(protoFile)
	if err != nil {
		return fmt.Errorf("error converting to JSON: %v", err)
	}
	jsonData, err := json.Marshal(jsonMap)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}
	return json.Unmarshal(jsonData, v)
}

func removeComments(input string) string {
	reSingleLine := regexp.MustCompile(`//.*`)
	input = reSingleLine.ReplaceAllString(input, "")
	reMultiLine := regexp.MustCompile(`/\*.*?\*/`)
	input = reMultiLine.ReplaceAllString(input, "")
	return strings.TrimSpace(input)
}

func ConvertProtoToJSON(protoFile *ProtoFile) (map[string]interface{}, error) {
	jsonMap := make(map[string]interface{})
	packageMap := make(map[string]interface{})
	packageMap["message"] = make(map[string]interface{})
	packageMap["enums"] = make(map[string]interface{})

	for messageName, message := range protoFile.Messages {
		fieldsList := []interface{}{}
		for name, field := range message.Fields {
			values := make(map[string]interface{})
			values["name"] = name
			values["type"] = field.Type
			values["number"] = field.Number
			fieldsList = append(fieldsList, values)
		}
		packageMap["message"].(map[string]interface{})[messageName] = fieldsList
	}

	for enumName, enum := range protoFile.Enums {
		valuesMap := make(map[string]interface{})
		for enumValueName, enumValueNumber := range enum.Values {
			valuesMap[enumValueName] = enumValueNumber
		}
		packageMap["enums"].(map[string]interface{})[enumName] = valuesMap
	}

	jsonMap[protoFile.PackageName] = packageMap

	return jsonMap, nil
}
