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
			{Key: "value", V: 1.2},
			{Key: "name", V: "Alice"},
			{Key: "active", V: true},
		},
	}

	simpleResult := "type " + DEFAULT_STRUCT_NAME + " struct {\n" +
		"\tId int `json:\"id\"`\n" +
		"\tValue float64 `json:\"value\"`\n" +
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

	nestedResult := "type User struct {\n" +
		"\tId int `json:\"id\"`\n" +
		"\tName string `json:\"name\"`\n" +
		"}\n\n" +
		"type " + DEFAULT_STRUCT_NAME + " struct {\n" +
		"\tUser User `json:\"user\"`\n" +
		"\tStatus string `json:\"status\"`\n" +
		"}"

	// {[{items [{[{id 1} {name Item One}]} {[{id 2} {name Item Two}]}]}]}
	arrayInput := parser.OrdererMap{
		Pairs: []parser.KV{
			{
				Key: "items",
				V: []interface{}{
					parser.OrdererMap{
						Pairs: []parser.KV{
							{Key: "id", V: 1},
							{Key: "name", V: "Item One"},
						},
					},
					parser.OrdererMap{
						Pairs: []parser.KV{
							{Key: "id", V: 2},
							{Key: "name", V: "Item Two"},
						},
					},
				},
			},
		},
	}

	arrayResult := "type Item struct {\n" +
		"\tId int `json:\"id\"`\n" +
		"\tName string `json:\"name\"`\n" +
		"}\n\n" +
		"type " + DEFAULT_STRUCT_NAME + " struct {\n" +
		"\tItems []Item `json:\"items\"`\n" +
		"}"

	// {[{id 12345} {name Joe} {last_name Doe} {pull_count 0} {creation_time 2025-08-05T14:02:08.152Z} {update_time 2025-08-05T14:02:08.152Z}]}
	camelcaseInput := parser.OrdererMap{
		Pairs: []parser.KV{
			{Key: "id", V: 12345},
			{Key: "name", V: "Joe"},
			{Key: "last_name", V: "Doe"},
			{Key: "pull-count", V: 0},
			{Key: "creation_time_complete", V: "2025-08-05T14:02:08.152Z"},
			{Key: "update_time-complete", V: "2025-08-05T14:02:08.152Z"},
		},
	}

	camelcaseResult := "type " + DEFAULT_STRUCT_NAME + " struct {\n" +
		"\tId int `json:\"id\"`\n" +
		"\tName string `json:\"name\"`\n" +
		"\tLastName string `json:\"last_name\"`\n" +
		"\tPullCount int `json:\"pull-count\"`\n" +
		"\tCreationTimeComplete string `json:\"creation_time_complete\"`\n" +
		"\tUpdateTimeComplete string `json:\"update_time-complete\"`\n" +
		"}"

	// {[{80/tcp {[]}} {8080/tcp {[]}} {8443/tcp {[]}} {9001/tcp {[]}}]}
	withNumbersInput := parser.OrdererMap{
		Pairs: []parser.KV{
			{Key: "80/tcp", V: map[string]any{}},
			{Key: "8080/tcp", V: map[string]any{}},
			{Key: "8443/tcp", V: map[string]any{}},
			{Key: "9001/tcp", V: map[string]any{}},
		},
	}

	withNumbersResult := "type " + DEFAULT_STRUCT_NAME + " struct {\n" +
		"\tN80Tcp any `json:\"80/tcp\"`\n" +
		"\tN8080Tcp any `json:\"8080/tcp\"`\n" +
		"\tN8443Tcp any `json:\"8443/tcp\"`\n" +
		"\tN9001Tcp any `json:\"9001/tcp\"`\n" +
		"}"

	// {[{id 123} {name <nil>} {lastname {[]}} {jobs []}]}
	nullInput := parser.OrdererMap{
		Pairs: []parser.KV{
			{Key: "id", V: 123},
			{Key: "name", V: nil},
			{Key: "lastname", V: map[string]any{}},
		},
	}

	nullResult := "type " + DEFAULT_STRUCT_NAME + " struct {\n" +
		"\tId int `json:\"id\"`\n" +
		"\tName any `json:\"name\"`\n" +
		"\tLastname any `json:\"lastname\"`\n" +
		"}"

	tests := []struct {
		name      string
		input     parser.OrdererMap
		result    string
		shoudlErr bool
	}{
		{
			"simple",
			simpleInput,
			simpleResult,
			false,
		},
		{
			"nested",
			nestedInput,
			nestedResult,
			false,
		},
		{
			"array",
			arrayInput,
			arrayResult,
			false,
		},
		{
			"camelcase",
			camelcaseInput,
			camelcaseResult,
			false,
		},
		{
			"withNumbers",
			withNumbersInput,
			withNumbersResult,
			false,
		},
		{
			"null",
			nullInput,
			nullResult,
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
