package session

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	terminalBundleID = "com.apple.Terminal"
	itermBundleID    = "com.googlecode.iterm2"
)

type DetectOptions struct {
	ActivateApp string
}

type FocusTarget struct {
	ActivateApp string
	Terminal    TerminalTarget
	Tmux        *TmuxTarget
}

type TerminalTarget struct {
	BundleID        string
	Name            string
	ITermSessionID  string
	TerminalSession string
	ClientTTY       string
}

type TmuxTarget struct {
	Socket      string
	ClientName  string
	ClientTTY   string
	Session     string
	WindowIndex string
	WindowID    string
	PaneIndex   string
	PaneID      string
}

func Detect(opts DetectOptions) FocusTarget {
	terminal := DetectTerminal()
	tmux := DetectTmux()
	if tmux != nil && tmux.ClientTTY != "" {
		terminal.ClientTTY = tmux.ClientTTY
	}

	activateApp := strings.TrimSpace(opts.ActivateApp)
	if activateApp != "" {
		activateApp = NormalizeActivateApp(activateApp)
	} else {
		activateApp = detectActivateApp(terminal)
	}

	return FocusTarget{
		ActivateApp: activateApp,
		Terminal:    terminal,
		Tmux:        tmux,
	}
}

func (target FocusTarget) ClickCommand() string {
	commands := make([]string, 0, 2)
	if target.Tmux != nil {
		if command := target.Tmux.ClickCommand(); command != "" {
			commands = append(commands, command)
		}
	}
	if command := target.FocusCommand(); command != "" {
		commands = append(commands, command)
	}
	return strings.Join(commands, " ; ")
}

