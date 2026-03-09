//go:build windows

package opencode

import (
	"fmt"
	"os/exec"
	"strings"
)

// processAlive checks whether a process with the given PID exists.
// On Windows, os.FindProcess always succeeds regardless of whether the
// process exists. We shell out to tasklist as a best-effort check.
func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/NH")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return !strings.Contains(string(out), "No tasks")
}
