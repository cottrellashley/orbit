package app

import (
	"context"
	"testing"

	"github.com/cottrellashley/orbit/internal/domain"
)

func TestSessionService_ListForProject(t *testing.T) {
	envRepo := newMockEnvRepo()
	provider := newMockSessionProvider()

	srv := domain.Server{PID: 1, Port: 9000, Hostname: "127.0.0.1", Directory: "/projects/web", Healthy: true}
	provider.servers = []domain.Server{srv}
	provider.sessions[9000] = []domain.Session{
		{ID: "s1", Title: "coding", ServerDir: "/projects/web", ServerPort: 9000},
	}

	projRepo := newMockProjectRepo(
		&domain.Project{Name: "web", Path: "/projects/web"},
	)

	svc := NewSessionService(envRepo, provider)
	svc.SetProjects(projRepo)

	sessions, err := svc.ListForProject(context.Background(), "web")
	if err != nil {
		t.Fatalf("ListForProject() error: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].EnvironmentName != "web" {
		t.Fatalf("expected env name 'web', got %q", sessions[0].EnvironmentName)
	}
}

func TestSessionService_ListForProject_NoRepo(t *testing.T) {
	envRepo := newMockEnvRepo()
	provider := newMockSessionProvider()
	svc := NewSessionService(envRepo, provider)

	_, err := svc.ListForProject(context.Background(), "any")
	if err == nil {
		t.Fatal("expected error when project repo not configured")
	}
}

func TestSessionService_ListForProject_NotFound(t *testing.T) {
	envRepo := newMockEnvRepo()
	provider := newMockSessionProvider()
	projRepo := newMockProjectRepo()

	svc := NewSessionService(envRepo, provider)
	svc.SetProjects(projRepo)

	_, err := svc.ListForProject(context.Background(), "ghost")
	if err == nil {
		t.Fatal("expected error for missing project")
	}
}

func TestSessionService_CreateSessionForProject(t *testing.T) {
	envRepo := newMockEnvRepo()
	provider := newMockSessionProvider()

	srv := domain.Server{PID: 1, Port: 9000, Hostname: "127.0.0.1", Directory: "/projects/api", Healthy: true}
	provider.servers = []domain.Server{srv}

	projRepo := newMockProjectRepo(
		&domain.Project{Name: "api", Path: "/projects/api"},
	)

	svc := NewSessionService(envRepo, provider)
	svc.SetProjects(projRepo)

	session, err := svc.CreateSessionForProject(context.Background(), "api", "new session")
	if err != nil {
		t.Fatalf("CreateSessionForProject() error: %v", err)
	}
	if session.EnvironmentName != "api" {
		t.Fatalf("expected env name 'api', got %q", session.EnvironmentName)
	}
	if session.Title != "new session" {
		t.Fatalf("expected title 'new session', got %q", session.Title)
	}
}

func TestSessionService_CreateSessionForProject_NoServer(t *testing.T) {
	envRepo := newMockEnvRepo()
	provider := newMockSessionProvider()
	// No servers running.
	provider.servers = nil

	projRepo := newMockProjectRepo(
		&domain.Project{Name: "api", Path: "/projects/api"},
	)

	svc := NewSessionService(envRepo, provider)
	svc.SetProjects(projRepo)

	_, err := svc.CreateSessionForProject(context.Background(), "api", "new")
	if err == nil {
		t.Fatal("expected error when no server is running")
	}
}

func TestSessionService_MatchProject(t *testing.T) {
	projects := []*domain.Project{
		{Name: "root", Path: "/projects"},
		{Name: "web", Path: "/projects/web"},
		{Name: "other", Path: "/other"},
	}

	tests := []struct {
		serverDir string
		wantName  string
		wantPath  string
	}{
		{"/projects/web", "web", "/projects/web"},
		{"/projects/web/src", "web", "/projects/web"},
		{"/projects", "root", "/projects"},
		{"/projects/api", "root", "/projects"},
		{"/other", "other", "/other"},
		{"/unrelated", "", ""},
		{"", "", ""},
	}

	for _, tt := range tests {
		name, path := matchProject(tt.serverDir, projects)
		if name != tt.wantName || path != tt.wantPath {
			t.Errorf("matchProject(%q) = (%q, %q), want (%q, %q)",
				tt.serverDir, name, path, tt.wantName, tt.wantPath)
		}
	}
}

// Test that the original environment-based methods still work.
func TestSessionService_ListAll_EnvironmentMapping(t *testing.T) {
	envRepo := newMockEnvRepo(
		&domain.Environment{Name: "myenv", Path: "/projects/web"},
	)
	provider := newMockSessionProvider()
	srv := domain.Server{PID: 1, Port: 9000, Hostname: "127.0.0.1", Directory: "/projects/web", Healthy: true}
	provider.servers = []domain.Server{srv}
	provider.sessions[9000] = []domain.Session{
		{ID: "s1", ServerDir: "/projects/web", ServerPort: 9000},
	}

	svc := NewSessionService(envRepo, provider)
	sessions, err := svc.ListAll(context.Background())
	if err != nil {
		t.Fatalf("ListAll() error: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].EnvironmentName != "myenv" {
		t.Fatalf("expected env name 'myenv', got %q", sessions[0].EnvironmentName)
	}
}