func (target FocusTarget) FocusCommand() string {
	if target.ActivateApp == "" {
		return ""
	}
	if target.ActivateApp == itermBundleID {
		if command := itermFocusCommand(target.Terminal); command != "" {
			return command
		}
	}
	if target.ActivateApp == terminalBundleID {
		if command := terminalFocusCommand(target.Terminal); command != "" {
			return command
		}
	}
	return activateAppCommand(target.ActivateApp)
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

	if target.ClientName != "" && target.Session != "" {
		args = append(args, "switch-client", "-c", target.ClientName, "-t", target.Session, ";")
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

func DetectTerminal() TerminalTarget {
	bundleID := strings.TrimSpace(os.Getenv("__CFBundleIdentifier"))
	name := bundleIDDisplayName(bundleID)

	if name == "" {
		name = terminalNameFromEnvironment()
	}
	if bundleID == "" {
		bundleID = bundleIDForTerminalName(name)
	}

	return TerminalTarget{
		BundleID:        bundleID,
		Name:            name,
		ITermSessionID:  normalizeITermSessionID(os.Getenv("ITERM_SESSION_ID")),
		TerminalSession: strings.TrimSpace(os.Getenv("TERM_SESSION_ID")),
	}
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
	args = append(args, "display-message", "-p", "-t", paneID, "#{client_name}\t#{client_tty}\t#{session_name}\t#{window_index}\t#{window_id}\t#{pane_index}\t#{pane_id}")

	output, err := exec.CommandContext(ctx, tmuxPath, args...).Output()
	if err != nil {
		return target
	}

	fields := strings.Split(strings.TrimSpace(string(output)), "\t")
	if len(fields) >= 7 {
		target.ClientName = fields[0]
		target.ClientTTY = fields[1]
		target.Session = fields[2]
		target.WindowIndex = fields[3]
		target.WindowID = fields[4]
		target.PaneIndex = fields[5]
		target.PaneID = fields[6]
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
		return terminalBundleID
	case "iterm", "iterm2", "iterm 2":
		return itermBundleID
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

func detectActivateApp(terminal TerminalTarget) string {
	if terminal.BundleID != "" {
		return terminal.BundleID
	}
	if terminal.Name != "" {
		return NormalizeActivateApp(terminal.Name)
	}
	return detectFrontmostBundleID()
}

func terminalNameFromEnvironment() string {
	for _, key := range []string{"LC_TERMINAL", "TERM_PROGRAM"} {
		value := strings.TrimSpace(os.Getenv(key))
		if value == "" || strings.EqualFold(value, "tmux") {
			continue
		}
		return value
	}
	return ""
}

func bundleIDForTerminalName(name string) string {
	if name == "" {
		return ""
	}
	normalized := NormalizeActivateApp(name)
	if strings.Contains(normalized, ".") {
		return normalized
	}
	return ""
}

func bundleIDDisplayName(bundleID string) string {
	switch bundleID {
	case terminalBundleID:
		return "Terminal"
	case itermBundleID:
		return "iTerm2"
	case "com.github.wez.wezterm":
		return "WezTerm"
	case "com.microsoft.VSCode":
		return "Visual Studio Code"
	case "com.microsoft.VSCodeInsiders":
		return "VS Code Insiders"
	case "com.todesktop.230313mzl4w4u92":
		return "Cursor"
	case "dev.warp.Warp-Stable":
		return "Warp"
	case "com.mitchellh.ghostty":
		return "Ghostty"
	case "net.kovidgoyal.kitty":
		return "kitty"
	case "org.alacritty":
		return "Alacritty"
	default:
		return ""
	}
}

func parseTmuxSocket(value string) string {
	if value == "" {
		return ""
	}
	parts := strings.SplitN(value, ",", 2)
	return parts[0]
}

func normalizeITermSessionID(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if idx := strings.LastIndex(trimmed, ":"); idx >= 0 && idx < len(trimmed)-1 {
		return trimmed[idx+1:]
	}
	return trimmed
}

func activateAppCommand(app string) string {
	verb := "application"
	if strings.Contains(app, ".") {
		verb = "application id"
	}
	script := fmt.Sprintf("tell %s %s to activate", verb, appleScriptString(app))
	return shellJoin([]string{"osascript", "-e", script})
}

func itermFocusCommand(terminal TerminalTarget) string {
	if terminal.ITermSessionID == "" && terminal.ClientTTY == "" {
		return shellJoin([]string{"open", "-b", itermBundleID})
	}
	script := fmt.Sprintf(`set targetSessionID to %s
set targetTTY to %s
tell application id %s
	activate
	repeat with aWindow in windows
		repeat with aTab in tabs of aWindow
			repeat with aSession in sessions of aTab
				try
					set sessionID to id of aSession as text
					set sessionTTY to tty of aSession as text
					if (targetSessionID is not "" and sessionID is targetSessionID) or (targetTTY is not "" and sessionTTY is targetTTY) then
						select aWindow
						select aTab
						select aSession
						return
					end if
				end try
			end repeat
		end repeat
	end repeat
end tell`, appleScriptString(terminal.ITermSessionID), appleScriptString(terminal.ClientTTY), appleScriptString(itermBundleID))
	return strings.Join([]string{
		shellJoin([]string{"open", "-b", itermBundleID}),
		shellJoin([]string{"osascript", "-e", script}),
	}, " ; ")
}

func terminalFocusCommand(terminal TerminalTarget) string {
	if terminal.ClientTTY == "" {
		return shellJoin([]string{"open", "-b", terminalBundleID})
	}
	script := fmt.Sprintf(`set targetTTY to %s
tell application id %s
	activate
	repeat with aWindow in windows
		repeat with aTab in tabs of aWindow
			try
				if (tty of aTab as text) is targetTTY then
					set selected of aTab to true
					set index of aWindow to 1
					return
				end if
			end try
		end repeat
	end repeat
end tell`, appleScriptString(terminal.ClientTTY), appleScriptString(terminalBundleID))
	return strings.Join([]string{
		shellJoin([]string{"open", "-b", terminalBundleID}),
		shellJoin([]string{"osascript", "-e", script}),
	}, " ; ")
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
