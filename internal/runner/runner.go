package runner

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"syscall"
	"time"
)

var ErrMissingCommand = errors.New("missing command")

type Options struct {
	Name         string
	Command      []string
	Cwd          string
	Stdin        io.Reader
	Stdout       io.Writer
	Stderr       io.Writer
	IdleTimeout  time.Duration
	OnIdle       func(IdleEvent)
	MatchPattern *regexp.Regexp
	OnMatch      func(MatchEvent)
}

type IdleEvent struct {
	Name    string
	IdleFor time.Duration
}

type MatchEvent struct {
	Name    string
	Pattern string
	Stream  string
	Line    string
}

type RunResult struct {
	Name      string
	Command   []string
	Started   bool
	StartedAt time.Time
	EndedAt   time.Time
	ExitCode  int
	Err       error
}

func Run(ctx context.Context, opts Options) RunResult {
	startedAt := time.Now()
	result := RunResult{
		Name:      taskName(opts.Name, opts.Command),
		Command:   append([]string(nil), opts.Command...),
		StartedAt: startedAt,
		ExitCode:  0,
	}

	if len(opts.Command) == 0 {
		result.EndedAt = time.Now()
		result.ExitCode = 2
		result.Err = ErrMissingCommand
		return result
	}

	cmd := exec.CommandContext(ctx, opts.Command[0], opts.Command[1:]...)
	cmd.Dir = opts.Cwd
	cmd.Stdin = opts.Stdin

	stdout := writerOrDiscard(opts.Stdout)
	stderr := writerOrDiscard(opts.Stderr)
	var idle *idleMonitor
	if opts.IdleTimeout > 0 && opts.OnIdle != nil {
		idle = newIdleMonitor(opts.IdleTimeout, func() {
			opts.OnIdle(IdleEvent{Name: result.Name, IdleFor: opts.IdleTimeout})
		})
		stdout = activityWriter{writer: stdout, touch: idle.touch}
		stderr = activityWriter{writer: stderr, touch: idle.touch}
	}

	var matchWriters []*lineMatchWriter
	if opts.MatchPattern != nil && opts.OnMatch != nil {
		sink := &matchSink{onMatch: opts.OnMatch}
		stdoutMatcher := &lineMatchWriter{writer: stdout, stream: "stdout", name: result.Name, pattern: opts.MatchPattern, sink: sink}
		stderrMatcher := &lineMatchWriter{writer: stderr, stream: "stderr", name: result.Name, pattern: opts.MatchPattern, sink: sink}
		stdout = stdoutMatcher
		stderr = stderrMatcher
		matchWriters = []*lineMatchWriter{stdoutMatcher, stderrMatcher}
	}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		result.EndedAt = time.Now()
		result.ExitCode = startErrorExitCode(err)
		result.Err = err
		return result
	}
	result.Started = true
	if idle != nil {
		idle.start(newSystemIdleTimer)
	}

	signals := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		defer signal.Stop(signals)
		for {
			select {
			case sig := <-signals:
				if cmd.Process != nil {
					_ = cmd.Process.Signal(sig)
				}
			case <-done:
				return
			}
		}
	}()

	err := cmd.Wait()
	if idle != nil {
		idle.close()
	}
	for _, writer := range matchWriters {
		writer.flush()
	}
	close(done)

	result.EndedAt = time.Now()
	result.Err = err
	result.ExitCode = exitCode(err)
	return result
}

func writerOrDiscard(w io.Writer) io.Writer {
	if w == nil {
		return io.Discard
	}
	return w
}

func taskName(name string, command []string) string {
	if name != "" {
		return name
	}
	if len(command) == 0 || command[0] == "" {
		return "task"
	}
	return filepath.Base(command[0])
}

func startErrorExitCode(err error) int {
	if errors.Is(err, exec.ErrNotFound) {
		return 127
	}
	return 1
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			if status.Exited() {
				return status.ExitStatus()
			}
			if status.Signaled() {
				return 128 + int(status.Signal())
			}
		}

		code := exitErr.ExitCode()
		if code >= 0 {
			return code
		}
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return 1
	}

	return 1
}
