package parser

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseJSON(t *testing.T) {
	simpleOrdererMap := OrdererMap{
		Pairs: []KV{
			{Key: "id", V: json.Number("1")},
			{Key: "value", V: json.Number("1.2")},
			{Key: "name", V: "Alice"},
			{Key: "active", V: true},
		},
	}

	arrayOrdererMap := OrdererMap{
		Pairs: []KV{
			{Key: "items", V: []interface{}{
				OrdererMap{
					Pairs: []KV{
						{Key: "id", V: json.Number("1")},
						{Key: "name", V: "Item One"},
					},
				},
				OrdererMap{
					Pairs: []KV{
						{Key: "id", V: json.Number("2")},
						{Key: "name", V: "Item Two"},
					},
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
						{Key: "id", V: json.Number("42")},
						{Key: "name", V: "Bob"},
					},
				},
			},
			{Key: "status", V: "ok"},
		},
	}

	camelcaseOrdererMap := OrdererMap{
		Pairs: []KV{
			{Key: "id", V: json.Number("12345")},
			{Key: "name", V: "Joe"},
			{Key: "last_name", V: "Doe"},
			{Key: "pull_count", V: json.Number("0")},
			{Key: "creation_time", V: "2025-08-05T14:02:08.152Z"},
			{Key: "update_time", V: "2025-08-05T14:02:08.152Z"},
		},
	}

	withNumbersOrdererMap := OrdererMap{
		Pairs: []KV{
			{Key: "80/tcp", V: OrdererMap{}},
			{Key: "8080/tcp", V: OrdererMap{}},
			{Key: "8443/tcp", V: OrdererMap{}},
			{Key: "9001/tcp", V: OrdererMap{}},
		},
	}

	nullOrdererMap := OrdererMap{
		Pairs: []KV{
			{Key: "id", V: json.Number("123")},
			{Key: "name", V: nil},
			{Key: "lastname", V: ""},
			{Key: "jobs", V: []interface{}(nil)},
		},
	}

	simpleArrayOrdererMap := OrdererMap{
		Pairs: []KV{
			{Key: "items", V: []interface{}{
				OrdererMap{
					Pairs: []KV{
						{Key: "id", V: json.Number("2")},
						{Key: "name", V: "Item Two"},
					},
				},
			}},
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
		{"valid null values", "testdata/null.json", false, nullOrdererMap},
		{"valid invalid", "testdata/invalid.json", true, OrdererMap{}},
		{"valid simple array", "testdata/simple_array.json", false, simpleArrayOrdererMap},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := os.ReadFile(tt.inputFile)
			result, err := DecodeJSON(json.NewDecoder(bytes.NewReader(data)))

			if tt.shouldErr && err == nil {
				t.Errorf("expected error but got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.shouldErr {
				if diff := cmp.Diff(tt.expectedOrdererMap, result); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
