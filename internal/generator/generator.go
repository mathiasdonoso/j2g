package generator

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/mathiasdonoso/j2g/internal/parser"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const DEFAULT_STRUCT_NAME = "Result"

type Builder struct {
	StructName string
}

func isOrdererMap(t any) bool {
	switch t.(type) {
	case parser.OrdererMap:
		return true
	default:
		return false
	}
}

func (b *Builder) BuildStruct(input parser.OrdererMap) (string, error) {
	structName := b.StructName

	if structName == "" {
		structName = DEFAULT_STRUCT_NAME
	}

	s := strings.Builder{}
	s.WriteString("type " + structName + " struct {")
	s.WriteString("\n")
	title := cases.Title(language.English)

	const SEPARATOR = "_"

	var nestedStructs []string
	for _, v := range input.Pairs {
		re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
		formatedKey := re.ReplaceAllString(v.Key, SEPARATOR)
		sections := strings.Split(formatedKey, SEPARATOR)
		formatedSections := sections[:0]

		for _, section := range sections {
			formatedSections = append(formatedSections, title.String(section))
		}

		keyName := strings.Join(formatedSections, SEPARATOR)
		keyName = strings.ReplaceAll(keyName, SEPARATOR, "")
		var vType string
		if isOrdererMap(v.V) {
			nestedValue, _ := v.V.(parser.OrdererMap)
			vType = keyName

			nestedBuilder := Builder{
				StructName: vType,
			}

			n, err := nestedBuilder.BuildStruct(nestedValue)
			if err != nil {
				fmt.Printf("error: %v\n", err)
			}

			nestedStructs = append(nestedStructs, n)
			nestedStructs = append(nestedStructs, "\n\n")
		}

		vType = fmt.Sprintf("%T", v.V)
		if vType == "parser.OrdererMap" {
			vType = keyName
		}

		if vType == "map[string]interface {}" || vType == "<nil>" {
			vType = "any"
		}

		if vType == "[]interface {}" {
			if len(keyName) > 1 {
				vType = keyName[:len(keyName)-1]
			} else {
				vType = fmt.Sprintf("%sType", keyName)
			}
			nestedBuilder := Builder{
				StructName: vType,
			}
			vType = fmt.Sprintf("[]%s", vType)

			arr, ok := v.V.([]any)
			if !ok {
				fmt.Printf("error: fields is not a slice\n")
			}

			var omaps []parser.OrdererMap
			for _, v := range arr {
				if m, ok := v.(parser.OrdererMap); ok {
					omaps = append(omaps, m)
				}
			}

			first := omaps[0]

			a, err := nestedBuilder.BuildStruct(first)
			if err != nil {
				fmt.Printf("error: %v\n", err)
			}

			nestedStructs = append(nestedStructs, a)
			nestedStructs = append(nestedStructs, "\n\n")
		}

		if _, ok := v.V.(json.Number); ok {
			stringVal := fmt.Sprintf("%v", v.V)
			if strings.Contains(stringVal, ".") {
				vType = "float64"
			} else {
				vType = "int"
			}
		}

		if !unicode.IsLetter(rune(keyName[0])) {
			keyName = "N" + keyName
		}

		s.WriteString("\t")
		s.WriteString(keyName + " " + vType + " ")
		s.WriteString("`json:\"" + v.Key + "\"`")
		s.WriteString("\n")
	}

	s.WriteString("}")

	result := strings.Builder{}
	for _, ns := range nestedStructs {
		result.WriteString(ns)
	}

	result.WriteString(s.String())

	return result.String(), nil
}
