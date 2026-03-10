package generator

import (
	"encoding/json"
	"testing"

	"github.com/mathiasdonoso/j2g/internal/parser"
)

// TestInferType_EmptyArray verifies that an empty []any value is typed as []any.
func TestInferType_EmptyArray(t *testing.T) {
	b := Builder{}
	goType, nested, err := b.inferType("Items", []any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if goType != "[]any" {
		t.Errorf("expected []any, got %q", goType)
	}
	if len(nested) != 0 {
		t.Errorf("expected no nested defs, got %v", nested)
	}
}

// TestInferType_ScalarArray verifies that an array of scalars gets the
// concrete element type (e.g. []string, []json.Number).
func TestInferType_ScalarArray(t *testing.T) {
	b := Builder{}

	tests := []struct {
		name     string
		value    []any
		wantType string
	}{
		{"string slice", []any{"a", "b"}, "[]string"},
		{"number slice", []any{json.Number("1"), json.Number("2")}, "[]json.Number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goType, nested, err := b.inferType("Values", tt.value)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if goType != tt.wantType {
				t.Errorf("expected %q, got %q", tt.wantType, goType)
			}
			if len(nested) != 0 {
				t.Errorf("expected no nested defs, got %v", nested)
			}
		})
	}
}

// TestInferType_SingleCharKeyName exercises the branch where keyName has
// length == 1, so elemTypeName becomes "<key>Type" instead of singularizing.
func TestInferType_SingleCharKeyName(t *testing.T) {
	b := Builder{}
	arr := []any{
		parser.OrdererMap{
			Pairs: []parser.KV{
				{Key: "id", V: json.Number("1")},
			},
		},
	}

	goType, nested, err := b.inferType("A", arr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if goType != "[]AType" {
		t.Errorf("expected []AType, got %q", goType)
	}
	if len(nested) != 1 {
		t.Errorf("expected 1 nested def, got %d", len(nested))
	}
}

// TestInferType_JsonNumber_Float verifies that a json.Number containing a
// decimal point is typed as float64.
func TestInferType_JsonNumber_Float(t *testing.T) {
	b := Builder{}
	goType, nested, err := b.inferType("Price", json.Number("3.14"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if goType != "float64" {
		t.Errorf("expected float64, got %q", goType)
	}
	if len(nested) != 0 {
		t.Errorf("expected no nested defs, got %v", nested)
	}
}

// TestInferType_JsonNumber_Int verifies that a json.Number without a decimal
// point is typed as int.
func TestInferType_JsonNumber_Int(t *testing.T) {
	b := Builder{}
	goType, nested, err := b.inferType("Count", json.Number("42"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if goType != "int" {
		t.Errorf("expected int, got %q", goType)
	}
	if len(nested) != 0 {
		t.Errorf("expected no nested defs, got %v", nested)
	}
}

// TestBuildStruct_WithJsonNumbers verifies that BuildStruct correctly infers
// int and float64 types from json.Number values (as produced by the parser).
func TestBuildStruct_WithJsonNumbers(t *testing.T) {
	input := parser.OrdererMap{
		Pairs: []parser.KV{
			{Key: "count", V: json.Number("5")},
			{Key: "price", V: json.Number("9.99")},
		},
	}

	b := Builder{}
	result, err := b.BuildStruct(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "type Result struct {\n\tCount int `json:\"count\"`\n\tPrice float64 `json:\"price\"`\n}"
	if result != want {
		t.Errorf("expected:\n%s\ngot:\n%s", want, result)
	}
}

// TestBuildStruct_EmptyKey verifies that BuildStruct returns an error when a
// JSON key normalizes to an empty identifier (e.g. "---"). This exercises the
// error path added to fix the latent panic at keyName[0], and also makes the
// error-return chains in inferType and BuildFromArray reachable.
func TestBuildStruct_EmptyKey(t *testing.T) {
	input := parser.OrdererMap{
		Pairs: []parser.KV{{Key: "---", V: "value"}},
	}
	b := Builder{}
	_, err := b.BuildStruct(input)
	if err == nil {
		t.Fatal("expected error for key that normalizes to empty, got nil")
	}
}

// TestInferType_NestedEmptyKey verifies that the error from BuildStruct
// propagates through the inferType nested-object path (L46-48).
func TestInferType_NestedEmptyKey(t *testing.T) {
	nested := parser.OrdererMap{
		Pairs: []parser.KV{{Key: "---", V: "value"}},
	}
	b := Builder{}
	_, _, err := b.inferType("Address", nested)
	if err == nil {
		t.Fatal("expected error propagated from nested BuildStruct, got nil")
	}
}

// TestInferType_ArrayNestedEmptyKey verifies that the error from BuildStruct
// propagates through the inferType array path (L91-93).
func TestInferType_ArrayNestedEmptyKey(t *testing.T) {
	badElem := parser.OrdererMap{
		Pairs: []parser.KV{{Key: "---", V: "value"}},
	}
	b := Builder{}
	_, _, err := b.inferType("Items", []any{badElem})
	if err == nil {
		t.Fatal("expected error propagated from array element BuildStruct, got nil")
	}
}

// TestBuildFromArray_EmptyKey verifies that BuildFromArray propagates a
// BuildStruct error when the root array element has an un-normalizable key (L144-146).
func TestBuildFromArray_EmptyKey(t *testing.T) {
	badElem := parser.OrdererMap{
		Pairs: []parser.KV{{Key: "---", V: "value"}},
	}
	b := Builder{}
	_, err := b.BuildFromArray([]any{badElem})
	if err == nil {
		t.Fatal("expected error from BuildFromArray with bad key, got nil")
	}
}

// TestBuildStruct_NestedEmptyKey verifies that the error from a nested
// BuildStruct call propagates through BuildStruct's own inferType call (L167-169).
func TestBuildStruct_NestedEmptyKey(t *testing.T) {
	// Outer object has a field "address" whose value is an inner object with an
	// un-normalizable key. BuildStruct calls inferType("Address", innerMap),
	// which calls nestedBuilder.BuildStruct(innerMap), which errors, and that
	// error surfaces back through BuildStruct's if err != nil check.
	inner := parser.OrdererMap{
		Pairs: []parser.KV{{Key: "---", V: "value"}},
	}
	outer := parser.OrdererMap{
		Pairs: []parser.KV{{Key: "address", V: inner}},
	}
	b := Builder{}
	_, err := b.BuildStruct(outer)
	if err == nil {
		t.Fatal("expected error propagated through BuildStruct → inferType → BuildStruct, got nil")
	}
}

// TestBuildStruct_EmptyArray verifies that an empty []any field is typed as []any.
func TestBuildStruct_EmptyArray(t *testing.T) {
	input := parser.OrdererMap{
		Pairs: []parser.KV{
			{Key: "items", V: []any{}},
		},
	}

	b := Builder{}
	result, err := b.BuildStruct(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "type Result struct {\n\tItems []any `json:\"items\"`\n}"
	if result != want {
		t.Errorf("expected:\n%s\ngot:\n%s", want, result)
	}
}
