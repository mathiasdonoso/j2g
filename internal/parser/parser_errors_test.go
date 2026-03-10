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

func TestDecodeJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "empty input causes EOF error",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeJSON(strings.NewReader(tt.input))
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestParseJSON(t *testing.T) {
	tests := []struct {
		name             string
		makeDecoder      func() tokenReader
		wantErr          bool
		wantErrContains  string
		wantResult       any
		checkResult      func(t *testing.T, result any)
	}{
		{
			name: "truncated object key returns error",
			makeDecoder: func() tokenReader {
				dec := json.NewDecoder(strings.NewReader(`{"broken`))
				dec.UseNumber()
				return dec
			},
			wantErr: true,
		},
		{
			name: "truncated array element returns error",
			makeDecoder: func() tokenReader {
				dec := json.NewDecoder(strings.NewReader(`["broken`))
				dec.UseNumber()
				return dec
			},
			wantErr: true,
		},
		{
			name: "object closing delimiter path — happy path",
			makeDecoder: func() tokenReader {
				dec := json.NewDecoder(bytes.NewReader([]byte(`{"k":"v"}`)))
				dec.UseNumber()
				return dec
			},
			wantErr: false,
			checkResult: func(t *testing.T, result any) {
				t.Helper()
				om, ok := result.(OrdererMap)
				if !ok {
					t.Fatalf("expected OrdererMap, got %T", result)
				}
				if len(om.Pairs) != 1 || om.Pairs[0].Key != "k" {
					t.Errorf("unexpected result: %+v", om)
				}
			},
		},
		{
			name: "array closing delimiter path — happy path",
			makeDecoder: func() tokenReader {
				dec := json.NewDecoder(bytes.NewReader([]byte(`[{"id":1}]`)))
				dec.UseNumber()
				return dec
			},
			wantErr: false,
			checkResult: func(t *testing.T, result any) {
				t.Helper()
				arr, ok := result.([]any)
				if !ok {
					t.Fatalf("expected []any, got %T", result)
				}
				if len(arr) != 1 {
					t.Errorf("expected 1 element, got %d", len(arr))
				}
			},
		},
		{
			name: "default branch — string scalar",
			makeDecoder: func() tokenReader {
				dec := json.NewDecoder(strings.NewReader(`"hello"`))
				dec.UseNumber()
				return dec
			},
			wantErr:    false,
			wantResult: "hello",
		},
		{
			name: "default branch — json.Number",
			makeDecoder: func() tokenReader {
				dec := json.NewDecoder(strings.NewReader(`42`))
				dec.UseNumber()
				return dec
			},
			wantErr: false,
			checkResult: func(t *testing.T, result any) {
				t.Helper()
				jn, ok := result.(json.Number)
				if !ok {
					t.Fatalf("expected json.Number, got %T", result)
				}
				if jn.String() != "42" {
					t.Errorf("expected 42, got %v", jn)
				}
			},
		},
		{
			name: "default branch — bool",
			makeDecoder: func() tokenReader {
				dec := json.NewDecoder(strings.NewReader(`true`))
				dec.UseNumber()
				return dec
			},
			wantErr:    false,
			wantResult: true,
		},
		{
			name: "default branch — null",
			makeDecoder: func() tokenReader {
				dec := json.NewDecoder(strings.NewReader(`null`))
				dec.UseNumber()
				return dec
			},
			wantErr:    false,
			wantResult: nil,
		},
		{
			name: "non-string object key returns error",
			makeDecoder: func() tokenReader {
				return &mockTokenReader{
					tokens: []json.Token{json.Delim('{'), json.Number("42")},
					errs:   []error{nil, nil},
					mores:  []bool{true},
				}
			},
			wantErr:         true,
			wantErrContains: "expected string key",
		},
		{
			name: "object closing delimiter error is propagated",
			makeDecoder: func() tokenReader {
				return &mockTokenReader{
					// '{' opens the object; More() returns false (empty); then Token()
					// for closing '}' returns an error.
					tokens: []json.Token{json.Delim('{'), nil},
					errs:   []error{nil, errors.New("close error")},
					mores:  []bool{false},
				}
			},
			wantErr:         true,
			wantErrContains: "reading closing delimiter",
		},
		{
			name: "array closing delimiter error is propagated",
			makeDecoder: func() tokenReader {
				return &mockTokenReader{
					tokens: []json.Token{json.Delim('['), nil},
					errs:   []error{nil, errors.New("close error")},
					mores:  []bool{false},
				}
			},
			wantErr:         true,
			wantErrContains: "reading closing delimiter",
		},
		{
			name: "unexpected delimiter returns error",
			makeDecoder: func() tokenReader {
				return &mockTokenReader{
					tokens: []json.Token{json.Delim('}')},
					errs:   []error{nil},
				}
			},
			wantErr:         true,
			wantErrContains: "unexpected delimiter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseJSON(tt.makeDecoder())

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrContains != "" && !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.wantErrContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantResult != nil {
				if result != tt.wantResult {
					t.Errorf("expected %v, got %v", tt.wantResult, result)
				}
			}

			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}
