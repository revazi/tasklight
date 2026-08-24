package cli

import (
	"reflect"
	"strings"
	"testing"

	taskconfig "github.com/revazi/tasklight/internal/config"
)

const (
	maxFuzzArgs      = 64
	maxFuzzArgBytes  = 16 << 10
	fuzzArgSeparator = "\x00"
)

func FuzzParseRunArgs(f *testing.F) {
	f.Add("--\x00echo\x00hello")
	f.Add("--name\x00Agent task\x00--idle\x001s\x00--match\x00approve|waiting\x00--sound\x00--\x00/bin/sh\x00-c\x00printf done")
	f.Add("--help")
	f.Add("--idle\x000s\x00--\x00echo")
	f.Add("--match\x00[\x00--\x00echo")
	f.Add("unexpected\x00--\x00echo")
	f.Add("")

	f.Fuzz(func(t *testing.T, encoded string) {
		args, ok := decodeFuzzArgs(encoded)
		if !ok {
			t.Skip()
		}

		opts, command, help, err := parseRunArgs(args, taskconfig.Run{})
		optsAgain, commandAgain, helpAgain, errAgain := parseRunArgs(args, taskconfig.Run{})
		if !reflect.DeepEqual(runSnapshot(opts), runSnapshot(optsAgain)) ||
			!reflect.DeepEqual(command, commandAgain) || help != helpAgain || errorString(err) != errorString(errAgain) {
			t.Fatalf("parseRunArgs(%q) is not deterministic", args)
		}
		if err != nil || help {
			return
		}

		separator := indexOfSeparator(args)
		if separator < 0 || separator == len(args)-1 {
			t.Fatalf("parseRunArgs(%q) succeeded without a command separator and command", args)
		}
		if !reflect.DeepEqual(command, args[separator+1:]) {
			t.Fatalf("command = %q, want %q", command, args[separator+1:])
		}
		if opts.idle < 0 {
			t.Fatalf("idle duration = %s, want non-negative", opts.idle)
		}
	})
}

func FuzzParseNotifyArgs(f *testing.F) {
	f.Add("--message\x00done")
	f.Add("--title\x00Task ' \" light\x00--subtitle\x00✅ finished\nnext line\x00--message\x00exit: 0")
	f.Add("--message\x00$(touch nope); `id` && false")
	f.Add("--sound=false\x00--message\x00quiet")
	f.Add("--help")
	f.Add("--title\x00only a title")
	f.Add("")

	f.Fuzz(func(t *testing.T, encoded string) {
		args, ok := decodeFuzzArgs(encoded)
		if !ok {
			t.Skip()
		}

		opts, help, err := parseNotifyArgs(args, taskconfig.Notify{})
		optsAgain, helpAgain, errAgain := parseNotifyArgs(args, taskconfig.Notify{})
		if opts != optsAgain || help != helpAgain || errorString(err) != errorString(errAgain) {
			t.Fatalf("parseNotifyArgs(%q) is not deterministic", args)
		}
		if err != nil || help {
			return
		}
		if opts.subtitle == "" && opts.message == "" {
			t.Fatalf("parseNotifyArgs(%q) succeeded without notification text", args)
		}
	})
}

type fuzzRunOptions struct {
	name        string
	cwd         string
	activateApp string
	idle        int64
	match       string
	sound       bool
}

func runSnapshot(opts runOptions) fuzzRunOptions {
	match := ""
	if opts.match != nil {
		match = opts.match.String()
	}
	return fuzzRunOptions{
		name:        opts.name,
		cwd:         opts.cwd,
		activateApp: opts.activateApp,
		idle:        int64(opts.idle),
		match:       match,
		sound:       opts.sound,
	}
}

func decodeFuzzArgs(encoded string) ([]string, bool) {
	if len(encoded) > maxFuzzArgBytes {
		return nil, false
	}
	if encoded == "" {
		return nil, true
	}
	args := strings.Split(encoded, fuzzArgSeparator)
	if len(args) > maxFuzzArgs {
		return nil, false
	}
	return args, true
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
