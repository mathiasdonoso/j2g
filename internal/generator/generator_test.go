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
		{
			"camelcase",
			camelcaseInput,
			camelcaseResult,
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
