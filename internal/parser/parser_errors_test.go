package parser

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// mockTokenReader implements tokenReader with pre-programmed token/error
// sequences. This lets tests exercise parser.go branches that json.Decoder
// itself can never produce (e.g. non-string object key, closing-delimiter
// read error, unexpected delimiter as first token).
type mockTokenReader struct {
	tokens []json.Token
	errs   []error // parallel to tokens; nil means no error for that call
	mores  []bool
	tIdx   int
	mIdx   int
}

func (m *mockTokenReader) Token() (json.Token, error) {
	if m.tIdx >= len(m.tokens) {
		return nil, errors.New("mock: no more tokens")
	}
	tok := m.tokens[m.tIdx]
	var err error
	if m.tIdx < len(m.errs) {
		err = m.errs[m.tIdx]
	}
	m.tIdx++
	return tok, err
}

func (m *mockTokenReader) More() bool {
	if m.mIdx >= len(m.mores) {
		return false
	}
	v := m.mores[m.mIdx]
	m.mIdx++
	return v
}

// errReader is an io.Reader that always returns an error after optionally
// emitting some initial bytes.
type errReader struct {
	prefix []byte
	pos    int
	err    error
}

func (r *errReader) Read(p []byte) (int, error) {
	if r.pos < len(r.prefix) {
		n := copy(p, r.prefix[r.pos:])
		r.pos += n
		return n, nil
	}
	return 0, r.err
}

