//go:build darwin
// +build darwin

package session

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

func detectFrontmostBundleID() string {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	output, err := exec.CommandContext(
		ctx,
		"osascript",
		"-e",
		`tell application "System Events" to get bundle identifier of first application process whose frontmost is true`,
	).Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}
