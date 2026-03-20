package app

import (
	"context"
	"testing"

	"github.com/cottrellashley/orbit/internal/domain"
)

func TestOpenService_Resolve_Legacy(t *testing.T) {
	envRepo := newMockEnvRepo(
		&domain.Environment{Name: "myenv", Path: "/projects/web"},
	)
	provider := newMockSessionProvider()
	sessSvc := NewSessionService(envRepo, provider)

	openSvc := NewOpenService(envRepo, sessSvc)

	plan, err := openSvc.Resolve(context.Background(), "myenv")
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if plan.Environment.Name != "myenv" {
		t.Fatalf("expected env name 'myenv', got %q", plan.Environment.Name)
	}
	if plan.Action != domain.OpenActionCreate {
		t.Fatalf("expected OpenActionCreate (no sessions), got %v", plan.Action)
	}
}

func TestOpenService_Resolve_WithSessions(t *testing.T) {
	envRepo := newMockEnvRepo(
		&domain.Environment{Name: "myenv", Path: "/projects/web"},
	)
	provider := newMockSessionProvider()
	srv := domain.Server{PID: 1, Port: 9000, Hostname: "127.0.0.1", Directory: "/projects/web", Healthy: true}
	provider.servers = []domain.Server{srv}
	provider.sessions[9000] = []domain.Session{
		{ID: "s1", ServerDir: "/projects/web", ServerPort: 9000},
	}

	sessSvc := NewSessionService(envRepo, provider)
	openSvc := NewOpenService(envRepo, sessSvc)

	plan, err := openSvc.Resolve(context.Background(), "myenv")
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if plan.Action != domain.OpenActionResume {
		t.Fatalf("expected OpenActionResume (1 session), got %v", plan.Action)
	}
	if len(plan.Sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(plan.Sessions))
	}
}

func TestOpenService_Resolve_MultipleSessions(t *testing.T) {
	envRepo := newMockEnvRepo(
		&domain.Environment{Name: "myenv", Path: "/projects/web"},
	)
	provider := newMockSessionProvider()
	srv := domain.Server{PID: 1, Port: 9000, Hostname: "127.0.0.1", Directory: "/projects/web", Healthy: true}
	provider.servers = []domain.Server{srv}
	provider.sessions[9000] = []domain.Session{
		{ID: "s1", ServerDir: "/projects/web", ServerPort: 9000},
		{ID: "s2", ServerDir: "/projects/web", ServerPort: 9000},
	}

	sessSvc := NewSessionService(envRepo, provider)
	openSvc := NewOpenService(envRepo, sessSvc)

	plan, err := openSvc.Resolve(context.Background(), "myenv")
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if plan.Action != domain.OpenActionSelect {
		t.Fatalf("expected OpenActionSelect (2 sessions), got %v", plan.Action)
	}
}

func TestOpenService_ResolveProject(t *testing.T) {
	envRepo := newMockEnvRepo()
	provider := newMockSessionProvider()
	srv := domain.Server{PID: 1, Port: 9000, Hostname: "127.0.0.1", Directory: "/projects/api", Healthy: true}
	provider.servers = []domain.Server{srv}
	provider.sessions[9000] = []domain.Session{
		{ID: "s1", ServerDir: "/projects/api", ServerPort: 9000},
	}

	projRepo := newMockProjectRepo(
		&domain.Project{Name: "api", Path: "/projects/api"},
	)

	sessSvc := NewSessionService(envRepo, provider)
	sessSvc.SetProjects(projRepo)

	openSvc := NewOpenService(envRepo, sessSvc)
	openSvc.SetProjects(projRepo, sessSvc)

	plan, err := openSvc.ResolveProject(context.Background(), "api")
	if err != nil {
		t.Fatalf("ResolveProject() error: %v", err)
	}
	if plan.Project.Name != "api" {
		t.Fatalf("expected project name 'api', got %q", plan.Project.Name)
	}
	if plan.Action != domain.OpenActionResume {
		t.Fatalf("expected OpenActionResume, got %v", plan.Action)
	}
	if !plan.ServerOnline {
		t.Fatal("expected ServerOnline to be true")
	}
}

