package github

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/cottrellashley/orbit/internal/domain"
)

// ---------------------------------------------------------------------------
// GH CLI Installer
// ---------------------------------------------------------------------------

// GHInstaller implements port.ToolInstaller for the GitHub CLI (gh).
type GHInstaller struct {
	lookPath func(string) (string, error)
	runCmd   func(ctx context.Context, name string, args ...string) ([]byte, error)
}

// GHInstallerOption configures a GHInstaller.
type GHInstallerOption func(*GHInstaller)

// WithGHLookPath overrides the path lookup function (for tests).
func WithGHLookPath(fn func(string) (string, error)) GHInstallerOption {
	return func(i *GHInstaller) { i.lookPath = fn }
}

// WithGHRunCmd overrides the command runner (for tests).
func WithGHRunCmd(fn func(ctx context.Context, name string, args ...string) ([]byte, error)) GHInstallerOption {
	return func(i *GHInstaller) { i.runCmd = fn }
}

// NewGHInstaller creates a ToolInstaller for the GitHub CLI.
func NewGHInstaller(opts ...GHInstallerOption) *GHInstaller {
	i := &GHInstaller{
		lookPath: exec.LookPath,
	}
	for _, o := range opts {
		o(i)
	}
	if i.runCmd == nil {
		i.runCmd = defaultGHRunCmd
	}
	return i
}

func (i *GHInstaller) Name() string        { return "gh" }
func (i *GHInstaller) Description() string { return "GitHub CLI" }

// Check returns the current install status and version.
func (i *GHInstaller) Check(ctx context.Context) (domain.ToolInfo, error) {
	info := domain.ToolInfo{
		Name:        i.Name(),
		Description: i.Description(),
		Status:      domain.InstallStatusNotInstalled,
	}

	_, err := i.lookPath("gh")
	if err != nil {
		return info, nil
	}

	info.Status = domain.InstallStatusInstalled
	out, err := i.runCmd(ctx, "gh", "--version")
	if err == nil {
		// gh --version prints "gh version X.Y.Z (date)\n..."
		line := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
		info.Version = line
	}
	return info, nil
}

// Install attempts to install gh via brew on macOS or the official
// install script on Linux.
func (i *GHInstaller) Install(ctx context.Context) (domain.InstallResult, error) {
	result := domain.InstallResult{Name: i.Name()}

	var cmd string
	switch runtime.GOOS {
	case "darwin":
		cmd = "brew install gh"
	case "linux":
		cmd = `(type -p wget >/dev/null || (sudo apt update && sudo apt install wget -y)) && ` +
			`sudo mkdir -p -m 755 /etc/apt/keyrings && ` +
			`wget -qO- https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo tee /etc/apt/keyrings/githubcli-archive-keyring.gpg > /dev/null && ` +
			`sudo chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg && ` +
			`echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null && ` +
			`sudo apt update && sudo apt install gh -y`
	default:
		result.Error = fmt.Sprintf("unsupported platform: %s", runtime.GOOS)
		return result, nil
	}

	out, err := i.runCmd(ctx, "sh", "-c", cmd)
	if err != nil {
		result.Error = fmt.Sprintf("install failed: %s — %s", err, strings.TrimSpace(string(out)))
		return result, nil
	}

	info, _ := i.Check(ctx)
	if info.Status == domain.InstallStatusInstalled {
		result.Success = true
		result.Version = info.Version
	} else {
		result.Error = "install completed but gh not found on PATH"
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// GitHub Copilot CLI Installer
// ---------------------------------------------------------------------------

// CopilotInstaller implements port.ToolInstaller for GitHub Copilot CLI
// (installed as a gh extension).
type CopilotInstaller struct {
	lookPath func(string) (string, error)
	runCmd   func(ctx context.Context, name string, args ...string) ([]byte, error)
}

// CopilotInstallerOption configures a CopilotInstaller.
type CopilotInstallerOption func(*CopilotInstaller)

// WithCopilotLookPath overrides the path lookup function (for tests).
func WithCopilotLookPath(fn func(string) (string, error)) CopilotInstallerOption {
	return func(i *CopilotInstaller) { i.lookPath = fn }
}

// WithCopilotRunCmd overrides the command runner (for tests).
func WithCopilotRunCmd(fn func(ctx context.Context, name string, args ...string) ([]byte, error)) CopilotInstallerOption {
	return func(i *CopilotInstaller) { i.runCmd = fn }
}

// NewCopilotInstaller creates a ToolInstaller for GitHub Copilot CLI.
func NewCopilotInstaller(opts ...CopilotInstallerOption) *CopilotInstaller {
	i := &CopilotInstaller{
		lookPath: exec.LookPath,
	}
	for _, o := range opts {
		o(i)
	}
	if i.runCmd == nil {
		i.runCmd = defaultGHRunCmd
	}
	return i
}

func (i *CopilotInstaller) Name() string        { return "gh-copilot" }
func (i *CopilotInstaller) Description() string { return "GitHub Copilot CLI (gh extension)" }

// Check returns the current install status. Copilot CLI is a gh extension,
// so we check for gh first, then check if the extension is installed.
func (i *CopilotInstaller) Check(ctx context.Context) (domain.ToolInfo, error) {
	info := domain.ToolInfo{
		Name:        i.Name(),
		Description: i.Description(),
		Status:      domain.InstallStatusNotInstalled,
	}

	// gh must be installed first.
	if _, err := i.lookPath("gh"); err != nil {
		return info, nil
	}

	// Check if gh-copilot extension is installed.
	out, err := i.runCmd(ctx, "gh", "extension", "list")
	if err != nil {
		return info, nil
	}

	if strings.Contains(string(out), "gh-copilot") {
		info.Status = domain.InstallStatusInstalled
		info.Version = "installed"
	}
	return info, nil
}

// Install installs gh-copilot as a gh extension.
func (i *CopilotInstaller) Install(ctx context.Context) (domain.InstallResult, error) {
	result := domain.InstallResult{Name: i.Name()}

	// Requires gh to be installed first.
	if _, err := i.lookPath("gh"); err != nil {
		result.Error = "GitHub CLI (gh) must be installed first"
		return result, nil
	}

	out, err := i.runCmd(ctx, "gh", "extension", "install", "github/gh-copilot")
	if err != nil {
		result.Error = fmt.Sprintf("install failed: %s — %s", err, strings.TrimSpace(string(out)))
		return result, nil
	}

	info, _ := i.Check(ctx)
	if info.Status == domain.InstallStatusInstalled {
		result.Success = true
		result.Version = info.Version
	} else {
		result.Error = "install completed but gh-copilot extension not found"
	}
	return result, nil
}

func defaultGHRunCmd(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}
