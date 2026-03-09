//go:build !windows

package opencode

import (
	"os"
	"syscall"
)

// processAlive checks whether a process with the given PID exists.
// On Unix, kill(pid, 0) succeeds if the process exists (even if we
// can't signal it).
func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}
