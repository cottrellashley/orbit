// Package uv implements port.ToolInstaller for the uv Python package
// manager (https://astral.sh/uv).
package uv

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/cottrellashley/orbit/internal/domain"
)

// Installer implements port.ToolInstaller for uv.
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

// NewInstaller creates a ToolInstaller for uv.
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

func (i *Installer) Name() string        { return "uv" }
func (i *Installer) Description() string { return "Python package manager (astral.sh/uv)" }

// Check returns the current install status and version.
func (i *Installer) Check(ctx context.Context) (domain.ToolInfo, error) {
	info := domain.ToolInfo{
		Name:        i.Name(),
		Description: i.Description(),
		Status:      domain.InstallStatusNotInstalled,
	}

	_, err := i.lookPath("uv")
	if err != nil {
		return info, nil
	}

	info.Status = domain.InstallStatusInstalled
	out, err := i.runCmd(ctx, "uv", "--version")
	if err == nil {
		info.Version = strings.TrimSpace(string(out))
	}
	return info, nil
}

// Install attempts to install uv via the official install script.
func (i *Installer) Install(ctx context.Context) (domain.InstallResult, error) {
	result := domain.InstallResult{Name: i.Name()}

	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		result.Error = fmt.Sprintf("unsupported platform: %s", runtime.GOOS)
		return result, nil
	}

	out, err := i.runCmd(ctx, "sh", "-c", "curl -LsSf https://astral.sh/uv/install.sh | sh")
	if err != nil {
		result.Error = fmt.Sprintf("install failed: %s — %s", err, strings.TrimSpace(string(out)))
		return result, nil
	}

	info, _ := i.Check(ctx)
	if info.Status == domain.InstallStatusInstalled {
		result.Success = true
		result.Version = info.Version
	} else {
		result.Error = "install script completed but uv not found on PATH"
	}
	return result, nil
}

func defaultRunCmd(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}
