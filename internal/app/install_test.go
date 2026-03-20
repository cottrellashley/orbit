package app

import (
	"context"
	"fmt"
	"testing"

	"github.com/cottrellashley/orbit/internal/domain"
)

// ---------------------------------------------------------------------------
// mockToolInstaller implements port.ToolInstaller for testing.
// ---------------------------------------------------------------------------

type mockToolInstaller struct {
	name        string
	description string
	info        domain.ToolInfo
	checkErr    error
	result      domain.InstallResult
	installErr  error
}

func (m *mockToolInstaller) Name() string        { return m.name }
func (m *mockToolInstaller) Description() string { return m.description }

func (m *mockToolInstaller) Check(_ context.Context) (domain.ToolInfo, error) {
	return m.info, m.checkErr
}

func (m *mockToolInstaller) Install(_ context.Context) (domain.InstallResult, error) {
	return m.result, m.installErr
}

// ---------------------------------------------------------------------------
// ListAll tests
// ---------------------------------------------------------------------------

func TestInstallService_ListAll_AllInstalled(t *testing.T) {
	svc := NewInstallService(
		&mockToolInstaller{
			name:        "tool-a",
			description: "Tool A",
			info: domain.ToolInfo{
				Name:    "tool-a",
				Status:  domain.InstallStatusInstalled,
				Version: "1.0.0",
			},
		},
		&mockToolInstaller{
			name:        "tool-b",
			description: "Tool B",
			info: domain.ToolInfo{
				Name:    "tool-b",
				Status:  domain.InstallStatusInstalled,
				Version: "2.0.0",
			},
		},
	)

	tools, err := svc.ListAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}
	if tools[0].Name != "tool-a" {
		t.Fatalf("expected 'tool-a', got %q", tools[0].Name)
	}
	if tools[1].Name != "tool-b" {
		t.Fatalf("expected 'tool-b', got %q", tools[1].Name)
	}
}

func TestInstallService_ListAll_MixedStatus(t *testing.T) {
	svc := NewInstallService(
		&mockToolInstaller{
			name: "installed-tool",
			info: domain.ToolInfo{
				Name:   "installed-tool",
				Status: domain.InstallStatusInstalled,
			},
		},
		&mockToolInstaller{
			name: "missing-tool",
			info: domain.ToolInfo{
				Name:   "missing-tool",
				Status: domain.InstallStatusNotInstalled,
			},
		},
	)

	tools, err := svc.ListAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}
	if tools[0].Status != domain.InstallStatusInstalled {
		t.Fatalf("expected installed, got %s", tools[0].Status)
	}
	if tools[1].Status != domain.InstallStatusNotInstalled {
		t.Fatalf("expected not_installed, got %s", tools[1].Status)
	}
}

func TestInstallService_ListAll_CheckError(t *testing.T) {
	svc := NewInstallService(
		&mockToolInstaller{
			name:        "broken-tool",
			description: "Broken",
			checkErr:    fmt.Errorf("check failed"),
		},
	)

	tools, err := svc.ListAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}
	if tools[0].Status != domain.InstallStatusUnknown {
		t.Fatalf("expected unknown status on check error, got %s", tools[0].Status)
	}
	if tools[0].Name != "broken-tool" {
		t.Fatalf("expected name 'broken-tool', got %q", tools[0].Name)
	}
}

func TestInstallService_ListAll_Empty(t *testing.T) {
	svc := NewInstallService()

	tools, err := svc.ListAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tools) != 0 {
		t.Fatalf("expected 0 tools, got %d", len(tools))
	}
}

// ---------------------------------------------------------------------------
// Install tests
// ---------------------------------------------------------------------------

func TestInstallService_Install_Success(t *testing.T) {
	svc := NewInstallService(
		&mockToolInstaller{
			name: "tool-a",
			result: domain.InstallResult{
				Name:    "tool-a",
				Success: true,
				Version: "1.0.0",
			},
		},
	)

	result, err := svc.Install(context.Background(), "tool-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	if result.Version != "1.0.0" {
		t.Fatalf("expected version '1.0.0', got %q", result.Version)
	}
}

func TestInstallService_Install_Failure(t *testing.T) {
	svc := NewInstallService(
		&mockToolInstaller{
			name: "tool-a",
			result: domain.InstallResult{
				Name:  "tool-a",
				Error: "install failed",
			},
		},
	)

	result, err := svc.Install(context.Background(), "tool-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Fatal("expected failure")
	}
	if result.Error != "install failed" {
		t.Fatalf("expected error 'install failed', got %q", result.Error)
	}
}

func TestInstallService_Install_UnknownTool(t *testing.T) {
	svc := NewInstallService(
		&mockToolInstaller{name: "tool-a"},
	)

	result, err := svc.Install(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
	if result.Success {
		t.Fatal("expected failure for unknown tool")
	}
	if result.Error == "" {
		t.Fatal("expected error message for unknown tool")
	}
}
