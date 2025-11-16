package integration

import (
	"bytes"
	"go/format"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mathiasdonoso/j2g/internal/cli"
)

func TestCli(t *testing.T) {
	tests := []struct {
		name       string
		inputFile  string
		outputFile string
		shouldErr  bool
	}{
		{"invalid", "testdata/input/invalid", "", true},
		{"simple", "testdata/input/simple.json", "testdata/output/simple.txt", false},
		{"simple_array", "testdata/input/simple.json", "testdata/output/simple.txt", false},
		{"array", "testdata/input/array.json", "testdata/output/array.txt", false},
		{"camelcase", "testdata/input/camelcase.json", "testdata/output/camelcase.txt", false},
		{"nested", "testdata/input/nested.json", "testdata/output/nested.txt", false},
		{"null", "testdata/input/null.json", "testdata/output/null.txt", false},
		{"with_numbers", "testdata/input/with_numbers.json", "testdata/output/with_numbers.txt", false},
		{"type_array", "testdata/input/type_array.json", "testdata/output/type_array.txt", false},
		{"empty_array", "testdata/input/empty_array.json", "testdata/output/empty_array.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, _ := os.ReadFile(tt.inputFile)
			expectedOutput, _ := os.ReadFile(tt.outputFile)
			var out bytes.Buffer

			j2g := cli.J2G{
				Input:  bytes.NewReader(input),
				Output: &out,
			}

			err := j2g.Start()

			if tt.shouldErr && err == nil {
				t.Errorf("expected error but got nil")
			}

			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.shouldErr && err == nil {
				// Strip trailing newline introduced by POSIX text-file semantics.
				// Editors routinely append a final '\n' even when it's not visible,
				// but Postman response bodies don't include it—so we normalize here.
				formattedWant, _ := format.Source([]byte(expectedOutput))
				formattedGot, _ := format.Source([]byte(out.Bytes()))
				sWant := strings.TrimSpace(string(formattedWant))
				sGot := strings.TrimSpace(string(formattedGot))

				if diff := cmp.Diff(sWant, sGot); diff != "" {
					t.Errorf("output mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
