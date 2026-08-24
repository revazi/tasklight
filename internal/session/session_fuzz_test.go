package session

import (
	"reflect"
	"strings"
	"testing"
)

const maxFuzzInputBytes = 16 << 10

func FuzzShellJoin(f *testing.F) {
	f.Add("plain", "contains spaces", "quote's value", "line one\nline two")
	f.Add("", "'", `\`, `"double"`)
	f.Add("; rm -rf /", "$(touch nope)", "`id`", "value && false")

	f.Fuzz(func(t *testing.T, first, second, third, fourth string) {
		args := []string{first, second, third, fourth}
		if stringsTotalBytes(args) > maxFuzzInputBytes {
			t.Skip()
		}

		command := shellJoin(args)
		decoded, ok := decodeShellJoin(command)
		if !ok {
			t.Fatalf("shellJoin(%q) produced malformed command %q", args, command)
		}
		if !reflect.DeepEqual(decoded, args) {
			t.Fatalf("shellJoin(%q) decoded as %q from %q", args, decoded, command)
		}
	})
}

func FuzzTmuxClickCommand(f *testing.F) {
	f.Add("/tmp/tmux/default", "/dev/ttys001", "project", "1", "@2", "%3")
	f.Add("socket's path", "client; false", "$(touch nope)", "0", "", "%1; false")
	f.Add("", "", "", "", "", "")

	f.Fuzz(func(t *testing.T, socket, client, session, windowIndex, windowID, paneID string) {
		values := []string{socket, client, session, windowIndex, windowID, paneID}
		if stringsTotalBytes(values) > maxFuzzInputBytes {
			t.Skip()
		}

		target := TmuxTarget{
			Socket:      socket,
			ClientName:  client,
			Session:     session,
			WindowIndex: windowIndex,
			WindowID:    windowID,
			PaneID:      paneID,
		}
		command := target.ClickCommand()
		trimmedPaneID := strings.TrimSpace(paneID)
		if trimmedPaneID == "" {
			if command != "" {
				t.Fatalf("ClickCommand() = %q with an empty pane ID", command)
			}
			return
		}

		args, ok := decodeShellJoin(command)
		if !ok {
			t.Fatalf("ClickCommand() produced malformed shell command %q", command)
		}
		if len(args) < 4 || args[len(args)-2] != "-t" || args[len(args)-1] != trimmedPaneID {
			t.Fatalf("ClickCommand() args = %q, want final pane target %q", args, trimmedPaneID)
		}
		if socket != "" && (len(args) < 3 || args[1] != "-S" || args[2] != socket) {
			t.Fatalf("ClickCommand() args = %q, want socket %q", args, socket)
		}
	})
}

func FuzzParseTmuxSocket(f *testing.F) {
	f.Add("/private/tmp/tmux-501/default,16652,7")
	f.Add("")
	f.Add(",pid,session")
	f.Add("socket-without-metadata")

	f.Fuzz(func(t *testing.T, value string) {
		if len(value) > maxFuzzInputBytes {
			t.Skip()
		}
		got := parseTmuxSocket(value)
		want, _, _ := strings.Cut(value, ",")
		if got != want {
			t.Fatalf("parseTmuxSocket(%q) = %q, want %q", value, got, want)
		}
	})
}

func FuzzNormalizeITermSessionID(f *testing.F) {
	f.Add("w0t0p0:59571519-5F27-46C1-95C1-2A6E9A2286F0")
	f.Add("  session-id  ")
	f.Add(":")
	f.Add("prefix:middle:suffix")

	f.Fuzz(func(t *testing.T, value string) {
		if len(value) > maxFuzzInputBytes {
			t.Skip()
		}
		trimmed := strings.TrimSpace(value)
		want := trimmed
		if idx := strings.LastIndex(trimmed, ":"); idx >= 0 && idx < len(trimmed)-1 {
			want = trimmed[idx+1:]
		}
		if got := normalizeITermSessionID(value); got != want {
			t.Fatalf("normalizeITermSessionID(%q) = %q, want %q", value, got, want)
		}
	})
}

func decodeShellJoin(command string) ([]string, bool) {
	if command == "" {
		return []string{}, true
	}

	args := make([]string, 0, 4)
	for offset := 0; offset < len(command); {
		if command[offset] != '\'' {
			return nil, false
		}
		offset++

		var value strings.Builder
		for {
			if offset >= len(command) {
				return nil, false
			}
			if command[offset] != '\'' {
				value.WriteByte(command[offset])
				offset++
				continue
			}
			if strings.HasPrefix(command[offset:], "'\\''") {
				value.WriteByte('\'')
				offset += len("'\\''")
				continue
			}
			offset++
			break
		}
		args = append(args, value.String())

		if offset == len(command) {
			return args, true
		}
		if command[offset] != ' ' {
			return nil, false
		}
		offset++
		if offset == len(command) {
			return nil, false
		}
	}
	return args, true
}

func stringsTotalBytes(values []string) int {
	total := 0
	for _, value := range values {
		total += len(value)
	}
	return total
}
