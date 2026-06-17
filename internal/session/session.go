package session

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type DetectOptions struct {
	ActivateApp string
}

type FocusTarget struct {
	ActivateApp string
	Tmux        *TmuxTarget
}

type TmuxTarget struct {
	Socket      string
	Session     string
	WindowIndex string
	WindowID    string
	PaneIndex   string
	PaneID      string
}

func Detect(opts DetectOptions) FocusTarget {
	activateApp := strings.TrimSpace(opts.ActivateApp)
	if activateApp != "" {
		activateApp = NormalizeActivateApp(activateApp)
	} else {
		activateApp = detectActivateApp()
	}

	return FocusTarget{
		ActivateApp: activateApp,
		Tmux:        DetectTmux(),
	}
}

func (target FocusTarget) ClickCommand() string {
	commands := make([]string, 0, 2)
	if target.Tmux != nil {
		if command := target.Tmux.ClickCommand(); command != "" {
			commands = append(commands, command)
		}
	}
	if target.ActivateApp != "" {
		commands = append(commands, activateAppCommand(target.ActivateApp))
	}
	return strings.Join(commands, " ; ")
}

func (target TmuxTarget) ClickCommand() string {
	paneID := strings.TrimSpace(target.PaneID)
	if paneID == "" {
		return ""
	}

	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		tmuxPath = "tmux"
	}

	args := []string{tmuxPath}
	if target.Socket != "" {
		args = append(args, "-S", target.Socket)
	}

	windowTarget := target.WindowID
	if windowTarget == "" && target.Session != "" && target.WindowIndex != "" {
		windowTarget = fmt.Sprintf("%s:%s", target.Session, target.WindowIndex)
	}

	if windowTarget != "" {
		args = append(args,
			"select-window", "-t", windowTarget,
			";",
			"select-pane", "-t", paneID,
		)
	} else {
		args = append(args, "select-pane", "-t", paneID)
	}

	return shellJoin(args)
}

func DetectTmux() *TmuxTarget {
	paneID := strings.TrimSpace(os.Getenv("TMUX_PANE"))
	if paneID == "" {
		return nil
	}

	target := &TmuxTarget{
		Socket: parseTmuxSocket(os.Getenv("TMUX")),
		PaneID: paneID,
	}

	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		return target
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	args := make([]string, 0, 8)
	if target.Socket != "" {
		args = append(args, "-S", target.Socket)
	}
	args = append(args, "display-message", "-p", "-t", paneID, "#{session_name}\t#{window_index}\t#{window_id}\t#{pane_index}\t#{pane_id}")

	output, err := exec.CommandContext(ctx, tmuxPath, args...).Output()
	if err != nil {
		return target
	}

	fields := strings.Split(strings.TrimSpace(string(output)), "\t")
	if len(fields) >= 5 {
		target.Session = fields[0]
		target.WindowIndex = fields[1]
		target.WindowID = fields[2]
		target.PaneIndex = fields[3]
		target.PaneID = fields[4]
	}

	return target
}

func NormalizeActivateApp(value string) string {
	trimmed := strings.TrimSpace(value)
	key := strings.ToLower(strings.TrimSuffix(trimmed, ".app"))
	key = strings.ReplaceAll(key, "_", " ")
	key = strings.Join(strings.Fields(key), " ")

	switch key {
	case "terminal", "apple terminal", "apple terminal.app", "apple terminal app":
		return "com.apple.Terminal"
	case "iterm", "iterm2", "iterm 2":
		return "com.googlecode.iterm2"
	case "wezterm":
		return "com.github.wez.wezterm"
	case "visual studio code", "vscode", "vs code", "code":
		return "com.microsoft.VSCode"
	case "visual studio code insiders", "vscode insiders", "code insiders":
		return "com.microsoft.VSCodeInsiders"
	case "cursor":
		return "com.todesktop.230313mzl4w4u92"
	case "warp":
		return "dev.warp.Warp-Stable"
	case "ghostty":
		return "com.mitchellh.ghostty"
	case "kitty":
		return "net.kovidgoyal.kitty"
	case "alacritty":
		return "org.alacritty"
	default:
		return trimmed
	}
}

func detectActivateApp() string {
	if bundleID := detectFrontmostBundleID(); bundleID != "" {
		return bundleID
	}
	return NormalizeActivateApp(os.Getenv("TERM_PROGRAM"))
}

func parseTmuxSocket(value string) string {
	if value == "" {
		return ""
	}
	parts := strings.SplitN(value, ",", 2)
	return parts[0]
}

func activateAppCommand(app string) string {
	verb := "application"
	if strings.Contains(app, ".") {
		verb = "application id"
	}
	script := fmt.Sprintf("tell %s %s to activate", verb, appleScriptString(app))
	return shellJoin([]string{"osascript", "-e", script})
}

func shellJoin(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = shellQuote(arg)
	}
	return strings.Join(quoted, " ")
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func appleScriptString(value string) string {
	escaped := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(value)
	return `"` + escaped + `"`
}
