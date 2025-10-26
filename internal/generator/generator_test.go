package generator

import (
	"testing"

	"github.com/mathiasdonoso/j2g/internal/parser"
)

func TestBuildStruct(t *testing.T) {
	// {[{id 1} {name Alice} {active true}]}
	simpleInput := parser.OrdererMap{
		Pairs: []parser.KV{
			{Key: "id", V: 1},
			{Key: "name", V: "Alice"},
			{Key: "active", V: true},
		},
	}

	simple := "type " + DEFAULT_STRUCT_NAME + " struct {\n" +
		"\tId int `json:\"id\"`\n" +
		"\tName string `json:\"name\"`\n" +
		"\tActive bool `json:\"active\"`\n" +
		"}"

	// {[{user {[{id 42} {name Bob}]}} {status ok}]}
	nestedInput := parser.OrdererMap{
		Pairs: []parser.KV{
			{
				Key: "user",
				V: parser.OrdererMap{
					Pairs: []parser.KV{
						{Key: "id", V: 42},
						{Key: "name", V: "Bob"},
					},
				},
			},
			{Key: "status", V: "ok"},
		},
	}

	nested := "type User struct {\n" +
		"\tId int `json:\"id\"`\n" +
		"\tName string `json:\"name\"`\n" +
		"}\n\n" +
		"type " + DEFAULT_STRUCT_NAME + " struct {\n" +
		"\tUser User `json:\"user\"`\n" +
		"\tStatus string `json:\"status\"`\n" +
		"}"

	// {[{items [{[{id 1} {name Item One}]} {[{id 2} {name Item Two}]}]}]}
	// arrayInput := []parser.OrdererMap{
	// 	Pairs: []parser.KV,
	// }

	tests := []struct {
		name      string
		input     parser.OrdererMap
		result    string
		shoudlErr bool
	}{
		{
			"simple",
			simpleInput,
			simple,
			false,
		},
		{
			"nested",
			nestedInput,
			nested,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := Builder{}
			result, err := builder.BuildStruct(tt.input)
			if tt.shoudlErr && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.shoudlErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.result != result {
				t.Errorf("malformed result, expected: %+v but got: %+v", tt.result, result)
			}
		})
	}
}
