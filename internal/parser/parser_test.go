package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestParseJSON(t *testing.T) {
	simpleOrdererMap := OrdererMap{
		Pairs: []KV{
			{Key: "id", V: 1},
			{Key: "value", V: 1.2},
			{Key: "name", V: "Alice"},
			{Key: "active", V: true},
		},
	}

	arrayOrdererMap := OrdererMap{
		Pairs: []KV{
			{Key: "items", V: OrdererMap{
				Pairs: []KV{
					{Key: "name", V: "Item One"},
					{Key: "name", V: "Item Two"},
				},
			}},
		},
	}

	nestedOrdererMap := OrdererMap{
		Pairs: []KV{
			{
				Key: "user",
				V: OrdererMap{
					Pairs: []KV{
						{Key: "id", V: 42},
						{Key: "name", V: "Bob"},
					},
				},
			},
			{Key: "status", V: "ok"},
		},
	}

	camelcaseOrdererMap := OrdererMap{
		Pairs: []KV{
			{Key: "id", V: 12345},
			{Key: "name", V: "Joe"},
			{Key: "last_name", V: "Doe"},
			{Key: "pull_count", V: 0},
			{Key: "creation_time", V: "2025-08-05T14:02:08.152Z"},
			{Key: "update_time", V: "2025-08-05T14:02:08.152Z"},
		},
	}

	withNumbersOrdererMap := OrdererMap{
		Pairs: []KV{
			{Key: "80/tcp", V: "{}"},
			{Key: "8080/tcp", V: "{}"},
			{Key: "8443/tcp", V: "{}"},
			{Key: "9001/tcp", V: "{}"},
		},
	}

	tests := []struct {
		name               string
		inputFile          string
		shouldErr          bool
		expectedOrdererMap OrdererMap
	}{
		{"valid simple", "testdata/simple.json", false, simpleOrdererMap},
		{"valid array", "testdata/array.json", false, arrayOrdererMap},
		{"valid nested", "testdata/nested.json", false, nestedOrdererMap},
		{"valid camelcase", "testdata/camelcase.json", false, camelcaseOrdererMap},
		{"valid with numbers", "testdata/with_numbers.json", false, withNumbersOrdererMap},
		{"valid null values", "testdata/null.json", false, camelcaseOrdererMap},
		{"valid invalid", "testdata/invalid.json", true, OrdererMap{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := os.ReadFile(tt.inputFile)
			result, err := DecodeJSON(json.NewDecoder(bytes.NewReader(data)))
			fmt.Printf("structure: %v\n", result)

			if tt.shouldErr && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.shouldErr && reflect.DeepEqual(tt.expectedOrdererMap, result) {
				t.Errorf("expected result to be %+v but got %+v", tt.expectedOrdererMap, result)
			}
		})
	}
}
