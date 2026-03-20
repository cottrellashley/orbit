package github

import (
	"context"
	"fmt"
	"testing"

	"github.com/cottrellashley/orbit/internal/domain"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func ghLookPathFound(path string) func(string) (string, error) {
	return func(_ string) (string, error) { return path, nil }
}

func ghLookPathNotFound() func(string) (string, error) {
	return func(name string) (string, error) {
		return "", fmt.Errorf("%s: not found", name)
	}
}

func ghRunCmdReturns(out string, err error) func(context.Context, string, ...string) ([]byte, error) {
	return func(_ context.Context, _ string, _ ...string) ([]byte, error) {
		return []byte(out), err
	}
}

// ---------------------------------------------------------------------------
// GH CLI Installer — Check
// ---------------------------------------------------------------------------

func TestGHInstaller_Check_Installed(t *testing.T) {
	inst := NewGHInstaller(
		WithGHLookPath(ghLookPathFound("/usr/local/bin/gh")),
		WithGHRunCmd(ghRunCmdReturns("gh version 2.40.0 (2026-01-15)\nhttps://github.com/cli/cli/releases/tag/v2.40.0\n", nil)),
	)

	info, err := inst.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Status != domain.InstallStatusInstalled {
		t.Fatalf("expected installed, got %s", info.Status)
	}
	if info.Version != "gh version 2.40.0 (2026-01-15)" {
		t.Fatalf("unexpected version: %q", info.Version)
	}
	if info.Name != "gh" {
		t.Fatalf("expected name 'gh', got %q", info.Name)
	}
}

func TestGHInstaller_Check_NotInstalled(t *testing.T) {
	inst := NewGHInstaller(
		WithGHLookPath(ghLookPathNotFound()),
	)

	info, err := inst.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Status != domain.InstallStatusNotInstalled {
		t.Fatalf("expected not_installed, got %s", info.Status)
	}
}

func TestGHInstaller_Check_VersionError(t *testing.T) {
	inst := NewGHInstaller(
		WithGHLookPath(ghLookPathFound("/usr/local/bin/gh")),
		WithGHRunCmd(ghRunCmdReturns("", fmt.Errorf("exec error"))),
	)

	info, err := inst.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Status != domain.InstallStatusInstalled {
		t.Fatalf("expected installed, got %s", info.Status)
	}
	if info.Version != "" {
		t.Fatalf("expected empty version, got %q", info.Version)
	}
}

// ---------------------------------------------------------------------------
// GH CLI Installer — Install
// ---------------------------------------------------------------------------

func TestGHInstaller_Install_Success(t *testing.T) {
	callCount := 0
	inst := NewGHInstaller(
		WithGHLookPath(ghLookPathFound("/usr/local/bin/gh")),
		WithGHRunCmd(func(_ context.Context, name string, args ...string) ([]byte, error) {
			callCount++
			if callCount == 1 {
				return []byte("installed"), nil
			}
			return []byte("gh version 2.40.0 (2026-01-15)\n"), nil
		}),
	)

	result, err := inst.Install(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
}

func TestGHInstaller_Install_Failure(t *testing.T) {
	inst := NewGHInstaller(
		WithGHLookPath(ghLookPathNotFound()),
		WithGHRunCmd(ghRunCmdReturns("brew: command not found", fmt.Errorf("exit 127"))),
	)

	result, err := inst.Install(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Fatal("expected failure")
	}
	if result.Error == "" {
		t.Fatal("expected error message")
	}
}

// ---------------------------------------------------------------------------
// GH CLI Installer — Name / Description
// ---------------------------------------------------------------------------

func TestGHInstaller_NameAndDescription(t *testing.T) {
	inst := NewGHInstaller()
	if inst.Name() != "gh" {
		t.Fatalf("expected name 'gh', got %q", inst.Name())
	}
	if inst.Description() == "" {
		t.Fatal("expected non-empty description")
	}
}

// ---------------------------------------------------------------------------
// Copilot Installer — Check
// ---------------------------------------------------------------------------

func TestCopilotInstaller_Check_Installed(t *testing.T) {
	inst := NewCopilotInstaller(
		WithCopilotLookPath(ghLookPathFound("/usr/local/bin/gh")),
		WithCopilotRunCmd(func(_ context.Context, name string, args ...string) ([]byte, error) {
			return []byte("github/gh-copilot\tv0.5.0\n"), nil
		}),
	)

	info, err := inst.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Status != domain.InstallStatusInstalled {
		t.Fatalf("expected installed, got %s", info.Status)
	}
	if info.Name != "gh-copilot" {
		t.Fatalf("expected name 'gh-copilot', got %q", info.Name)
	}
}

func TestCopilotInstaller_Check_NotInstalled_NoGH(t *testing.T) {
	inst := NewCopilotInstaller(
		WithCopilotLookPath(ghLookPathNotFound()),
	)

	info, err := inst.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Status != domain.InstallStatusNotInstalled {
		t.Fatalf("expected not_installed, got %s", info.Status)
	}
}

func TestCopilotInstaller_Check_NotInstalled_ExtMissing(t *testing.T) {
	inst := NewCopilotInstaller(
		WithCopilotLookPath(ghLookPathFound("/usr/local/bin/gh")),
		WithCopilotRunCmd(func(_ context.Context, name string, args ...string) ([]byte, error) {
			// extension list returns no copilot
			return []byte("github/gh-dash\tv1.0.0\n"), nil
		}),
	)

	info, err := inst.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Status != domain.InstallStatusNotInstalled {
		t.Fatalf("expected not_installed, got %s", info.Status)
	}
}

func TestCopilotInstaller_Check_ExtListError(t *testing.T) {
	inst := NewCopilotInstaller(
		WithCopilotLookPath(ghLookPathFound("/usr/local/bin/gh")),
		WithCopilotRunCmd(ghRunCmdReturns("", fmt.Errorf("gh failed"))),
	)

	info, err := inst.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Status != domain.InstallStatusNotInstalled {
		t.Fatalf("expected not_installed, got %s", info.Status)
	}
}

// ---------------------------------------------------------------------------
// Copilot Installer — Install
// ---------------------------------------------------------------------------

func TestCopilotInstaller_Install_Success(t *testing.T) {
	callCount := 0
	inst := NewCopilotInstaller(
		WithCopilotLookPath(ghLookPathFound("/usr/local/bin/gh")),
		WithCopilotRunCmd(func(_ context.Context, name string, args ...string) ([]byte, error) {
			callCount++
			if callCount == 1 {
				return []byte("installed"), nil
			}
			// Check after install — extension list
			return []byte("github/gh-copilot\tv0.5.0\n"), nil
		}),
	)

	result, err := inst.Install(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
}

func TestCopilotInstaller_Install_NoGH(t *testing.T) {
	inst := NewCopilotInstaller(
		WithCopilotLookPath(ghLookPathNotFound()),
	)

	result, err := inst.Install(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Fatal("expected failure when gh is not installed")
	}
	if result.Error == "" {
		t.Fatal("expected error message")
	}
}

func TestCopilotInstaller_Install_Failure(t *testing.T) {
	inst := NewCopilotInstaller(
		WithCopilotLookPath(ghLookPathFound("/usr/local/bin/gh")),
		WithCopilotRunCmd(ghRunCmdReturns("extension install error", fmt.Errorf("exit 1"))),
	)

	result, err := inst.Install(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Fatal("expected failure")
	}
	if result.Error == "" {
		t.Fatal("expected error message")
	}
}

// ---------------------------------------------------------------------------
// Copilot Installer — Name / Description
// ---------------------------------------------------------------------------

func TestCopilotInstaller_NameAndDescription(t *testing.T) {
	inst := NewCopilotInstaller()
	if inst.Name() != "gh-copilot" {
		t.Fatalf("expected name 'gh-copilot', got %q", inst.Name())
	}
	if inst.Description() == "" {
		t.Fatal("expected non-empty description")
	}
}
