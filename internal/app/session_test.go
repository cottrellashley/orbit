package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
)

// ---------------------------------------------------------------------------
// helpers / stubs
// ---------------------------------------------------------------------------

type stubEnvRepo struct {
	envs []*domain.Environment
}

func (r *stubEnvRepo) List() ([]*domain.Environment, error) { return r.envs, nil }
func (r *stubEnvRepo) Get(name string) (*domain.Environment, error) {
	for _, e := range r.envs {
		if e.Name == name {
			return e, nil
		}
	}
	return nil, errors.New("not found")
}
func (r *stubEnvRepo) GetByPath(path string) (*domain.Environment, error) {
	for _, e := range r.envs {
		if e.Path == path {
			return e, nil
		}
	}
	return nil, errors.New("not found")
}
func (r *stubEnvRepo) Save(envs []*domain.Environment) error { r.envs = envs; return nil }
func (r *stubEnvRepo) Delete(name string) error {
	for i, e := range r.envs {
		if e.Name == name {
			r.envs = append(r.envs[:i], r.envs[i+1:]...)
			return nil
		}
	}
	return errors.New("not found")
}

type stubProvider struct {
	servers  []domain.Server
	sessions map[int][]domain.Session // key: server port
	getErr   error
}

func (p *stubProvider) DiscoverServers(_ context.Context) ([]domain.Server, error) {
	return p.servers, nil
}
func (p *stubProvider) ListSessions(_ context.Context, srv domain.Server) ([]domain.Session, error) {
	return p.sessions[srv.Port], nil
}
func (p *stubProvider) GetSession(_ context.Context, srv domain.Server, id string) (*domain.Session, error) {
	if p.getErr != nil {
		return nil, p.getErr
	}
	for _, s := range p.sessions[srv.Port] {
		if s.ID == id {
			return &s, nil
		}
	}
	return nil, errors.New("not found")
}
func (p *stubProvider) CreateSession(_ context.Context, srv domain.Server, title string) (*domain.Session, error) {
	s := &domain.Session{
		ID:        "new-session",
		Title:     title,
		ServerDir: srv.Directory,
		CreatedAt: time.Now(),
	}
	return s, nil
}
func (p *stubProvider) AbortSession(_ context.Context, _ domain.Server, _ string) error  { return nil }
func (p *stubProvider) DeleteSession(_ context.Context, _ domain.Server, _ string) error { return nil }
func (p *stubProvider) IsInstalled() bool                                                { return true }
func (p *stubProvider) Version(_ context.Context) (string, error)                        { return "v1.0.0", nil }

// ---------------------------------------------------------------------------
// matchEnvironment
// ---------------------------------------------------------------------------

func TestMatchEnvironment_NoEnvs(t *testing.T) {
	name, path := matchEnvironment("/some/dir", nil)
	if name != "" || path != "" {
		t.Errorf("expected empty match, got name=%q path=%q", name, path)
	}
}

func TestMatchEnvironment_ExactMatch(t *testing.T) {
	envs := []*domain.Environment{
		{Name: "proj", Path: "/home/user/proj"},
	}
	name, path := matchEnvironment("/home/user/proj", envs)
	if name != "proj" {
		t.Errorf("name = %q, want %q", name, "proj")
	}
	if path != "/home/user/proj" {
		t.Errorf("path = %q, want %q", path, "/home/user/proj")
	}
}

func TestMatchEnvironment_SubdirMatch(t *testing.T) {
	envs := []*domain.Environment{
		{Name: "proj", Path: "/home/user/proj"},
	}
	// Server is in a subdirectory of the environment.
	name, path := matchEnvironment("/home/user/proj/subdir", envs)
	if name != "proj" {
		t.Errorf("name = %q, want %q", name, "proj")
	}
	_ = path
}

func TestMatchEnvironment_LongestPrefixWins(t *testing.T) {
	envs := []*domain.Environment{
		{Name: "parent", Path: "/home/user"},
		{Name: "child", Path: "/home/user/proj"},
	}
	name, _ := matchEnvironment("/home/user/proj/src", envs)
	if name != "child" {
		t.Errorf("name = %q, want %q (longest prefix should win)", name, "child")
	}
}

func TestMatchEnvironment_NoMatch(t *testing.T) {
	envs := []*domain.Environment{
		{Name: "proj", Path: "/home/user/proj"},
	}
	name, path := matchEnvironment("/tmp/other", envs)
	if name != "" || path != "" {
		t.Errorf("expected no match, got name=%q path=%q", name, path)
	}
}

