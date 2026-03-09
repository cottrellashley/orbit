//go:build !windows

package opencode

import (
	"os"
	"syscall"
)

func interrupt(p *os.Process) error {
	return p.Signal(syscall.SIGTERM)
}
