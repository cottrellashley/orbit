package opencode

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/cottrellashley/orbit/internal/domain"
)

// Installer implements port.ToolInstaller for the opencode binary.
type Installer struct {
	binary string

	// lookPath and runCmd are injectable for testing.
	lookPath func(string) (string, error)
	runCmd   func(ctx context.Context, name string, args ...string) ([]byte, error)
}

// InstallerOption configures an Installer.
type InstallerOption func(*Installer)

// WithInstallerBinary overrides the binary name (default "opencode").
func WithInstallerBinary(binary string) InstallerOption {
	return func(i *Installer) { i.binary = binary }
}

// WithInstallerLookPath overrides the path lookup function (for tests).
func WithInstallerLookPath(fn func(string) (string, error)) InstallerOption {
	return func(i *Installer) { i.lookPath = fn }
}

// WithInstallerRunCmd overrides the command runner (for tests).
func WithInstallerRunCmd(fn func(ctx context.Context, name string, args ...string) ([]byte, error)) InstallerOption {
	return func(i *Installer) { i.runCmd = fn }
}

// NewInstaller creates a ToolInstaller for the opencode binary.
func NewInstaller(opts ...InstallerOption) *Installer {
	i := &Installer{
		binary:   "opencode",
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

func (i *Installer) Name() string        { return "opencode" }
func (i *Installer) Description() string { return "AI coding agent (opencode.ai)" }

// Check returns the current install status and version.
func (i *Installer) Check(ctx context.Context) (domain.ToolInfo, error) {
	info := domain.ToolInfo{
		Name:        i.Name(),
		Description: i.Description(),
		Status:      domain.InstallStatusNotInstalled,
	}

	_, err := i.lookPath(i.binary)
	if err != nil {
		return info, nil
	}

	info.Status = domain.InstallStatusInstalled
	out, err := i.runCmd(ctx, i.binary, "--version")
	if err == nil {
		info.Version = strings.TrimSpace(string(out))
	}
	return info, nil
}

// Install attempts to install opencode via the official install script.
func (i *Installer) Install(ctx context.Context) (domain.InstallResult, error) {
	result := domain.InstallResult{Name: i.Name()}

	// opencode supports macOS and Linux via curl-pipe-sh.
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		result.Error = fmt.Sprintf("unsupported platform: %s", runtime.GOOS)
		return result, nil
	}

	out, err := i.runCmd(ctx, "sh", "-c", "curl -fsSL https://opencode.ai/install | bash")
	if err != nil {
		result.Error = fmt.Sprintf("install failed: %s — %s", err, strings.TrimSpace(string(out)))
		return result, nil
	}

	// Verify installation.
	info, _ := i.Check(ctx)
	if info.Status == domain.InstallStatusInstalled {
		result.Success = true
		result.Version = info.Version
	} else {
		result.Error = "install script completed but binary not found on PATH"
	}
	return result, nil
}

func defaultRunCmd(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}
