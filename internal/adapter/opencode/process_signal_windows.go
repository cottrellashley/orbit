//go:build windows

package opencode

import "os"

func interrupt(p *os.Process) error {
	return p.Kill()
}
