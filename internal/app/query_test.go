package app

import (
	"context"
	"testing"

	"github.com/cottrellashley/orbit/internal/domain"
)

func TestStateQuerier_Projects(t *testing.T) {
	repo := newMockProjectRepo(
		&domain.Project{Name: "a", Path: "/a"},
		&domain.Project{Name: "b", Path: "/b"},
	)
	projSvc := NewProjectService(repo, nil)
	q := NewStateQuerier(projSvc, nil, nil, nil)

	projects := q.Projects()
	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}
}

func TestStateQuerier_Projects_Nil(t *testing.T) {
	q := NewStateQuerier(nil, nil, nil, nil)
	projects := q.Projects()
	if projects != nil {
		t.Fatalf("expected nil, got %v", projects)
	}
}

func TestStateQuerier_Project(t *testing.T) {
	repo := newMockProjectRepo(
		&domain.Project{Name: "alpha", Path: "/a"},
	)
	projSvc := NewProjectService(repo, nil)
	q := NewStateQuerier(projSvc, nil, nil, nil)

	p := q.Project("alpha")
	if p == nil || p.Name != "alpha" {
		t.Fatalf("expected project 'alpha', got %v", p)
	}

	p = q.Project("missing")
	if p != nil {
		t.Fatalf("expected nil for missing project, got %v", p)
	}
}

func TestStateQuerier_Sessions(t *testing.T) {
	envRepo := newMockEnvRepo()
	provider := newMockSessionProvider()
	srv := domain.Server{PID: 1, Port: 9000, Hostname: "127.0.0.1", Directory: "/test", Healthy: true}
	provider.servers = []domain.Server{srv}
	provider.sessions[9000] = []domain.Session{
		{ID: "s1", ServerDir: "/test", ServerPort: 9000},
	}

	sessSvc := NewSessionService(envRepo, provider)
	q := NewStateQuerier(nil, sessSvc, nil, nil)

	sessions := q.Sessions(context.Background())
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
}

func TestStateQuerier_Servers(t *testing.T) {
	envRepo := newMockEnvRepo()
	provider := newMockSessionProvider()
	provider.servers = []domain.Server{
		{PID: 1, Port: 9000, Hostname: "127.0.0.1", Directory: "/test", Healthy: true},
	}

	sessSvc := NewSessionService(envRepo, provider)
	q := NewStateQuerier(nil, sessSvc, nil, nil)

	servers := q.Servers(context.Background())
	if len(servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(servers))
	}
}

func TestStateQuerier_DoctorReport(t *testing.T) {
	cfgDir := t.TempDir()
	envRepo := newMockEnvRepo()
	provider := newMockSessionProvider()

	docSvc := NewDoctorService(cfgDir, nil, envRepo, provider)
	q := NewStateQuerier(nil, nil, docSvc, nil)

	report := q.DoctorReport(context.Background())
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if len(report.Results) == 0 {
		t.Fatal("expected at least some check results")
	}
}

func TestStateQuerier_DoctorReport_Nil(t *testing.T) {
	q := NewStateQuerier(nil, nil, nil, nil)
	report := q.DoctorReport(context.Background())
	if report != nil {
		t.Fatal("expected nil report when doctor not configured")
	}
}

func TestStateQuerier_ProjectPlan(t *testing.T) {
	envRepo := newMockEnvRepo()
	provider := newMockSessionProvider()
	projRepo := newMockProjectRepo(
		&domain.Project{Name: "api", Path: "/projects/api"},
	)

	sessSvc := NewSessionService(envRepo, provider)
	sessSvc.SetProjects(projRepo)

	openSvc := NewOpenService(envRepo, sessSvc)
	openSvc.SetProjects(projRepo, sessSvc)

	projSvc := NewProjectService(projRepo, nil)
	q := NewStateQuerier(projSvc, sessSvc, nil, openSvc)

	plan := q.ProjectPlan(context.Background(), "api")
	if plan == nil {
		t.Fatal("expected non-nil plan")
	}
	if plan.Project.Name != "api" {
		t.Fatalf("expected project 'api', got %q", plan.Project.Name)
	}
}

func TestStateQuerier_ProjectPlan_NilOpen(t *testing.T) {
	q := NewStateQuerier(nil, nil, nil, nil)
	plan := q.ProjectPlan(context.Background(), "any")
	if plan != nil {
		t.Fatal("expected nil plan when open service not configured")
	}
}
