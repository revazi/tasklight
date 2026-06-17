package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestExecuteRunSuccess(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Execute([]string{"run", "--", "/bin/sh", "-c", "printf ok"}, strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if got := stdout.String(); got != "ok" {
		t.Fatalf("stdout = %q, want %q", got, "ok")
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
}

func TestExecuteRunPreservesFailureExitCode(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Execute([]string{"run", "--", "/bin/sh", "-c", "exit 42"}, strings.NewReader(""), &stdout, &stderr)

	if code != 42 {
		t.Fatalf("code = %d, want 42; stderr=%q", code, stderr.String())
	}
}

func TestExecuteRunMissingCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Execute([]string{"run"}, strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "missing command") {
		t.Fatalf("stderr = %q, want missing command", stderr.String())
	}
}

func TestExecuteRunRequiresSeparator(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Execute([]string{"run", "/bin/sh", "-c", "exit 0"}, strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "missing -- before command") {
		t.Fatalf("stderr = %q, want missing -- message", stderr.String())
	}
}

func TestExecuteRunHelp(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Execute([]string{"run", "--help"}, strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "tasklight run [options] -- <command>") {
		t.Fatalf("stdout = %q, want run help", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}
