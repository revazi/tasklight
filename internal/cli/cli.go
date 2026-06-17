package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/revazi/tasklight/internal/doctor"
	"github.com/revazi/tasklight/internal/notify"
	"github.com/revazi/tasklight/internal/runner"
	"github.com/revazi/tasklight/internal/session"
)

type runOptions struct {
	name        string
	cwd         string
	activateApp string
}

type notifyOptions struct {
	title       string
	subtitle    string
	message     string
	activateApp string
	iconPath    string
	sound       bool
}

var Version = "dev"

var detectFocusTarget = session.Detect

func Execute(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	return ExecuteWithNotifier(args, stdin, stdout, stderr, notify.DefaultNotifier())
}

func ExecuteWithNotifier(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, notifier notify.Notifier) int {
	if len(args) == 0 {
		printRootHelp(stdout)
		return 0
	}

	switch args[0] {
	case "-h", "--help", "help":
		printRootHelp(stdout)
		return 0
	case "-v", "--version", "version":
		fmt.Fprintf(stdout, "tasklight %s\n", Version)
		return 0
	case "run":
		return executeRun(args[1:], stdin, stdout, stderr, notifier)
	case "notify":
		return executeNotify(args[1:], stdout, stderr, notifier)
	case "doctor":
		return executeDoctor(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "tasklight: unknown command %q\n\n", args[0])
		printRootHelp(stderr)
		return 2
	}
}

func executeRun(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, notifier notify.Notifier) int {
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

	focusTarget := detectFocusTarget(session.DetectOptions{ActivateApp: opts.activateApp})

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

	if notifier != nil {
		if err := notifier.Notify(notificationForResult(result, focusTarget)); err != nil {
			fmt.Fprintf(stderr, "tasklight: warning: notification failed: %v\n", err)
		}
	}

	return result.ExitCode
}

func executeDoctor(args []string, stdout io.Writer, stderr io.Writer) int {
	if hasHelpFlag(args) {
		printDoctorHelp(stdout)
		return 0
	}
	if len(args) > 0 {
		fmt.Fprintf(stderr, "tasklight doctor: unexpected argument: %s\n\n", strings.Join(args, " "))
		printDoctorHelp(stderr)
		return 2
	}
	return doctor.Run(stdout)
}

func executeNotify(args []string, stdout io.Writer, stderr io.Writer, notifier notify.Notifier) int {
	opts, help, err := parseNotifyArgs(args)
	if help {
		printNotifyHelp(stdout)
		return 0
	}
	if err != nil {
		fmt.Fprintf(stderr, "tasklight notify: %v\n\n", err)
		printNotifyHelp(stderr)
		return 2
	}

	focusTarget := detectFocusTarget(session.DetectOptions{ActivateApp: opts.activateApp})
	notification := notify.Notification{
		Title:        opts.title,
		Subtitle:     opts.subtitle,
		Message:      opts.message,
		Sound:        opts.sound,
		ActivateApp:  focusTarget.ActivateApp,
		ClickCommand: focusTarget.ClickCommand(),
		IconPath:     opts.iconPath,
	}

	if notifier != nil {
		if err := notifier.Notify(notification); err != nil {
			fmt.Fprintf(stderr, "tasklight notify: notification failed: %v\n", err)
			return 1
		}
	}

	return 0
}

func notificationForResult(result runner.RunResult, focusTarget session.FocusTarget) notify.Notification {
	duration := formatDuration(result.EndedAt.Sub(result.StartedAt))
	notification := notify.Notification{
		Title:        "Tasklight",
		ActivateApp:  focusTarget.ActivateApp,
		ClickCommand: focusTarget.ClickCommand(),
	}

	if result.ExitCode == 0 && result.Err == nil {
		notification.Subtitle = fmt.Sprintf("✅ %s finished in %s", result.Name, duration)
		return notification
	}

	notification.Subtitle = fmt.Sprintf("❌ %s failed after %s", result.Name, duration)
	notification.Message = fmt.Sprintf("Exit code: %d", result.ExitCode)
	return notification
}

