package uv

import (
	"context"
	"fmt"
	"testing"

	"github.com/cottrellashley/orbit/internal/domain"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func lookFound(path string) func(string) (string, error) {
	return func(_ string) (string, error) { return path, nil }
}

func lookNotFound() func(string) (string, error) {
	return func(name string) (string, error) {
		return "", fmt.Errorf("%s: not found", name)
	}
}

func cmdReturns(out string, err error) func(context.Context, string, ...string) ([]byte, error) {
	return func(_ context.Context, _ string, _ ...string) ([]byte, error) {
		return []byte(out), err
	}
}

// ---------------------------------------------------------------------------
// Check tests
// ---------------------------------------------------------------------------

func TestInstaller_Check_Installed(t *testing.T) {
	inst := NewInstaller(
		WithLookPath(lookFound("/usr/local/bin/uv")),
		WithRunCmd(cmdReturns("uv 0.5.1\n", nil)),
	)

	info, err := inst.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Status != domain.InstallStatusInstalled {
		t.Fatalf("expected installed, got %s", info.Status)
	}
	if info.Version != "uv 0.5.1" {
		t.Fatalf("expected version 'uv 0.5.1', got %q", info.Version)
	}
	if info.Name != "uv" {
		t.Fatalf("expected name 'uv', got %q", info.Name)
	}
}

func TestInstaller_Check_NotInstalled(t *testing.T) {
	inst := NewInstaller(
		WithLookPath(lookNotFound()),
	)

	info, err := inst.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Status != domain.InstallStatusNotInstalled {
		t.Fatalf("expected not_installed, got %s", info.Status)
	}
	if info.Version != "" {
		t.Fatalf("expected empty version, got %q", info.Version)
	}
}

func TestInstaller_Check_VersionError(t *testing.T) {
	inst := NewInstaller(
		WithLookPath(lookFound("/usr/local/bin/uv")),
		WithRunCmd(cmdReturns("", fmt.Errorf("version error"))),
	)

	info, err := inst.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Status != domain.InstallStatusInstalled {
		t.Fatalf("expected installed, got %s", info.Status)
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
		WithLookPath(lookFound("/usr/local/bin/uv")),
		WithRunCmd(func(_ context.Context, name string, args ...string) ([]byte, error) {
			callCount++
			if callCount == 1 {
				return []byte("installed ok"), nil
			}
			return []byte("uv 0.5.1\n"), nil
		}),
	)

	result, err := inst.Install(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	if result.Version != "uv 0.5.1" {
		t.Fatalf("expected version 'uv 0.5.1', got %q", result.Version)
	}
}

func TestInstaller_Install_Failure(t *testing.T) {
	inst := NewInstaller(
		WithLookPath(lookNotFound()),
		WithRunCmd(cmdReturns("curl failed", fmt.Errorf("exit 1"))),
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
		WithLookPath(lookNotFound()),
		WithRunCmd(func(_ context.Context, name string, args ...string) ([]byte, error) {
			callCount++
			if callCount == 1 {
				return []byte("done"), nil
			}
			return nil, fmt.Errorf("not found")
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
	if inst.Name() != "uv" {
		t.Fatalf("expected name 'uv', got %q", inst.Name())
	}
	if inst.Description() == "" {
		t.Fatal("expected non-empty description")
	}
}
