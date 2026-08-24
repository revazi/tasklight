package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	taskconfig "github.com/revazi/tasklight/internal/config"
	"github.com/revazi/tasklight/internal/doctor"
	"github.com/revazi/tasklight/internal/notify"
	"github.com/revazi/tasklight/internal/runner"
	"github.com/revazi/tasklight/internal/session"
)

type runOptions struct {
	name        string
	cwd         string
	activateApp string
	idle        time.Duration
	match       *regexp.Regexp
	sound       bool
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

var (
	detectFocusTarget = session.Detect
	loadConfig        = taskconfig.Load
	runDoctor         = doctor.Run
	runFocusDoctor    = doctor.RunFocus
)

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
	if hasHelpFlag(args) {
		printRunHelp(stdout)
		return 0
	}
	configuration, err := loadConfig()
	if err != nil {
		fmt.Fprintf(stderr, "tasklight run: configuration error: %v\n", err)
		return 2
	}
	opts, command, help, err := parseRunArgs(args, configuration.Run)
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
	var notificationMu sync.Mutex
	sendNotification := func(notification notify.Notification) error {
		if notifier == nil {
			return nil
		}
		notificationMu.Lock()
		defer notificationMu.Unlock()
		return notifier.Notify(notification)
	}

	runOpts := runner.Options{
		Name:    opts.name,
		Command: command,
		Cwd:     opts.cwd,
		Stdin:   stdin,
		Stdout:  stdout,
		Stderr:  stderr,
	}
	if opts.idle > 0 && notifier != nil {
		runOpts.IdleTimeout = opts.idle
		runOpts.OnIdle = func(event runner.IdleEvent) {
			if err := sendNotification(notificationForIdle(event, focusTarget, opts.sound)); err != nil {
				fmt.Fprintf(stderr, "tasklight: warning: idle notification failed: %v\n", err)
			}
		}
	}
	if opts.match != nil && notifier != nil {
		runOpts.MatchPattern = opts.match
		runOpts.OnMatch = func(event runner.MatchEvent) {
			if err := sendNotification(notificationForMatch(event, focusTarget, opts.sound)); err != nil {
				fmt.Fprintf(stderr, "tasklight: warning: match notification failed: %v\n", err)
			}
		}
	}

	result := runner.Run(context.Background(), runOpts)

	if result.Err != nil && !result.Started {
		fmt.Fprintf(stderr, "tasklight run: failed to start %q: %v\n", command[0], result.Err)
	}

	if notifier != nil {
		if err := sendNotification(notificationForResult(result, focusTarget, opts.sound)); err != nil {
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

	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	focus := fs.Bool("focus", false, "show detailed click-to-focus diagnostics")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(stderr, "tasklight doctor: %v\n\n", err)
		printDoctorHelp(stderr)
		return 2
	}
	if len(fs.Args()) > 0 {
		fmt.Fprintf(stderr, "tasklight doctor: unexpected argument: %s\n\n", strings.Join(fs.Args(), " "))
		printDoctorHelp(stderr)
		return 2
	}
	if *focus {
		return runFocusDoctor(stdout)
	}
	return runDoctor(stdout)
}

func executeNotify(args []string, stdout io.Writer, stderr io.Writer, notifier notify.Notifier) int {
	if hasHelpFlag(args) {
		printNotifyHelp(stdout)
		return 0
	}
	configuration, err := loadConfig()
	if err != nil {
		fmt.Fprintf(stderr, "tasklight notify: configuration error: %v\n", err)
		return 2
	}
	opts, help, err := parseNotifyArgs(args, configuration.Notify)
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

func notificationForMatch(event runner.MatchEvent, focusTarget session.FocusTarget, sound bool) notify.Notification {
	return notify.Notification{
		Title:        "Tasklight",
		Subtitle:     fmt.Sprintf("👀 %s needs attention", event.Name),
		Message:      fmt.Sprintf("Matched %s: %s", event.Stream, conciseOutputLine(event.Line)),
		Sound:        sound,
		ActivateApp:  focusTarget.ActivateApp,
		ClickCommand: focusTarget.ClickCommand(),
	}
}

func conciseOutputLine(line string) string {
	sanitized := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, line)
	sanitized = strings.Join(strings.Fields(sanitized), " ")
	if sanitized == "" {
		return "(empty line)"
	}

	const maxRunes = 160
	runes := []rune(sanitized)
	if len(runes) > maxRunes {
		return string(runes[:maxRunes]) + "…"
	}
	return sanitized
}

func notificationForIdle(event runner.IdleEvent, focusTarget session.FocusTarget, sound bool) notify.Notification {
	return notify.Notification{
		Title:        "Tasklight",
		Subtitle:     fmt.Sprintf("⚠️ %s is still running but idle", event.Name),
		Message:      fmt.Sprintf("No output for %s", formatDuration(event.IdleFor)),
		Sound:        sound,
		ActivateApp:  focusTarget.ActivateApp,
		ClickCommand: focusTarget.ClickCommand(),
	}
}

func notificationForResult(result runner.RunResult, focusTarget session.FocusTarget, sound bool) notify.Notification {
	duration := formatDuration(result.EndedAt.Sub(result.StartedAt))
	notification := notify.Notification{
		Title:        "Tasklight",
		Sound:        sound,
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

func parseRunArgs(args []string, defaults taskconfig.Run) (runOptions, []string, bool, error) {
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

	opts := runOptions{
		activateApp: defaults.ActivateApp,
		idle:        defaults.Idle,
		sound:       defaults.Sound,
	}
	matchPattern := defaults.Match
	fs.StringVar(&opts.name, "name", "", "human-readable task name")
	fs.StringVar(&opts.cwd, "cwd", "", "working directory for the command")
	fs.StringVar(&opts.activateApp, "activate-app", opts.activateApp, "app name or bundle ID to activate when clicking the notification")
	fs.DurationVar(&opts.idle, "idle", opts.idle, "notify after this duration without stdout/stderr output")
	fs.StringVar(&matchPattern, "match", matchPattern, "notify when an output line matches this Go regexp")
	fs.BoolVar(&opts.sound, "sound", opts.sound, "play the default notification sound")

	if err := fs.Parse(flagArgs); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return runOptions{}, nil, true, nil
		}
		return runOptions{}, nil, false, err
	}
	if len(fs.Args()) > 0 {
		return runOptions{}, nil, false, fmt.Errorf("unexpected argument before --: %s", strings.Join(fs.Args(), " "))
	}
	idleSet := false
	matchSet := matchPattern != ""
	fs.Visit(func(visited *flag.Flag) {
		switch visited.Name {
		case "idle":
			idleSet = true
		case "match":
			matchSet = true
		}
	})
	if idleSet && opts.idle <= 0 {
		return runOptions{}, nil, false, errors.New("--idle must be greater than zero")
	}
	if matchSet {
		compiled, err := regexp.Compile(matchPattern)
		if err != nil {
			return runOptions{}, nil, false, fmt.Errorf("invalid --match regexp: %w", err)
		}
		opts.match = compiled
	}

	return opts, command, false, nil
}

func parseNotifyArgs(args []string, defaults taskconfig.Notify) (notifyOptions, bool, error) {
	if hasHelpFlag(args) {
		return notifyOptions{}, true, nil
	}

	fs := flag.NewFlagSet("notify", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	opts := notifyOptions{
		title:       "Tasklight",
		activateApp: defaults.ActivateApp,
		sound:       defaults.Sound,
	}
	fs.StringVar(&opts.title, "title", opts.title, "notification title")
	fs.StringVar(&opts.subtitle, "subtitle", "", "notification subtitle")
	fs.StringVar(&opts.message, "message", "", "notification body/message")
	fs.StringVar(&opts.activateApp, "activate-app", opts.activateApp, "app name or bundle ID to activate when clicking the notification")
	fs.StringVar(&opts.iconPath, "icon", "", "path to a notification icon image")
	fs.BoolVar(&opts.sound, "sound", opts.sound, "play the platform's default notification sound when supported")

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
  tasklight doctor [--focus]

Options:
  --focus    Show terminal/tmux targets, generated commands, provider, and log paths
  -h, --help Show this help

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
  --idle duration         Notify after this duration without stdout/stderr output
  --match regexp          Notify when an output line matches this Go regexp
  --sound                 Play the default notification sound
  -h, --help              Show this help

Examples:
  tasklight run -- pnpm test
  tasklight run --idle 5m -- pi "continue implementation"
  tasklight run --match 'approve|waiting|failed' -- your-agent
  tasklight run -- sh -c 'exit 42'
  tasklight run --cwd frontend -- pnpm build
`)
}
