package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

const localConfigName = ".tasklight.toml"

type Config struct {
	Run    Run
	Notify Notify
}

type Run struct {
	ActivateApp string
	Sound       bool
	Idle        time.Duration
	Match       string
}

type Notify struct {
	ActivateApp string
	Sound       bool
}

type rawConfig struct {
	Run    rawRun    `toml:"run"`
	Notify rawNotify `toml:"notify"`
}

type rawRun struct {
	ActivateApp *string `toml:"activate_app"`
	Sound       *bool   `toml:"sound"`
	Idle        *string `toml:"idle"`
	Match       *string `toml:"match"`
}

type rawNotify struct {
	ActivateApp *string `toml:"activate_app"`
	Sound       *bool   `toml:"sound"`
}

// Load reads optional global and project-local Tasklight configuration.
func Load() (Config, error) {
	globalPath := globalConfigPath()
	cwd, err := os.Getwd()
	if err != nil {
		return Config{}, fmt.Errorf("determine working directory: %w", err)
	}
	return loadPaths(globalPath, filepath.Join(cwd, localConfigName))
}

func globalConfigPath() string {
	if xdgConfigHome := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "tasklight", "config.toml")
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, ".config", "tasklight", "config.toml")
}

func loadPaths(paths ...string) (Config, error) {
	var config Config
	for _, path := range paths {
		if path == "" {
			continue
		}
		raw, found, err := decodeFile(path)
		if err != nil {
			return Config{}, err
		}
		if !found {
			continue
		}
		if err := apply(path, raw, &config); err != nil {
			return Config{}, err
		}
	}
	return config, nil
}

func decodeFile(path string) (rawConfig, bool, error) {
	var raw rawConfig
	metadata, err := toml.DecodeFile(path, &raw)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return rawConfig{}, false, nil
		}
		return rawConfig{}, false, fmt.Errorf("read %s: %w", path, err)
	}
	if undecoded := metadata.Undecoded(); len(undecoded) > 0 {
		keys := make([]string, len(undecoded))
		for i, key := range undecoded {
			keys[i] = key.String()
		}
		return rawConfig{}, false, fmt.Errorf("read %s: unknown configuration key(s): %s", path, strings.Join(keys, ", "))
	}
	return raw, true, nil
}

func apply(path string, raw rawConfig, config *Config) error {
	if raw.Run.ActivateApp != nil {
		config.Run.ActivateApp = *raw.Run.ActivateApp
	}
	if raw.Run.Sound != nil {
		config.Run.Sound = *raw.Run.Sound
	}
	if raw.Run.Idle != nil {
		idle, err := parseIdle(*raw.Run.Idle)
		if err != nil {
			return fmt.Errorf("read %s: run.idle: %w", path, err)
		}
		config.Run.Idle = idle
	}
	if raw.Run.Match != nil {
		if *raw.Run.Match != "" {
			if _, err := regexp.Compile(*raw.Run.Match); err != nil {
				return fmt.Errorf("read %s: run.match: invalid Go regexp: %w", path, err)
			}
		}
		config.Run.Match = *raw.Run.Match
	}
	if raw.Notify.ActivateApp != nil {
		config.Notify.ActivateApp = *raw.Notify.ActivateApp
	}
	if raw.Notify.Sound != nil {
		config.Notify.Sound = *raw.Notify.Sound
	}
	return nil
}

func parseIdle(value string) (time.Duration, error) {
	if value == "" {
		return 0, nil
	}
	idle, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid Go duration %q: %w", value, err)
	}
	if idle <= 0 {
		return 0, errors.New("must be greater than zero or an empty string to disable")
	}
	return idle, nil
}
