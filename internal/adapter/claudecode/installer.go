// Package claudecode implements port.ToolInstaller for Claude Code
// (https://docs.anthropic.com/en/docs/agents-and-tools/claude-code).
package claudecode

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/cottrellashley/orbit/internal/domain"
)

// Installer implements port.ToolInstaller for Claude Code.
type Installer struct {
	lookPath func(string) (string, error)
	runCmd   func(ctx context.Context, name string, args ...string) ([]byte, error)
}

// Option configures an Installer.
type Option func(*Installer)

// WithLookPath overrides the path lookup function (for tests).
func WithLookPath(fn func(string) (string, error)) Option {
	return func(i *Installer) { i.lookPath = fn }
}

// WithRunCmd overrides the command runner (for tests).
func WithRunCmd(fn func(ctx context.Context, name string, args ...string) ([]byte, error)) Option {
	return func(i *Installer) { i.runCmd = fn }
}

// NewInstaller creates a ToolInstaller for Claude Code.
func NewInstaller(opts ...Option) *Installer {
	i := &Installer{
		lookPath: exec.LookPath,
	}
	for _, o := range opts {
		o(i)
	}
	if i.runCmd == nil {
		i.runCmd = defaultRunCmd
	}
	return i
}

func (i *Installer) Name() string        { return "claude-code" }
func (i *Installer) Description() string { return "Anthropic Claude Code CLI" }

// Check returns the current install status and version.
func (i *Installer) Check(ctx context.Context) (domain.ToolInfo, error) {
	info := domain.ToolInfo{
		Name:        i.Name(),
		Description: i.Description(),
		Status:      domain.InstallStatusNotInstalled,
	}

	_, err := i.lookPath("claude")
	if err != nil {
		return info, nil
	}

	info.Status = domain.InstallStatusInstalled
	out, err := i.runCmd(ctx, "claude", "--version")
	if err == nil {
		info.Version = strings.TrimSpace(string(out))
	}
	return info, nil
}

// Install attempts to install Claude Code via npm.
func (i *Installer) Install(ctx context.Context) (domain.InstallResult, error) {
	result := domain.InstallResult{Name: i.Name()}

	// Check that npm is available.
	if _, err := i.lookPath("npm"); err != nil {
		result.Error = "npm is required to install Claude Code but was not found on PATH"
		return result, nil
	}

	out, err := i.runCmd(ctx, "npm", "install", "-g", "@anthropic-ai/claude-code")
	if err != nil {
		result.Error = fmt.Sprintf("install failed: %s — %s", err, strings.TrimSpace(string(out)))
		return result, nil
	}

	info, _ := i.Check(ctx)
	if info.Status == domain.InstallStatusInstalled {
		result.Success = true
		result.Version = info.Version
	} else {
		result.Error = "install completed but claude not found on PATH"
	}
	return result, nil
}

func defaultRunCmd(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}
