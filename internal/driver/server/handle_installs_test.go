package server

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/cottrellashley/orbit/internal/domain"
)

// ---------------------------------------------------------------------------
// handleListInstalls
// ---------------------------------------------------------------------------

func TestHandleListInstalls_OK(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	s.SetInstallService(&mockInstallService{
		tools: []domain.ToolInfo{
			{Name: "opencode", Description: "AI coding agent", Status: domain.InstallStatusInstalled, Version: "1.2.3"},
			{Name: "uv", Description: "Python package manager", Status: domain.InstallStatusNotInstalled},
		},
	})

	w := doRequest(s, "GET", "/api/installs", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var tools []toolInfoJSON
	decodeJSON(t, w, &tools)
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}
	if tools[0].Name != "opencode" {
		t.Fatalf("expected 'opencode', got %q", tools[0].Name)
	}
	if tools[0].Status != "installed" {
		t.Fatalf("expected status 'installed', got %q", tools[0].Status)
	}
	if tools[0].Version != "1.2.3" {
		t.Fatalf("expected version '1.2.3', got %q", tools[0].Version)
	}
	if tools[1].Status != "not_installed" {
		t.Fatalf("expected status 'not_installed', got %q", tools[1].Status)
	}
}

func TestHandleListInstalls_ServiceUnavailable(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	// Do NOT set installSvc — nil-guarded.

	w := doRequest(s, "GET", "/api/installs", "")
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestHandleListInstalls_Error(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	s.SetInstallService(&mockInstallService{
		listErr: fmt.Errorf("internal failure"),
	})

	w := doRequest(s, "GET", "/api/installs", "")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleListInstalls_Empty(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	s.SetInstallService(&mockInstallService{
		tools: []domain.ToolInfo{},
	})

	w := doRequest(s, "GET", "/api/installs", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var tools []toolInfoJSON
	decodeJSON(t, w, &tools)
	if len(tools) != 0 {
		t.Fatalf("expected 0 tools, got %d", len(tools))
	}
}

// ---------------------------------------------------------------------------
// handleInstallTool
// ---------------------------------------------------------------------------

func TestHandleInstallTool_Success(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	s.SetInstallService(&mockInstallService{
		result: domain.InstallResult{
			Name:    "opencode",
			Success: true,
			Version: "1.2.3",
		},
	})

	w := doRequest(s, "POST", "/api/installs/opencode", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result installResultJSON
	decodeJSON(t, w, &result)
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	if result.Name != "opencode" {
		t.Fatalf("expected name 'opencode', got %q", result.Name)
	}
	if result.Version != "1.2.3" {
		t.Fatalf("expected version '1.2.3', got %q", result.Version)
	}
}

func TestHandleInstallTool_InstallFailed(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	s.SetInstallService(&mockInstallService{
		result: domain.InstallResult{
			Name:  "uv",
			Error: "curl failed",
		},
	})

	w := doRequest(s, "POST", "/api/installs/uv", "")
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", w.Code, w.Body.String())
	}

	var result installResultJSON
	decodeJSON(t, w, &result)
	if result.Success {
		t.Fatal("expected failure")
	}
	if result.Error != "curl failed" {
		t.Fatalf("expected error 'curl failed', got %q", result.Error)
	}
}

func TestHandleInstallTool_NotFound(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	s.SetInstallService(&mockInstallService{
		result: domain.InstallResult{
			Name:  "nonexistent",
			Error: "unknown tool: nonexistent",
		},
		installErr: fmt.Errorf("tool %q: not found", "nonexistent"),
	})

	w := doRequest(s, "POST", "/api/installs/nonexistent", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleInstallTool_ServiceUnavailable(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	// installSvc not set.

	w := doRequest(s, "POST", "/api/installs/opencode", "")
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}
