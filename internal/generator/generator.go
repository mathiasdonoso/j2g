package generator

import (
	"fmt"
	"strings"

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

	var nestedStructs []string

	for _, v := range input.Pairs {
		var vType string
		if isOrdererMap(v.V) {
			nestedValue, _ := v.V.(parser.OrdererMap)
			vType = title.String(v.Key)

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
			vType = title.String(v.Key)
		}

		s.WriteString("\t")
		s.WriteString(title.String(v.Key) + " " + vType + " ")
		s.WriteString("`json:\"" + v.Key + "\"`")
		s.WriteString("\n")
	}

	s.WriteString("}")

	result := strings.Builder{}
	for _, ns := range nestedStructs {
		result.WriteString(ns)
	}

	result.WriteString(s.String())
	result.WriteString("\n")

	return result.String(), nil
}
