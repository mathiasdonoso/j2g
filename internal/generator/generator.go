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

// normalizeFieldName converts a JSON key into a valid, title-cased Go identifier.
// Non-alphanumeric characters are used as word separators and then removed.
// For example: "last_name" → "LastName", "80/tcp" → "80Tcp".
func normalizeFieldName(key string) string {
	segments := nonAlphanumeric.Split(key, -1)
	var b strings.Builder
	for _, seg := range segments {
		if seg != "" {
			b.WriteString(titleCaser.String(seg))
		}
	}
	return b.String()
}

// inferType determines the Go type string for a JSON value v associated with
// the normalized field name keyName. It returns the Go type, any nested struct
// definitions that must be emitted before the containing struct, and any error.
func (b *Builder) inferType(keyName string, v any) (goType string, nestedDefs []string, err error) {
	// Nested object: recursively build a named struct.
	if nestedValue, ok := v.(parser.OrdererMap); ok {
		nestedBuilder := Builder{StructName: keyName}
		def, err := nestedBuilder.BuildStruct(nestedValue)
		if err != nil {
			return "", nil, err
		}
		return keyName, []string{def}, nil
	}

	// Array: determine element type and emit nested struct if needed.
	if _, ok := v.([]any); ok {
		arr := v.([]any)

		if len(arr) == 0 {
			return "[]any", nil, nil
		}

		arrType := fmt.Sprintf("%T", arr[0])

		// Basic scalar types (e.g. []string, []int, []json.Number).
		if !strings.Contains(arrType, "interface") &&
			!strings.Contains(arrType, "any") &&
			!strings.Contains(arrType, "parser") {
			return "[]" + arrType, nil, nil
		}

		// Complex element type: derive struct name from field name via singularize.
		var elemTypeName string
		if len(keyName) > 1 {
			elemTypeName = keyName[:len(keyName)-1]
		} else {
			elemTypeName = fmt.Sprintf("%sType", keyName)
		}

		// Collect OrdererMap elements and use the first to define the struct.
		var omaps []parser.OrdererMap
		for _, elem := range arr {
			if m, ok := elem.(parser.OrdererMap); ok {
				omaps = append(omaps, m)
			}
		}
		first := parser.OrdererMap{}
		if len(omaps) > 0 {
			first = omaps[0]
		}

		nestedBuilder := Builder{StructName: elemTypeName}
		def, err := nestedBuilder.BuildStruct(first)
		if err != nil {
			return "", nil, err
		}
		return "[]" + elemTypeName, []string{def}, nil
	}

	// json.Number: distinguish int vs float64 by presence of a decimal point.
	if jn, ok := v.(json.Number); ok {
		if strings.Contains(jn.String(), ".") {
			return "float64", nil, nil
		}
		return "int", nil, nil
	}

	// Basic Go types.
	switch v.(type) {
	case string:
		return "string", nil, nil
	case bool:
		return "bool", nil, nil
	case nil:
		return "any", nil, nil
	default:
		goType := fmt.Sprintf("%T", v)
		if goType == "map[string]interface {}" {
			return "any", nil, nil
		}
		return goType, nil, nil
	}
}

// BuildFromArray generates struct definitions for a root-level JSON array.
// It uses the first element to define the element type and emits a slice alias.
// If the array is empty, it emits `type Results []interface{}` (using StructName
// if set). If the first element is not a JSON object, it returns an error.
func (b *Builder) BuildFromArray(arr []any) (string, error) {
	elemName := b.StructName
	if elemName == "" {
		elemName = defaultStructName
	}
	sliceName := elemName + "s"

	if len(arr) == 0 {
		return fmt.Sprintf("type %s []interface{}", sliceName), nil
	}

	first, ok := arr[0].(parser.OrdererMap)
	if !ok {
		return "", fmt.Errorf("root array element is not a JSON object")
	}

	elemBuilder := Builder{StructName: elemName}
	def, err := elemBuilder.BuildStruct(first)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s\n\ntype %s []%s", def, sliceName, elemName), nil
}

// nestedStructName returns the name of the nested struct that would be generated
// for the given field value, or an empty string if the field does not produce a
// named struct.
func nestedStructName(keyName string, v any) string {
	if _, ok := v.(parser.OrdererMap); ok {
		return keyName
	}
	if arr, ok := v.([]any); ok && len(arr) > 0 {
		// Only arrays of objects produce a named nested struct.
		if _, ok := arr[0].(parser.OrdererMap); ok {
			if len(keyName) > 1 {
				return keyName[:len(keyName)-1]
			}
			return fmt.Sprintf("%sType", keyName)
		}
	}
	return ""
}

func (b *Builder) BuildStruct(input parser.OrdererMap) (string, error) {
	structName := b.StructName
	if structName == "" {
		structName = defaultStructName
	}

	// Collision detection: only when the caller explicitly set a struct name.
	if b.StructName != "" {
		for _, kv := range input.Pairs {
			keyName := normalizeFieldName(kv.Key)
			nested := nestedStructName(keyName, kv.V)
			if nested != "" && nested == structName {
				return "", fmt.Errorf("--name %q conflicts with a nested struct of the same name", structName)
			}
		}
	}

	var fields strings.Builder
	var allNestedDefs strings.Builder

	for _, v := range input.Pairs {
		keyName := normalizeFieldName(v.Key)

		goType, nestedDefs, err := b.inferType(keyName, v.V)
		if err != nil {
			return "", err
		}

		for _, def := range nestedDefs {
			allNestedDefs.WriteString(def)
			allNestedDefs.WriteString("\n\n")
		}

		// Guard against a numeric first character in the field name.
		if !unicode.IsLetter(rune(keyName[0])) {
			keyName = "N" + keyName
		}

		fields.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"`\n", keyName, goType, v.Key))
	}

	var out strings.Builder
	if allNestedDefs.Len() > 0 {
		out.WriteString(allNestedDefs.String())
	}
	out.WriteString(fmt.Sprintf("type %s struct {\n%s}", structName, fields.String()))

	return out.String(), nil
}