func formatDuration(duration time.Duration) string {
	if duration < 0 {
		duration = 0
	}
	if duration > 0 && duration < time.Second {
		return "<1s"
	}

	duration = duration.Round(time.Second)
	hours := int(duration / time.Hour)
	duration -= time.Duration(hours) * time.Hour
	minutes := int(duration / time.Minute)
	duration -= time.Duration(minutes) * time.Minute
	seconds := int(duration / time.Second)

	parts := make([]string, 0, 3)
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if seconds > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%ds", seconds))
	}

	return strings.Join(parts, " ")
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
	fs.StringVar(&opts.activateApp, "activate-app", "", "app name or bundle ID to activate when clicking the notification")

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

func parseNotifyArgs(args []string) (notifyOptions, bool, error) {
	if hasHelpFlag(args) {
		return notifyOptions{}, true, nil
	}

	fs := flag.NewFlagSet("notify", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	opts := notifyOptions{title: "Tasklight"}
	fs.StringVar(&opts.title, "title", opts.title, "notification title")
	fs.StringVar(&opts.subtitle, "subtitle", "", "notification subtitle")
	fs.StringVar(&opts.message, "message", "", "notification body/message")
	fs.StringVar(&opts.activateApp, "activate-app", "", "app name or bundle ID to activate when clicking the notification")
	fs.StringVar(&opts.iconPath, "icon", "", "path to a notification icon image")
	fs.BoolVar(&opts.sound, "sound", false, "play the platform's default notification sound when supported")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return notifyOptions{}, true, nil
		}
		return notifyOptions{}, false, err
	}
	if len(fs.Args()) > 0 {
		return notifyOptions{}, false, fmt.Errorf("unexpected argument: %s", strings.Join(fs.Args(), " "))
	}
	if opts.subtitle == "" && opts.message == "" {
		return notifyOptions{}, false, errors.New("missing notification text; provide --subtitle or --message")
	}

	return opts, false, nil
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
  notify    Send a Tasklight desktop notification
  doctor    Check notification/focus provider availability
  version   Show version
  help      Show this help

Examples:
  tasklight run -- pnpm test
  tasklight run -- pytest
  tasklight run -- pi "fix this failing test"
  tasklight notify --subtitle "✅ Tests finished" --message "All checks passed"
  tasklight doctor

Use "tasklight <command> --help" for command options.
`)
}

func printDoctorHelp(w io.Writer) {
	fmt.Fprint(w, `Check Tasklight notification and focus integration.

Usage:
  tasklight doctor

Checks:
  - platform notification provider availability
  - optional macOS terminal-notifier support
  - bundled Tasklight notification icon/sender app
  - optional tmux focus support
`)
}

func printNotifyHelp(w io.Writer) {
	fmt.Fprint(w, `Send a desktop notification through Tasklight.

Usage:
  tasklight notify [options]

Options:
  --title string          Notification title (default "Tasklight")
  --subtitle string       Notification subtitle
  --message string        Notification body/message
  --activate-app string   App name or bundle ID to activate when clicking the notification
  --icon string           Path to a notification icon image
  --sound                 Play the platform's default notification sound when supported
  -h, --help              Show this help

Examples:
  tasklight notify --subtitle "✅ Pi is ready" --message "Finished in 45s"
  tasklight notify --title "Pi" --subtitle "✅ Task finished" --message "Updated tests" --activate-app Terminal
`)
}

func printRunHelp(w io.Writer) {
	fmt.Fprint(w, `Run a command through Tasklight.

Usage:
  tasklight run [options] -- <command> [args...]

Options:
  --name string           Human-readable task name, used by notifications
  --cwd string            Working directory for the command
  --activate-app string   App name or bundle ID to activate when clicking the notification
  -h, --help              Show this help

Examples:
  tasklight run -- pnpm test
  tasklight run -- sh -c 'exit 42'
  tasklight run --cwd frontend -- pnpm build
`)
}
