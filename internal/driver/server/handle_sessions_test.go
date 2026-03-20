package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
)

func TestHandleListServers(t *testing.T) {
	sessSvc := &mockSessionService{
		servers: []domain.Server{
			{PID: 100, Port: 3001, Hostname: "localhost", Directory: "/tmp/a", Healthy: true},
		},
	}
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, sessSvc)

	w := doRequest(s, "GET", "/api/servers", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var servers []serverJSON
	decodeJSON(t, w, &servers)
	if len(servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(servers))
	}
	if servers[0].Port != 3001 {
		t.Fatalf("expected port 3001, got %d", servers[0].Port)
	}
}

func TestHandleListSessions(t *testing.T) {
	now := time.Date(2026, 3, 9, 12, 0, 0, 0, time.UTC)
	sessSvc := &mockSessionService{
		sessions: []domain.Session{
			{ID: "s1", Title: "test", Status: "idle", CreatedAt: now, UpdatedAt: now},
		},
	}
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, sessSvc)

	w := doRequest(s, "GET", "/api/sessions", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var sessions []sessionJSON
	decodeJSON(t, w, &sessions)
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].ID != "s1" {
		t.Fatalf("expected session ID 's1', got %q", sessions[0].ID)
	}
}

func TestHandleGetSession(t *testing.T) {
	now := time.Date(2026, 3, 9, 12, 0, 0, 0, time.UTC)
	sessSvc := &mockSessionService{
		sessions: []domain.Session{
			{ID: "s1", Title: "test", Status: "idle", CreatedAt: now, UpdatedAt: now},
		},
	}
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, sessSvc)

	// Found
	w := doRequest(s, "GET", "/api/sessions/s1", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Not found
	w = doRequest(s, "GET", "/api/sessions/missing", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandleDoctor(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})

	w := doRequest(s, "GET", "/api/doctor", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var report struct {
		OK      bool `json:"ok"`
		Results []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"results"`
	}
	decodeJSON(t, w, &report)
	if !report.OK {
		t.Fatal("expected report.OK to be true")
	}
	if len(report.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(report.Results))
	}
}

// ---------------------------------------------------------------------------
// Response shape consistency tests
// ---------------------------------------------------------------------------

func TestErrorResponse_Shape(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})

	// Any 404 should produce {"error":"..."}
	w := doRequest(s, "GET", "/api/projects/nope", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	var errResp map[string]string
	decodeJSON(t, w, &errResp)
	if _, ok := errResp["error"]; !ok {
		t.Fatalf("expected error key in response, got %v", errResp)
	}
}

func TestContentType_JSON(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})

	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/projects"},
		{"GET", "/api/environments"},
		{"GET", "/api/servers"},
		{"GET", "/api/sessions"},
		{"GET", "/api/doctor"},
		{"GET", "/api/github/status"},
		{"GET", "/api/github/repos"},
	}

	for _, ep := range endpoints {
		w := doRequest(s, ep.method, ep.path, "")
		ct := w.Header().Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("%s %s: expected Content-Type 'application/json', got %q", ep.method, ep.path, ct)
		}
	}
}
