package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
)

func TestHandleListProjects_Empty(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	w := doRequest(s, "GET", "/api/projects", "")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var projects []projectJSON
	decodeJSON(t, w, &projects)
	if len(projects) != 0 {
		t.Fatalf("expected 0 projects, got %d", len(projects))
	}
}

func TestHandleListProjects_WithData(t *testing.T) {
	projSvc := &mockProjectService{
		projects: []*domain.Project{
			{
				Name:         "alpha",
				Path:         "/home/user/alpha",
				Description:  "first project",
				Topology:     domain.TopologySingleRepo,
				Integrations: []domain.IntegrationTag{domain.TagGit, domain.TagPython},
				Repos: []domain.RepoInfo{
					{Path: "/home/user/alpha", RemoteURL: "git@github.com:x/y.git", CurrentBranch: "main"},
				},
				CreatedAt: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	s := newTestServer(&mockEnvironmentService{}, projSvc, &mockSessionService{})
	w := doRequest(s, "GET", "/api/projects", "")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var projects []projectJSON
	decodeJSON(t, w, &projects)
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	p := projects[0]
	if p.Name != "alpha" {
		t.Fatalf("expected name 'alpha', got %q", p.Name)
	}
	if p.Topology != "single-repo" {
		t.Fatalf("expected topology 'single-repo', got %q", p.Topology)
	}
	if len(p.Integrations) != 2 {
		t.Fatalf("expected 2 integrations, got %d", len(p.Integrations))
	}
	if len(p.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(p.Repos))
	}
	if p.Repos[0].RemoteURL != "git@github.com:x/y.git" {
		t.Fatalf("expected remote URL, got %q", p.Repos[0].RemoteURL)
	}
}

func TestHandleGetProject(t *testing.T) {
	projSvc := &mockProjectService{
		projects: []*domain.Project{
			{
				Name:      "beta",
				Path:      "/home/user/beta",
				Topology:  domain.TopologyMultiRepo,
				CreatedAt: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	s := newTestServer(&mockEnvironmentService{}, projSvc, &mockSessionService{})

	// Found
	w := doRequest(s, "GET", "/api/projects/beta", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var p projectJSON
	decodeJSON(t, w, &p)
	if p.Name != "beta" {
		t.Fatalf("expected name 'beta', got %q", p.Name)
	}

	// Not found
	w = doRequest(s, "GET", "/api/projects/ghost", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandleRegisterProject(t *testing.T) {
	projSvc := &mockProjectService{}
	s := newTestServer(&mockEnvironmentService{}, projSvc, &mockSessionService{})

	body := `{"name":"myproj","path":"/tmp/myproj","description":"test"}`
	w := doRequest(s, "POST", "/api/projects", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var p projectJSON
	decodeJSON(t, w, &p)
	if p.Name != "myproj" {
		t.Fatalf("expected name 'myproj', got %q", p.Name)
	}
	if p.Description != "test" {
		t.Fatalf("expected description 'test', got %q", p.Description)
	}
}

func TestHandleRegisterProject_Validation(t *testing.T) {
	projSvc := &mockProjectService{}
	s := newTestServer(&mockEnvironmentService{}, projSvc, &mockSessionService{})

	// Missing name
	w := doRequest(s, "POST", "/api/projects", `{"path":"/tmp/p"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing name, got %d", w.Code)
	}

	// Missing path
	w = doRequest(s, "POST", "/api/projects", `{"name":"x"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing path, got %d", w.Code)
	}

	// Invalid JSON
	w = doRequest(s, "POST", "/api/projects", `{bad json}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad JSON, got %d", w.Code)
	}
}

func TestHandleDeleteProject(t *testing.T) {
	projSvc := &mockProjectService{
		projects: []*domain.Project{
			{Name: "deleteme", Path: "/tmp/del"},
		},
	}
	s := newTestServer(&mockEnvironmentService{}, projSvc, &mockSessionService{})

	// Delete existing
	w := doRequest(s, "DELETE", "/api/projects/deleteme", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Delete non-existent
	w = doRequest(s, "DELETE", "/api/projects/ghost", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