func TestMatchEnvironment_EmptyServerDir(t *testing.T) {
	envs := []*domain.Environment{
		{Name: "proj", Path: "/home/user/proj"},
	}
	name, path := matchEnvironment("", envs)
	if name != "" || path != "" {
		t.Errorf("expected empty match for empty dir, got name=%q path=%q", name, path)
	}
}

// ---------------------------------------------------------------------------
// pathContains
// ---------------------------------------------------------------------------

func TestPathContains(t *testing.T) {
	tests := []struct {
		dir    string
		target string
		want   bool
	}{
		{"/a/b", "/a/b", true},
		{"/a/b", "/a/b/c", true},
		{"/a/b", "/a/bc", false},
		{"/a/b", "/a", false},
		{"/", "/any", true},
	}
	for _, tc := range tests {
		got := pathContains(tc.dir, tc.target)
		if got != tc.want {
			t.Errorf("pathContains(%q, %q) = %v, want %v", tc.dir, tc.target, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// SessionService.DiscoverServers
// ---------------------------------------------------------------------------

func TestDiscoverServers_NoLifecycle(t *testing.T) {
	provider := &stubProvider{
		servers: []domain.Server{
			{Port: 4000, Directory: "/proj"},
		},
	}
	svc := NewSessionService(&stubEnvRepo{}, provider)

	servers, err := svc.DiscoverServers(context.Background())
	if err != nil {
		t.Fatalf("DiscoverServers() error: %v", err)
	}
	if len(servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(servers))
	}
}

func TestDiscoverServers_DeduplicatesByPort(t *testing.T) {
	provider := &stubProvider{
		servers: []domain.Server{
			{Port: 4000, Directory: "/proj"},
			{Port: 4001, Directory: "/other"},
		},
	}
	svc := NewSessionService(&stubEnvRepo{}, provider)

	// Inject a lifecycle whose server shares port 4000.
	lc := &stubLifecycle{srv: &domain.Server{Port: 4000, Directory: "/proj"}}
	svc.SetLifecycle(lc)

	servers, err := svc.DiscoverServers(context.Background())
	if err != nil {
		t.Fatalf("DiscoverServers() error: %v", err)
	}
	// Should deduplicate port 4000, leaving 2 total (not 3).
	if len(servers) != 2 {
		t.Errorf("expected 2 servers after dedup, got %d", len(servers))
	}
}

// stubLifecycle implements port.ServerLifecycle for tests.
type stubLifecycle struct {
	srv *domain.Server
}

func (l *stubLifecycle) Start(_ context.Context) (*domain.ManagedServer, error) { return nil, nil }
func (l *stubLifecycle) Stop(_ context.Context) error                           { return nil }
func (l *stubLifecycle) Status(_ context.Context) (*domain.ManagedServer, error) {
	return nil, nil
}
func (l *stubLifecycle) Server(_ context.Context) *domain.Server { return l.srv }

// ---------------------------------------------------------------------------
// SessionService.ListAll
// ---------------------------------------------------------------------------

func TestListAll_EnrichesWithEnvironment(t *testing.T) {
	envs := []*domain.Environment{
		{Name: "myproject", Path: "/home/user/myproject"},
	}
	provider := &stubProvider{
		servers: []domain.Server{
			{Port: 4000, Directory: "/home/user/myproject"},
		},
		sessions: map[int][]domain.Session{
			4000: {
				{ID: "s1", Title: "Session 1"},
			},
		},
	}
	svc := NewSessionService(&stubEnvRepo{envs: envs}, provider)

	sessions, err := svc.ListAll(context.Background())
	if err != nil {
		t.Fatalf("ListAll() error: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].EnvironmentName != "myproject" {
		t.Errorf("EnvironmentName = %q, want %q", sessions[0].EnvironmentName, "myproject")
	}
	if sessions[0].EnvironmentPath != "/home/user/myproject" {
		t.Errorf("EnvironmentPath = %q, want %q", sessions[0].EnvironmentPath, "/home/user/myproject")
	}
}

func TestListAll_NoServers(t *testing.T) {
	provider := &stubProvider{}
	svc := NewSessionService(&stubEnvRepo{}, provider)

	sessions, err := svc.ListAll(context.Background())
	if err != nil {
		t.Fatalf("ListAll() error: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}
}
