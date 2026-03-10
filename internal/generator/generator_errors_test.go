package generator

import (
	"encoding/json"
	"testing"

	"github.com/mathiasdonoso/j2g/internal/parser"
)

func TestInferType(t *testing.T) {
	tests := []struct {
		name            string
		keyName         string
		value           any
		wantType        string
		wantNestedCount int
		wantErr         bool
	}{
		{
			name:            "empty array is typed as []any",
			keyName:         "Items",
			value:           []any{},
			wantType:        "[]any",
			wantNestedCount: 0,
		},
		{
			name:            "scalar string array",
			keyName:         "Values",
			value:           []any{"a", "b"},
			wantType:        "[]string",
			wantNestedCount: 0,
		},
		{
			name:            "scalar number array",
			keyName:         "Values",
			value:           []any{json.Number("1"), json.Number("2")},
			wantType:        "[]json.Number",
			wantNestedCount: 0,
		},
		{
			name:    "single-char key name uses <key>Type as element type",
			keyName: "A",
			value: []any{
				parser.OrdererMap{
					Pairs: []parser.KV{
						{Key: "id", V: json.Number("1")},
					},
				},
			},
			wantType:        "[]AType",
			wantNestedCount: 1,
		},
		{
			name:            "json.Number with decimal is float64",
			keyName:         "Price",
			value:           json.Number("3.14"),
			wantType:        "float64",
			wantNestedCount: 0,
		},
		{
			name:            "json.Number without decimal is int",
			keyName:         "Count",
			value:           json.Number("42"),
			wantType:        "int",
			wantNestedCount: 0,
		},
		{
			name:    "nested object with empty key propagates error",
			keyName: "Address",
			value: parser.OrdererMap{
				Pairs: []parser.KV{{Key: "---", V: "value"}},
			},
			wantErr: true,
		},
		{
			name:    "array with element containing empty key propagates error",
			keyName: "Items",
			value: []any{
				parser.OrdererMap{
					Pairs: []parser.KV{{Key: "---", V: "value"}},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Builder{}
			goType, nested, err := b.inferType(tt.keyName, tt.value)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if goType != tt.wantType {
				t.Errorf("expected type %q, got %q", tt.wantType, goType)
			}
			if len(nested) != tt.wantNestedCount {
				t.Errorf("expected %d nested defs, got %d: %v", tt.wantNestedCount, len(nested), nested)
			}
		})
	}
}

func TestBuildStruct_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   parser.OrdererMap
		want    string
		wantErr bool
	}{
		{
			name: "json numbers produce int and float64 fields",
			input: parser.OrdererMap{
				Pairs: []parser.KV{
					{Key: "count", V: json.Number("5")},
					{Key: "price", V: json.Number("9.99")},
				},
			},
			want: "type Result struct {\n\tCount int `json:\"count\"`\n\tPrice float64 `json:\"price\"`\n}",
		},
		{
			name: "empty array field is typed as []any",
			input: parser.OrdererMap{
				Pairs: []parser.KV{
					{Key: "items", V: []any{}},
				},
			},
			want: "type Result struct {\n\tItems []any `json:\"items\"`\n}",
		},
		{
			name: "key that normalizes to empty returns error",
			input: parser.OrdererMap{
				Pairs: []parser.KV{{Key: "---", V: "value"}},
			},
			wantErr: true,
		},
		{
			name: "nested object with empty key propagates error",
			input: parser.OrdererMap{
				Pairs: []parser.KV{
					{
						Key: "address",
						V: parser.OrdererMap{
							Pairs: []parser.KV{{Key: "---", V: "value"}},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Builder{}
			result, err := b.BuildStruct(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.want {
				t.Errorf("expected:\n%s\ngot:\n%s", tt.want, result)
			}
		})
	}
}

func TestBuildFromArray_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   []any
		wantErr bool
	}{
		{
			name: "element with key normalizing to empty returns error",
			input: []any{
				parser.OrdererMap{
					Pairs: []parser.KV{{Key: "---", V: "value"}},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Builder{}
			_, err := b.BuildFromArray(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
