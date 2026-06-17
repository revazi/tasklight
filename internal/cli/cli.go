package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"tasklight/internal/runner"
)

type runOptions struct {
	name string
	cwd  string
}

func Execute(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printRootHelp(stdout)
		return 0
	}

	switch args[0] {
	case "-h", "--help", "help":
		printRootHelp(stdout)
		return 0
	case "run":
		return executeRun(args[1:], stdin, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "tasklight: unknown command %q\n\n", args[0])
		printRootHelp(stderr)
		return 2
	}
}

func executeRun(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	opts, command, help, err := parseRunArgs(args)
	if help {
		printRunHelp(stdout)
		return 0
	}
	if err != nil {
		fmt.Fprintf(stderr, "tasklight run: %v\n\n", err)
		printRunHelp(stderr)
		return 2
	}

	result := runner.Run(context.Background(), runner.Options{
		Name:    opts.name,
		Command: command,
		Cwd:     opts.cwd,
		Stdin:   stdin,
		Stdout:  stdout,
		Stderr:  stderr,
	})

	if result.Err != nil && !result.Started {
		fmt.Fprintf(stderr, "tasklight run: failed to start %q: %v\n", command[0], result.Err)
	}

	return result.ExitCode
}

func parseRunArgs(args []string) (runOptions, []string, bool, error) {
	if len(args) == 0 {
		return runOptions{}, nil, false, errors.New("missing command; use: tasklight run -- <command>")
	}

	separator := indexOfSeparator(args)
	if separator == -1 {
		if hasHelpFlag(args) {
			return runOptions{}, nil, true, nil
		}
		return runOptions{}, nil, false, errors.New("missing -- before command; use: tasklight run -- <command>")
	}

	flagArgs := args[:separator]
	command := args[separator+1:]
	if len(command) == 0 {
		return runOptions{}, nil, false, errors.New("missing command after --")
	}

	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var opts runOptions
	fs.StringVar(&opts.name, "name", "", "human-readable task name")
	fs.StringVar(&opts.cwd, "cwd", "", "working directory for the command")

	if err := fs.Parse(flagArgs); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return runOptions{}, nil, true, nil
		}
		return runOptions{}, nil, false, err
	}
	if len(fs.Args()) > 0 {
		return runOptions{}, nil, false, fmt.Errorf("unexpected argument before --: %s", strings.Join(fs.Args(), " "))
	}

	return opts, command, false, nil
}

func indexOfSeparator(args []string) int {
	for i, arg := range args {
		if arg == "--" {
			return i
		}
	}
	return -1
}

func hasHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" || arg == "help" {
			return true
		}
	}
	return false
}

func printRootHelp(w io.Writer) {
	fmt.Fprint(w, `Tasklight watches long-running developer tasks.

Usage:
  tasklight <command> [options]

Commands:
  run       Run a command and preserve output, stdin, and exit code
  help      Show this help

Examples:
  tasklight run -- pnpm test
  tasklight run -- pytest
  tasklight run -- pi "fix this failing test"

Use "tasklight run --help" for run options.
`)
}

func printRunHelp(w io.Writer) {
	fmt.Fprint(w, `Run a command through Tasklight.

Usage:
  tasklight run [options] -- <command> [args...]

Options:
  --name string   Human-readable task name, used by later notifications
  --cwd string    Working directory for the command
  -h, --help      Show this help

Examples:
  tasklight run -- pnpm test
  tasklight run -- sh -c 'exit 42'
  tasklight run --cwd frontend -- pnpm build
`)
}
