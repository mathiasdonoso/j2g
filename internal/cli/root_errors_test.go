package cli

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// errWriter is an io.Writer that always returns an error.
type errWriter struct{ err error }

func (w *errWriter) Write(p []byte) (int, error) { return 0, w.err }

// TestStart_InvalidJSON verifies that Start surfaces the parse error when the
// input is invalid JSON.
func TestStart_InvalidJSON(t *testing.T) {
	j2g := J2G{
		Input:  strings.NewReader(`{broken`),
		Output: &bytes.Buffer{},
	}
	err := j2g.Start()
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

// TestStart_UnsupportedRootType verifies that Start returns an error when
// DecodeJSON produces a value that is neither an OrdererMap nor a []any.
// A bare JSON scalar (e.g. a quoted string at root level) triggers the
// default branch in the type switch.
func TestStart_UnsupportedRootType(t *testing.T) {
	// A bare JSON string is a valid JSON document whose parsed type is
	// `string`, which matches neither OrdererMap nor []any.
	j2g := J2G{
		Input:  strings.NewReader(`"hello"`),
		Output: &bytes.Buffer{},
	}
	err := j2g.Start()
	if err == nil {
		t.Fatal("expected error for unsupported root type, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported root JSON type") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestStart_WriteError verifies that Start propagates an error returned by the
// output writer.
func TestStart_WriteError(t *testing.T) {
	writeErr := fmt.Errorf("write failed")
	j2g := J2G{
		Input:  strings.NewReader(`{"id": 1}`),
		Output: &errWriter{err: writeErr},
	}
	err := j2g.Start()
	if err == nil {
		t.Fatal("expected write error, got nil")
	}
	if !strings.Contains(err.Error(), "write failed") {
		t.Errorf("unexpected error message: %v", err)
	}
}
