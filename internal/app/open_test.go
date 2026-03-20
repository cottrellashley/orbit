package app

import (
	"context"
	"errors"
	"testing"

	"github.com/cottrellashley/orbit/internal/domain"
)

// ---------------------------------------------------------------------------
// OpenService.Resolve
// ---------------------------------------------------------------------------

type stubSessionQuerier struct {
	sessions []domain.Session
	servers  []domain.Server
	listErr  error
}

func (q *stubSessionQuerier) ListForEnvironment(_ context.Context, _ string) ([]domain.Session, error) {
	return q.sessions, q.listErr
}
func (q *stubSessionQuerier) DiscoverServers(_ context.Context) ([]domain.Server, error) {
	return q.servers, nil
}

func makeEnvRepo(envs ...*domain.Environment) *stubEnvRepo {
	return &stubEnvRepo{envs: envs}
}

func TestResolve_NoSessions_ActionCreate(t *testing.T) {
	env := &domain.Environment{Name: "proj", Path: "/proj"}
	repo := makeEnvRepo(env)
	querier := &stubSessionQuerier{}

	svc := NewOpenService(repo, querier)
	plan, err := svc.Resolve(context.Background(), "proj")
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if plan.Action != domain.OpenActionCreate {
		t.Errorf("Action = %v, want %v", plan.Action, domain.OpenActionCreate)
	}
	if plan.Environment.Name != "proj" {
		t.Errorf("Environment.Name = %q, want %q", plan.Environment.Name, "proj")
	}
	if len(plan.Sessions) != 0 {
		t.Errorf("Sessions = %v, want empty", plan.Sessions)
	}
}

func TestResolve_OneSession_ActionResume(t *testing.T) {
	env := &domain.Environment{Name: "proj", Path: "/proj"}
	repo := makeEnvRepo(env)
	querier := &stubSessionQuerier{
		sessions: []domain.Session{{ID: "s1", Title: "my session"}},
	}

	svc := NewOpenService(repo, querier)
	plan, err := svc.Resolve(context.Background(), "proj")
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if plan.Action != domain.OpenActionResume {
		t.Errorf("Action = %v, want %v", plan.Action, domain.OpenActionResume)
	}
}

func TestResolve_MultipleSessions_ActionSelect(t *testing.T) {
	env := &domain.Environment{Name: "proj", Path: "/proj"}
	repo := makeEnvRepo(env)
	querier := &stubSessionQuerier{
		sessions: []domain.Session{
			{ID: "s1"},
			{ID: "s2"},
		},
	}

	svc := NewOpenService(repo, querier)
	plan, err := svc.Resolve(context.Background(), "proj")
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if plan.Action != domain.OpenActionSelect {
		t.Errorf("Action = %v, want %v", plan.Action, domain.OpenActionSelect)
	}
}

func TestResolve_DiscoveryFailure_StillSucceeds(t *testing.T) {
	// If session discovery fails, Resolve should still succeed with an empty
	// session list (discovery failure is non-fatal).
	env := &domain.Environment{Name: "proj", Path: "/proj"}
	repo := makeEnvRepo(env)
	querier := &stubSessionQuerier{listErr: errors.New("provider down")}

	svc := NewOpenService(repo, querier)
	plan, err := svc.Resolve(context.Background(), "proj")
	if err != nil {
		t.Fatalf("Resolve() should succeed even when sessions fail: %v", err)
	}
	if plan.Action != domain.OpenActionCreate {
		t.Errorf("Action = %v, want %v (no sessions)", plan.Action, domain.OpenActionCreate)
	}
}

func TestResolve_EnvNotFound_ReturnsError(t *testing.T) {
	repo := makeEnvRepo() // empty
	querier := &stubSessionQuerier{}

	svc := NewOpenService(repo, querier)
	_, err := svc.Resolve(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for missing environment, got nil")
	}
}

// ---------------------------------------------------------------------------
// OpenService.ServerForEnvironment
// ---------------------------------------------------------------------------

func TestServerForEnvironment_Found(t *testing.T) {
	env := &domain.Environment{Name: "proj", Path: "/proj"}
	repo := makeEnvRepo(env)
	querier := &stubSessionQuerier{
		servers: []domain.Server{{Port: 4000, Directory: "/proj"}},
	}

	svc := NewOpenService(repo, querier)
	srv, err := svc.ServerForEnvironment(context.Background(), "proj")
	if err != nil {
		t.Fatalf("ServerForEnvironment() error: %v", err)
	}
	if srv == nil {
		t.Fatal("expected server, got nil")
	}
	if srv.Port != 4000 {
		t.Errorf("Port = %d, want 4000", srv.Port)
	}
}

func TestServerForEnvironment_NotRunning(t *testing.T) {
	env := &domain.Environment{Name: "proj", Path: "/proj"}
	repo := makeEnvRepo(env)
	querier := &stubSessionQuerier{} // no servers

	svc := NewOpenService(repo, querier)
	srv, err := svc.ServerForEnvironment(context.Background(), "proj")
	if err != nil {
		t.Fatalf("ServerForEnvironment() error: %v", err)
	}
	if srv != nil {
		t.Errorf("expected nil server, got %+v", srv)
	}
}
