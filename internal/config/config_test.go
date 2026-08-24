package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadPathsMergesGlobalAndProjectConfig(t *testing.T) {
	dir := t.TempDir()
	globalPath := filepath.Join(dir, "global.toml")
	localPath := filepath.Join(dir, "local.toml")
	writeConfig(t, globalPath, `
[run]
activate_app = "iTerm2"
sound = true
idle = "5m"
match = "approve|waiting"

[notify]
activate_app = "Terminal"
sound = true
`)
	writeConfig(t, localPath, `
[run]
sound = false
idle = "30s"

[notify]
activate_app = "Cursor"
`)

	got, err := loadPaths(globalPath, localPath)
	if err != nil {
		t.Fatalf("loadPaths() error = %v", err)
	}
	if got.Run.ActivateApp != "iTerm2" || got.Run.Sound || got.Run.Idle != 30*time.Second || got.Run.Match != "approve|waiting" {
		t.Fatalf("Run = %#v, want merged global/project settings", got.Run)
	}
	if got.Notify.ActivateApp != "Cursor" || !got.Notify.Sound {
		t.Fatalf("Notify = %#v, want merged global/project settings", got.Notify)
	}
}

func TestLoadPathsAllowsProjectConfigToDisableDefaults(t *testing.T) {
	dir := t.TempDir()
	globalPath := filepath.Join(dir, "global.toml")
	localPath := filepath.Join(dir, "local.toml")
	writeConfig(t, globalPath, "[run]\nidle = \"5m\"\nmatch = \"waiting\"\n")
	writeConfig(t, localPath, "[run]\nidle = \"\"\nmatch = \"\"\n")

	got, err := loadPaths(globalPath, localPath)
	if err != nil {
		t.Fatalf("loadPaths() error = %v", err)
	}
	if got.Run.Idle != 0 || got.Run.Match != "" {
		t.Fatalf("Run = %#v, want disabled idle and match defaults", got.Run)
	}
}

func TestLoadPathsIgnoresMissingFiles(t *testing.T) {
	got, err := loadPaths(filepath.Join(t.TempDir(), "missing.toml"))
	if err != nil {
		t.Fatalf("loadPaths() error = %v", err)
	}
	if got != (Config{}) {
		t.Fatalf("Config = %#v, want zero defaults", got)
	}
}

func TestLoadPathsRejectsInvalidConfig(t *testing.T) {
	tests := map[string]struct {
		content string
		want    string
	}{
		"syntax": {
			content: "[run\nidle = \"5m\"",
			want:    "read ",
		},
		"unknown key": {
			content: "[run]\nunknown = true\n",
			want:    "unknown configuration key(s): run.unknown",
		},
		"idle duration": {
			content: "[run]\nidle = \"later\"\n",
			want:    "run.idle: invalid Go duration",
		},
		"non-positive idle": {
			content: "[run]\nidle = \"0s\"\n",
			want:    "run.idle: must be greater than zero",
		},
		"match regexp": {
			content: "[run]\nmatch = \"[\"\n",
			want:    "run.match: invalid Go regexp",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.toml")
			writeConfig(t, path, test.content)

			_, err := loadPaths(path)
			if err == nil || !strings.Contains(err.Error(), test.want) || !strings.Contains(err.Error(), path) {
				t.Fatalf("loadPaths() error = %v, want path and %q", err, test.want)
			}
		})
	}
}

func TestGlobalConfigPathUsesXDGConfigHome(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)

	want := filepath.Join(xdg, "tasklight", "config.toml")
	if got := globalConfigPath(); got != want {
		t.Fatalf("globalConfigPath() = %q, want %q", got, want)
	}
}

func writeConfig(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%q): %v", path, err)
	}
}
