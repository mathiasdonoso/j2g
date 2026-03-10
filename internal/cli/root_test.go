package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestStart_RootObject(t *testing.T) {
	input := `{"id": 1, "name": "Alice"}`
	var out bytes.Buffer

	j2g := J2G{
		Input:  strings.NewReader(input),
		Output: &out,
	}

	if err := j2g.Start(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "type Result struct") {
		t.Errorf("expected output to contain 'type Result struct', got:\n%s", got)
	}
}

func TestStart_RootArray(t *testing.T) {
	input := `[{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]`
	var out bytes.Buffer

	j2g := J2G{
		Input:  strings.NewReader(input),
		Output: &out,
	}

	if err := j2g.Start(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "type Result struct") {
		t.Errorf("expected output to contain 'type Result struct', got:\n%s", got)
	}
	if !strings.Contains(got, "type Results []Result") {
		t.Errorf("expected output to contain 'type Results []Result', got:\n%s", got)
	}
}

func TestStart_RootArray_ScalarElements(t *testing.T) {
	input := `[1, 2, 3]`
	var out bytes.Buffer

	j2g := J2G{
		Input:  strings.NewReader(input),
		Output: &out,
	}

	err := j2g.Start()
	if err == nil {
		t.Fatalf("expected error for scalar root array, got nil")
	}
	if !strings.Contains(err.Error(), "root array element is not a JSON object") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestStart_RootArray_Empty(t *testing.T) {
	input := `[]`
	var out bytes.Buffer

	j2g := J2G{
		Input:  strings.NewReader(input),
		Output: &out,
	}

	if err := j2g.Start(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "type Results []interface{}") {
		t.Errorf("expected output to contain 'type Results []interface{}', got:\n%s", got)
	}
}
