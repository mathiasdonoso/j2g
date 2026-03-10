package main

import (
	"bytes"
	"flag"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestIsDebugMode verifies that isDebugMode returns true only when DEBUG=1.
func TestIsDebugMode(t *testing.T) {
	tests := []struct {
		envVal string
		want   bool
	}{
		{"1", true},
		{"0", false},
		{"", false},
		{"2", false},
		{"true", false},
	}

	for _, tt := range tests {
		t.Run("DEBUG="+tt.envVal, func(t *testing.T) {
			t.Setenv("DEBUG", tt.envVal)
			if got := isDebugMode(); got != tt.want {
				t.Errorf("isDebugMode() = %v, want %v (DEBUG=%q)", got, tt.want, tt.envVal)
			}
		})
	}
}

// TestInitLogger_DebugDisabled verifies that initLogger is a no-op when debug
// mode is off (DEBUG != 1). We confirm the default logger is not replaced.
func TestInitLogger_DebugDisabled(t *testing.T) {
	t.Setenv("DEBUG", "0")
	before := slog.Default()
	initLogger()
	after := slog.Default()
	// Logger should remain the same instance when debug mode is off.
	if before != after {
		t.Error("initLogger replaced the default logger when debug mode was off")
	}
}

// TestInitLogger_DebugEnabled verifies that initLogger installs a new slog
// handler when DEBUG=1.
func TestInitLogger_DebugEnabled(t *testing.T) {
	t.Setenv("DEBUG", "1")
	before := slog.Default()
	initLogger()
	after := slog.Default()
	// Logger should have been replaced when debug mode is on.
	if before == after {
		t.Error("initLogger did not replace the default logger when debug mode was on")
	}
	// Restore default logger to avoid polluting other tests.
	slog.SetDefault(before)
}

// TestShowErrorMessage_DebugEnabled verifies that showErrorMessage prints
// nothing when DEBUG=1 (the error text is suppressed so it doesn't interfere
// with structured debug output).
func TestShowErrorMessage_DebugEnabled(t *testing.T) {
	t.Setenv("DEBUG", "1")

	// Capture stdout by redirecting os.Stdout.
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	showErrorMessage()

	w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	if buf.Len() != 0 {
		t.Errorf("showErrorMessage printed output in debug mode: %q", buf.String())
	}
}

// TestShowErrorMessage_DebugDisabled verifies that showErrorMessage prints the
// ErrorText when debug mode is off.
func TestShowErrorMessage_DebugDisabled(t *testing.T) {
	t.Setenv("DEBUG", "0")

	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	showErrorMessage()

	w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	if buf.Len() == 0 {
		t.Error("showErrorMessage printed nothing when debug mode was off")
	}
}

// resetFlags resets flag.CommandLine so that checkFlags (which registers -h
// and --help on the global FlagSet) can be called more than once in a single
// test binary run without panicking with "flag redefined".
func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

// TestCheckFlags_NoFlags verifies that checkFlags does not exit when no
// help flags are present. We pass an empty args list by resetting os.Args.
func TestCheckFlags_NoFlags(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() {
		os.Args = origArgs
		resetFlags()
	})
	resetFlags()
	os.Args = []string{"j2g"}

	// If checkFlags panics or calls os.Exit this test will fail or the process
	// will terminate. Running without -h should be a no-op.
	checkFlags()
}

// TestMain_HelpFlag verifies that running the binary with -h prints usage text
// and exits with code 0.  This is a subprocess test: it re-executes the test
// binary in a special mode to exercise the checkFlags → os.Exit(0) path and
// the main() function itself.
func TestMain_HelpFlag(t *testing.T) {
	if os.Getenv("J2G_TEST_SUBPROCESS") == "help" {
		// We are the subprocess: run main() with -h.
		os.Args = []string{"j2g", "-h"}
		main()
		// main() calls os.Exit via checkFlags; if we reach here the test is wrong.
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_HelpFlag")
	cmd.Env = append(os.Environ(), "J2G_TEST_SUBPROCESS=help")
	out, err := cmd.CombinedOutput()
	// -h causes os.Exit(0), so err should be nil.
	if err != nil {
		t.Fatalf("subprocess exited with error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(string(out), "j2g") {
		t.Errorf("expected usage text in output, got:\n%s", out)
	}
}

// TestMain_ValidInput verifies that main() processes valid JSON from stdin and
// writes Go struct output to stdout. It calls main() in-process (redirecting
// os.Stdin and os.Stdout) so that the defer body inside main() is covered by
// the instrumented test binary's coverage profile.
func TestMain_ValidInput(t *testing.T) {
	// Redirect os.Stdin to a pipe containing valid JSON.
	origStdin := os.Stdin
	origStdout := os.Stdout
	origArgs := os.Args
	t.Cleanup(func() {
		os.Stdin = origStdin
		os.Stdout = origStdout
		os.Args = origArgs
		resetFlags()
	})
	resetFlags()

	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating stdin pipe: %v", err)
	}
	stdinW.WriteString(`{"id": 1, "name": "Alice"}`)
	stdinW.Close()
	os.Stdin = stdinR

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating stdout pipe: %v", err)
	}
	os.Stdout = stdoutW

	os.Args = []string{"j2g"}
	main()

	stdoutW.Close()

	var buf bytes.Buffer
	buf.ReadFrom(stdoutR)

	if !strings.Contains(buf.String(), "type Result struct") {
		t.Errorf("expected struct output, got:\n%s", buf.String())
	}
}

// TestMain_InvalidInput verifies that main() exits with code 1 and prints an
// error message when the JSON input is invalid.
func TestMain_InvalidInput(t *testing.T) {
	if os.Getenv("J2G_TEST_SUBPROCESS") == "invalid" {
		r, w, err := os.Pipe()
		if err != nil {
			os.Exit(2)
		}
		w.WriteString(`{broken`)
		w.Close()
		os.Stdin = r
		os.Args = []string{"j2g"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_InvalidInput")
	cmd.Env = append(os.Environ(), "J2G_TEST_SUBPROCESS=invalid")
	out, err := cmd.CombinedOutput()
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected exit error, got: %v\noutput: %s", err, out)
	}
	if exitErr.ExitCode() != 1 {
		t.Errorf("expected exit code 1, got %d\noutput: %s", exitErr.ExitCode(), out)
	}
}
