package generator

import (
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

		if !unicode.IsLetter(rune(keyName[0])) {
			keyName = "N" + keyName
		}

		if vType == "map[string]interface {}" || vType == "<nil>" {
			vType = "any"
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