// TestDecodeJSON_ReaderError verifies that DecodeJSON propagates an error from
// the underlying reader when no valid token can be read.
func TestDecodeJSON_ReaderError(t *testing.T) {
	// An empty reader causes json.Decoder.Token() to return io.EOF,
	// which DecodeJSON surfaces as an error.
	_, err := DecodeJSON(strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
}

// TestParseJSON_ObjectKeyTokenError exercises the error path triggered when
// reading the key token inside an object fails mid-stream.
func TestParseJSON_ObjectKeyTokenError(t *testing.T) {
	// Feed a JSON stream that starts a valid object but then truncates
	// immediately after the opening brace so that reading the key token fails.
	// We seed the decoder with `{"` which starts the key token but is never
	// finished — dec.Token() returns an error on the second call.
	dec := json.NewDecoder(strings.NewReader(`{`))
	dec.UseNumber()

	// Manually consume the '{' token so the decoder enters the object body.
	tok, err := dec.Token()
	if err != nil {
		t.Fatalf("setup: unexpected error reading '{': %v", err)
	}
	if tok != json.Delim('{') {
		t.Fatalf("setup: expected '{', got %v", tok)
	}

	// Now craft a decoder that will see More() == true but then fail on the
	// next Token() call. We do this by wrapping a second decoder that starts
	// partway through a key string so reading the token errors.
	decBroken := json.NewDecoder(strings.NewReader(`{"broken`))
	decBroken.UseNumber()

	// The first call to Token() will yield '{' (the object delimiter).
	// Calling ParseJSON on this decoder exercises the '{' branch, and the
	// first key token read (dec.Token()) will fail.
	_, parseErr := ParseJSON(decBroken)
	if parseErr == nil {
		t.Fatal("expected error for truncated object key, got nil")
	}
}

// TestParseJSON_ArrayValueError exercises the error path where ParseJSON fails
// while reading an element inside an array.
func TestParseJSON_ArrayValueError(t *testing.T) {
	// `["broken` starts an array and then starts a string token but cuts off.
	dec := json.NewDecoder(strings.NewReader(`["broken`))
	dec.UseNumber()

	_, err := ParseJSON(dec)
	if err == nil {
		t.Fatal("expected error for truncated array element, got nil")
	}
}

// TestParseJSON_ObjectClosingDelimiterError exercises the error path where
// reading the closing '}' token of an object fails.
//
// This path is reached only if the closing-delimiter read itself returns an
// error. In practice json.Decoder never returns an error on a well-formed
// closing brace, but we can verify the error-return by crafting a raw
// decoder scenario.  The coverage path in parser.go line 58-60 reads the
// token after all pairs have been consumed; if the reader errors at that point
// the error is surfaced.
//
// We exercise this by using the public API with a valid object so we at least
// confirm the non-error path is covered, and pair it with a test using
// DecodeJSON on a reader that errors mid-close.
func TestParseJSON_ObjectClosingDelimPath(t *testing.T) {
	// Confirm the happy-path closing-delim read works (also covered via other
	// tests, but included here for completeness of this group).
	data := `{"k":"v"}`
	result, err := DecodeJSON(bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	om, ok := result.(OrdererMap)
	if !ok {
		t.Fatalf("expected OrdererMap, got %T", result)
	}
	if len(om.Pairs) != 1 || om.Pairs[0].Key != "k" {
		t.Errorf("unexpected result: %+v", om)
	}
}

// TestParseJSON_ArrayClosingDelimPath exercises the array closing-delimiter
// read success path and verifies DecodeJSON returns a valid []any.
func TestParseJSON_ArrayClosingDelimPath(t *testing.T) {
	data := `[{"id":1}]`
	result, err := DecodeJSON(bytes.NewReader([]byte(data)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	arr, ok := result.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", result)
	}
	if len(arr) != 1 {
		t.Errorf("expected 1 element, got %d", len(arr))
	}
}

// TestParseJSON_DefaultBranch_NonDelim exercises the default (non-Delim) branch
// in ParseJSON by passing a decoder that yields a scalar token at the top level.
func TestParseJSON_DefaultBranch_NonDelim(t *testing.T) {
	dec := json.NewDecoder(strings.NewReader(`"hello"`))
	dec.UseNumber()

	result, err := ParseJSON(dec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello" {
		t.Errorf("expected \"hello\", got %v", result)
	}
}

// TestParseJSON_DefaultBranch_Number exercises the default branch with a
// json.Number token.
func TestParseJSON_DefaultBranch_Number(t *testing.T) {
	dec := json.NewDecoder(strings.NewReader(`42`))
	dec.UseNumber()

	result, err := ParseJSON(dec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	jn, ok := result.(json.Number)
	if !ok {
		t.Fatalf("expected json.Number, got %T", result)
	}
	if jn.String() != "42" {
		t.Errorf("expected 42, got %v", jn)
	}
}

// TestParseJSON_DefaultBranch_Bool exercises the default branch with a boolean.
func TestParseJSON_DefaultBranch_Bool(t *testing.T) {
	dec := json.NewDecoder(strings.NewReader(`true`))
	dec.UseNumber()

	result, err := ParseJSON(dec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, ok := result.(bool)
	if !ok {
		t.Fatalf("expected bool, got %T", result)
	}
	if !b {
		t.Errorf("expected true, got false")
	}
}

// TestParseJSON_DefaultBranch_Nil exercises the default branch with a null token.
func TestParseJSON_DefaultBranch_Nil(t *testing.T) {
	dec := json.NewDecoder(strings.NewReader(`null`))
	dec.UseNumber()

	result, err := ParseJSON(dec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

// TestParseJSON_NonStringKey exercises the !ok branch (L47-49) that fires when
// an object key token is not a string. json.Decoder never produces this, so we
// use a mockTokenReader that yields a json.Number as the key.
func TestParseJSON_NonStringKey(t *testing.T) {
	mock := &mockTokenReader{
		tokens: []json.Token{json.Delim('{'), json.Number("42")},
		errs:   []error{nil, nil},
		mores:  []bool{true},
	}
	_, err := ParseJSON(mock)
	if err == nil {
		t.Fatal("expected error for non-string key, got nil")
	}
	if !strings.Contains(err.Error(), "expected string key") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestParseJSON_ObjectClosingDelimError exercises the error path (L58-60) where
// reading the closing '}' token fails.
func TestParseJSON_ObjectClosingDelimError(t *testing.T) {
	closeErr := errors.New("close error")
	mock := &mockTokenReader{
		// '{' opens the object; More() returns false (empty); then Token() for
		// closing '}' returns an error.
		tokens: []json.Token{json.Delim('{'), nil},
		errs:   []error{nil, closeErr},
		mores:  []bool{false},
	}
	_, err := ParseJSON(mock)
	if err == nil {
		t.Fatal("expected error reading closing '}', got nil")
	}
	if !strings.Contains(err.Error(), "reading closing delimiter") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestParseJSON_ArrayClosingDelimError exercises the error path (L72-74) where
// reading the closing ']' token fails.
func TestParseJSON_ArrayClosingDelimError(t *testing.T) {
	closeErr := errors.New("close error")
	mock := &mockTokenReader{
		tokens: []json.Token{json.Delim('['), nil},
		errs:   []error{nil, closeErr},
		mores:  []bool{false},
	}
	_, err := ParseJSON(mock)
	if err == nil {
		t.Fatal("expected error reading closing ']', got nil")
	}
	if !strings.Contains(err.Error(), "reading closing delimiter") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestParseJSON_UnexpectedDelimiter exercises the default branch (L77-80) that
// fires when the first token is an unexpected delimiter such as '}' or ']'.
// json.Decoder never produces this at the top level, so we use a mock.
func TestParseJSON_UnexpectedDelimiter(t *testing.T) {
	mock := &mockTokenReader{
		tokens: []json.Token{json.Delim('}')},
		errs:   []error{nil},
	}
	_, err := ParseJSON(mock)
	if err == nil {
		t.Fatal("expected error for unexpected delimiter, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected delimiter") {
		t.Errorf("unexpected error message: %v", err)
	}
}
