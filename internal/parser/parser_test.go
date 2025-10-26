package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestParseJSON(t *testing.T) {
	tests := []struct {
		name      string
		inputFile string
		shouldErr bool
	}{
		{"valid simple", "testdata/simple.json", false},
		{"valid array", "testdata/array.json", false},
		{"valid nested", "testdata/nested.json", false},
		{"valid invalid", "testdata/invalid.json", true},
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
		})
	}
}