func TestOpenService_ResolveProject_NoSessions(t *testing.T) {
	envRepo := newMockEnvRepo()
	provider := newMockSessionProvider()
	// Server running but no sessions.
	srv := domain.Server{PID: 1, Port: 9000, Hostname: "127.0.0.1", Directory: "/projects/api", Healthy: true}
	provider.servers = []domain.Server{srv}

	projRepo := newMockProjectRepo(
		&domain.Project{Name: "api", Path: "/projects/api"},
	)

	sessSvc := NewSessionService(envRepo, provider)
	sessSvc.SetProjects(projRepo)

	openSvc := NewOpenService(envRepo, sessSvc)
	openSvc.SetProjects(projRepo, sessSvc)

	plan, err := openSvc.ResolveProject(context.Background(), "api")
	if err != nil {
		t.Fatalf("ResolveProject() error: %v", err)
	}
	if plan.Action != domain.OpenActionCreate {
		t.Fatalf("expected OpenActionCreate, got %v", plan.Action)
	}
	if !plan.ServerOnline {
		t.Fatal("expected ServerOnline to be true (server running, just no sessions)")
	}
}

func TestOpenService_ResolveProject_NotConfigured(t *testing.T) {
	envRepo := newMockEnvRepo()
	provider := newMockSessionProvider()
	sessSvc := NewSessionService(envRepo, provider)
	openSvc := NewOpenService(envRepo, sessSvc)

	_, err := openSvc.ResolveProject(context.Background(), "any")
	if err == nil {
		t.Fatal("expected error when project repository not configured")
	}
}

func TestOpenService_ServerForProject(t *testing.T) {
	envRepo := newMockEnvRepo()
	provider := newMockSessionProvider()
	srv := domain.Server{PID: 1, Port: 9000, Hostname: "127.0.0.1", Directory: "/projects/api", Healthy: true}
	provider.servers = []domain.Server{srv}

	projRepo := newMockProjectRepo(
		&domain.Project{Name: "api", Path: "/projects/api"},
	)

	sessSvc := NewSessionService(envRepo, provider)
	openSvc := NewOpenService(envRepo, sessSvc)
	openSvc.SetProjects(projRepo, sessSvc)

	found, err := openSvc.ServerForProject(context.Background(), "api")
	if err != nil {
		t.Fatalf("ServerForProject() error: %v", err)
	}
	if found == nil {
		t.Fatal("expected server to be found")
	}
	if found.Port != 9000 {
		t.Fatalf("expected port 9000, got %d", found.Port)
	}
}

func TestOpenService_ServerForProject_NotRunning(t *testing.T) {
	envRepo := newMockEnvRepo()
	provider := newMockSessionProvider()

	projRepo := newMockProjectRepo(
		&domain.Project{Name: "api", Path: "/projects/api"},
	)

	sessSvc := NewSessionService(envRepo, provider)
	openSvc := NewOpenService(envRepo, sessSvc)
	openSvc.SetProjects(projRepo, sessSvc)

	found, err := openSvc.ServerForProject(context.Background(), "api")
	if err != nil {
		t.Fatalf("ServerForProject() error: %v", err)
	}
	if found != nil {
		t.Fatal("expected nil server when none running")
	}
}

func TestOpenService_ServerForEnvironment_Legacy(t *testing.T) {
	envRepo := newMockEnvRepo(
		&domain.Environment{Name: "myenv", Path: "/projects/web"},
	)
	provider := newMockSessionProvider()
	srv := domain.Server{PID: 1, Port: 9000, Hostname: "127.0.0.1", Directory: "/projects/web", Healthy: true}
	provider.servers = []domain.Server{srv}

	sessSvc := NewSessionService(envRepo, provider)
	openSvc := NewOpenService(envRepo, sessSvc)

	found, err := openSvc.ServerForEnvironment(context.Background(), "myenv")
	if err != nil {
		t.Fatalf("ServerForEnvironment() error: %v", err)
	}
	if found == nil {
		t.Fatal("expected server to be found")
	}
}
