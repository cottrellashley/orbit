package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
)

func TestHandleListEnvironments_Legacy(t *testing.T) {
	envSvc := &mockEnvironmentService{
		envs: []*domain.Environment{
			{
				Name:      "myenv",
				Path:      "/home/user/myenv",
				CreatedAt: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	s := newTestServer(envSvc, &mockProjectService{}, &mockSessionService{})

	w := doRequest(s, "GET", "/api/environments", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var envs []environmentJSON
	decodeJSON(t, w, &envs)
	if len(envs) != 1 {
		t.Fatalf("expected 1 environment, got %d", len(envs))
	}
	if envs[0].Name != "myenv" {
		t.Fatalf("expected name 'myenv', got %q", envs[0].Name)
	}
}

func TestHandleRegisterEnvironment_Legacy(t *testing.T) {
	envSvc := &mockEnvironmentService{}
	s := newTestServer(envSvc, &mockProjectService{}, &mockSessionService{})

	body := `{"name":"legacy","path":"/tmp/legacy","description":"compat test"}`
	w := doRequest(s, "POST", "/api/environments", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var env environmentJSON
	decodeJSON(t, w, &env)
	if env.Name != "legacy" {
		t.Fatalf("expected name 'legacy', got %q", env.Name)
	}
}

func TestHandleDeleteEnvironment_Legacy(t *testing.T) {
	envSvc := &mockEnvironmentService{
		envs: []*domain.Environment{
			{Name: "rmme", Path: "/tmp/rmme"},
		},
	}
	s := newTestServer(envSvc, &mockProjectService{}, &mockSessionService{})

	w := doRequest(s, "DELETE", "/api/environments/rmme", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Not found
	w = doRequest(s, "DELETE", "/api/environments/ghost", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
