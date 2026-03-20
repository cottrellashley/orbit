package opencode

import (
	"context"
	"fmt"
	"testing"

	"github.com/cottrellashley/orbit/internal/domain"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func lookPathFound(path string) func(string) (string, error) {
	return func(_ string) (string, error) { return path, nil }
}

func lookPathNotFound() func(string) (string, error) {
	return func(name string) (string, error) {
		return "", fmt.Errorf("%s: not found", name)
	}
}

func runCmdReturns(out string, err error) func(context.Context, string, ...string) ([]byte, error) {
	return func(_ context.Context, _ string, _ ...string) ([]byte, error) {
		return []byte(out), err
	}
}

// ---------------------------------------------------------------------------
// Check tests
// ---------------------------------------------------------------------------

func TestInstaller_Check_Installed(t *testing.T) {
	inst := NewInstaller(
		WithInstallerLookPath(lookPathFound("/usr/local/bin/opencode")),
		WithInstallerRunCmd(runCmdReturns("1.2.3\n", nil)),
	)

	info, err := inst.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Status != domain.InstallStatusInstalled {
		t.Fatalf("expected status installed, got %s", info.Status)
	}
	if info.Version != "1.2.3" {
		t.Fatalf("expected version '1.2.3', got %q", info.Version)
	}
	if info.Name != "opencode" {
		t.Fatalf("expected name 'opencode', got %q", info.Name)
	}
}

func TestInstaller_Check_NotInstalled(t *testing.T) {
	inst := NewInstaller(
		WithInstallerLookPath(lookPathNotFound()),
	)

	info, err := inst.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Status != domain.InstallStatusNotInstalled {
		t.Fatalf("expected status not_installed, got %s", info.Status)
	}
	if info.Version != "" {
		t.Fatalf("expected empty version, got %q", info.Version)
	}
}

func TestInstaller_Check_VersionError(t *testing.T) {
	inst := NewInstaller(
		WithInstallerLookPath(lookPathFound("/usr/local/bin/opencode")),
		WithInstallerRunCmd(runCmdReturns("", fmt.Errorf("version error"))),
	)

	info, err := inst.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Status != domain.InstallStatusInstalled {
		t.Fatalf("expected status installed, got %s", info.Status)
	}
	if info.Version != "" {
		t.Fatalf("expected empty version on error, got %q", info.Version)
	}
}

// ---------------------------------------------------------------------------
// Install tests
// ---------------------------------------------------------------------------

func TestInstaller_Install_Success(t *testing.T) {
	callCount := 0
	inst := NewInstaller(
		WithInstallerLookPath(lookPathFound("/usr/local/bin/opencode")),
		WithInstallerRunCmd(func(_ context.Context, name string, args ...string) ([]byte, error) {
			callCount++
			if callCount == 1 {
				// Install command
				return []byte("installed ok"), nil
			}
			// Version check after install
			return []byte("1.2.3\n"), nil
		}),
	)

	result, err := inst.Install(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	if result.Version != "1.2.3" {
		t.Fatalf("expected version '1.2.3', got %q", result.Version)
	}
}

func TestInstaller_Install_Failure(t *testing.T) {
	inst := NewInstaller(
		WithInstallerLookPath(lookPathNotFound()),
		WithInstallerRunCmd(runCmdReturns("connection refused", fmt.Errorf("exit 1"))),
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

func TestInstaller_Install_ScriptOK_BinaryMissing(t *testing.T) {
	callCount := 0
	inst := NewInstaller(
		WithInstallerLookPath(lookPathNotFound()),
		WithInstallerRunCmd(func(_ context.Context, name string, args ...string) ([]byte, error) {
			callCount++
			if callCount == 1 {
				return []byte("done"), nil // install succeeds
			}
			return nil, fmt.Errorf("not found") // but version check fails
		}),
	)

	result, err := inst.Install(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Fatal("expected failure when binary not found after install")
	}
	if result.Error == "" {
		t.Fatal("expected error message")
	}
}

// ---------------------------------------------------------------------------
// Name / Description
// ---------------------------------------------------------------------------

func TestInstaller_NameAndDescription(t *testing.T) {
	inst := NewInstaller()
	if inst.Name() != "opencode" {
		t.Fatalf("expected name 'opencode', got %q", inst.Name())
	}
	if inst.Description() == "" {
		t.Fatal("expected non-empty description")
	}
}

// ---------------------------------------------------------------------------
// Options
// ---------------------------------------------------------------------------

func TestInstaller_WithBinary(t *testing.T) {
	inst := NewInstaller(
		WithInstallerBinary("mybin"),
		WithInstallerLookPath(func(name string) (string, error) {
			if name != "mybin" {
				t.Fatalf("expected lookup of 'mybin', got %q", name)
			}
			return "/usr/bin/mybin", nil
		}),
		WithInstallerRunCmd(runCmdReturns("9.9.9", nil)),
	)

	info, _ := inst.Check(context.Background())
	if info.Status != domain.InstallStatusInstalled {
		t.Fatalf("expected installed with custom binary")
	}
}
