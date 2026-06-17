package runner

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRunSuccessStreamsOutput(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	result := Run(context.Background(), Options{
		Command: []string{"/bin/sh", "-c", "printf 'hello'; printf 'warn' >&2"},
		Stdout:  &stdout,
		Stderr:  &stderr,
	})

	if result.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0 (err: %v)", result.ExitCode, result.Err)
	}
	if result.Err != nil {
		t.Fatalf("Err = %v, want nil", result.Err)
	}
	if got := stdout.String(); got != "hello" {
		t.Fatalf("stdout = %q, want %q", got, "hello")
	}
	if got := stderr.String(); got != "warn" {
		t.Fatalf("stderr = %q, want %q", got, "warn")
	}
	if result.Name != "sh" {
		t.Fatalf("Name = %q, want %q", result.Name, "sh")
	}
	if result.StartedAt.IsZero() || result.EndedAt.IsZero() || result.EndedAt.Before(result.StartedAt) {
		t.Fatalf("invalid timing: started=%v ended=%v", result.StartedAt, result.EndedAt)
	}
}

func TestRunFailurePreservesExitCode(t *testing.T) {
	result := Run(context.Background(), Options{
		Command: []string{"/bin/sh", "-c", "exit 42"},
	})

	if result.ExitCode != 42 {
		t.Fatalf("ExitCode = %d, want 42 (err: %v)", result.ExitCode, result.Err)
	}
	if result.Err == nil {
		t.Fatal("Err = nil, want non-nil")
	}
}

func TestRunMissingCommand(t *testing.T) {
	result := Run(context.Background(), Options{})

	if result.ExitCode != 2 {
		t.Fatalf("ExitCode = %d, want 2", result.ExitCode)
	}
	if !errors.Is(result.Err, ErrMissingCommand) {
		t.Fatalf("Err = %v, want ErrMissingCommand", result.Err)
	}
	if result.Started {
		t.Fatal("Started = true, want false")
	}
}

func TestRunWritesBeforeCommandExits(t *testing.T) {
	writer := &signalWriter{seen: make(chan struct{})}
	done := make(chan RunResult, 1)

	go func() {
		done <- Run(context.Background(), Options{
			Command: []string{"/bin/sh", "-c", "printf ready; sleep 0.2"},
			Stdout:  writer,
		})
	}()

	select {
	case <-writer.seen:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for live stdout")
	}

	select {
	case result := <-done:
		if result.ExitCode != 0 {
			t.Fatalf("ExitCode = %d, want 0 (err: %v)", result.ExitCode, result.Err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for command to exit")
	}

	if !strings.Contains(writer.String(), "ready") {
		t.Fatalf("stdout = %q, want to contain ready", writer.String())
	}
}

type signalWriter struct {
	buf  bytes.Buffer
	seen chan struct{}
}

func (w *signalWriter) Write(p []byte) (int, error) {
	if w.buf.Len() == 0 {
		close(w.seen)
	}
	return w.buf.Write(p)
}

func (w *signalWriter) String() string {
	return w.buf.String()
}
