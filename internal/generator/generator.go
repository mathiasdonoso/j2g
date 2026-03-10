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

const defaultStructName = "Result"

var nonAlphanumeric = regexp.MustCompile(`[^a-zA-Z0-9]+`)
var titleCaser = cases.Title(language.English)

type Builder struct {
	StructName string
}

func (b *Builder) BuildStruct(input parser.OrdererMap) (string, error) {
	structName := b.StructName

	if structName == "" {
		structName = defaultStructName
	}

	s := strings.Builder{}
	s.WriteString("type " + structName + " struct {")
	s.WriteString("\n")

	const SEPARATOR = "_"

	var out strings.Builder
	for _, v := range input.Pairs {
		formattedKey := nonAlphanumeric.ReplaceAllString(v.Key, SEPARATOR)
		sections := strings.Split(formattedKey, SEPARATOR)
		formatedSections := sections[:0]

		for _, section := range sections {
			formatedSections = append(formatedSections, titleCaser.String(section))
		}

		keyName := strings.Join(formatedSections, SEPARATOR)
		keyName = strings.ReplaceAll(keyName, SEPARATOR, "")
		var vType string
		if nestedValue, ok := v.V.(parser.OrdererMap); ok {
			vType = keyName

			nestedBuilder := Builder{
				StructName: vType,
			}

			n, err := nestedBuilder.BuildStruct(nestedValue)
			if err != nil {
				return "", err
			}

			out.WriteString(n)
			out.WriteString("\n\n")
		} else {
			vType = fmt.Sprintf("%T", v.V)

			if vType == "map[string]interface {}" || vType == "<nil>" {
				vType = "any"
			}

			if vType == "[]interface {}" {
				vType = "any"
				arr, ok := v.V.([]any)
				if !ok {
					return "", fmt.Errorf("fields is not a slice")
				}

				var arrType string
				if len(arr) > 0 {
					arrType = fmt.Sprintf("%T", arr[0])
					vType = arrType
				}

				basicType := true
				if strings.Contains(arrType, "interface") || strings.Contains(arrType, "any") || strings.Contains(arrType, "parser") {
					basicType = false
					if len(keyName) > 1 {
						vType = keyName[:len(keyName)-1]
					} else {
						vType = fmt.Sprintf("%sType", keyName)
					}
				}

				nestedBuilder := Builder{
					StructName: vType,
				}
				vType = fmt.Sprintf("[]%s", vType)

				if !basicType {
					var omaps []parser.OrdererMap
					for _, v := range arr {
						if m, ok := v.(parser.OrdererMap); ok {
							omaps = append(omaps, m)
						}
					}
					first := parser.OrdererMap{}
					if len(omaps) > 0 {
						first = omaps[0]
					}

					a, err := nestedBuilder.BuildStruct(first)
					if err != nil {
						return "", err
					}

					out.WriteString(a)
					out.WriteString("\n\n")
				}
			}
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

	out.WriteString(s.String())

	return out.String(), nil
}
