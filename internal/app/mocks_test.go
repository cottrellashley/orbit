package app

import (
	"context"
	"fmt"
	"sort"

	"github.com/cottrellashley/orbit/internal/domain"
)

// ---------------------------------------------------------------------------
// Mock implementations of port interfaces for app-layer tests.
// These live in the app package (test-only) so tests can construct
// services directly without importing adapters.
// ---------------------------------------------------------------------------

// mockProjectRepo is an in-memory ProjectRepository for testing.
type mockProjectRepo struct {
	projects []*domain.Project
	saveErr  error
}

func newMockProjectRepo(projects ...*domain.Project) *mockProjectRepo {
	return &mockProjectRepo{projects: projects}
}

func (m *mockProjectRepo) List() ([]*domain.Project, error) {
	if m.projects == nil {
		return []*domain.Project{}, nil
	}
	return m.projects, nil
}

func (m *mockProjectRepo) Get(name string) (*domain.Project, error) {
	for _, p := range m.projects {
		if p.Name == name {
			return p, nil
		}
	}
	return nil, fmt.Errorf("project %q: %w", name, domain.ErrNotFound)
}

func (m *mockProjectRepo) GetByPath(path string) (*domain.Project, error) {
	for _, p := range m.projects {
		if p.Path == path {
			return p, nil
		}
	}
	return nil, fmt.Errorf("project at path %q: %w", path, domain.ErrNotFound)
}

func (m *mockProjectRepo) Save(projects []*domain.Project) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.projects = projects
	return nil
}

func (m *mockProjectRepo) Delete(name string) error {
	for i, p := range m.projects {
		if p.Name == name {
			m.projects = append(m.projects[:i], m.projects[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("project %q: %w", name, domain.ErrNotFound)
}

// mockEnvRepo is an in-memory EnvironmentRepository for testing.
type mockEnvRepo struct {
	envs    []*domain.Environment
	saveErr error
}

func newMockEnvRepo(envs ...*domain.Environment) *mockEnvRepo {
	return &mockEnvRepo{envs: envs}
}

func (m *mockEnvRepo) List() ([]*domain.Environment, error) {
	if m.envs == nil {
		return []*domain.Environment{}, nil
	}
	return m.envs, nil
}

func (m *mockEnvRepo) Get(name string) (*domain.Environment, error) {
	for _, e := range m.envs {
		if e.Name == name {
			return e, nil
		}
	}
	return nil, fmt.Errorf("environment %q: %w", name, domain.ErrNotFound)
}

func (m *mockEnvRepo) GetByPath(path string) (*domain.Environment, error) {
	for _, e := range m.envs {
		if e.Path == path {
			return e, nil
		}
	}
	return nil, fmt.Errorf("environment at path %q: %w", path, domain.ErrNotFound)
}

func (m *mockEnvRepo) Save(envs []*domain.Environment) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.envs = envs
	return nil
}

func (m *mockEnvRepo) Delete(name string) error {
	for i, e := range m.envs {
		if e.Name == name {
			m.envs = append(m.envs[:i], m.envs[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("environment %q: %w", name, domain.ErrNotFound)
}

// mockSessionProvider implements port.SessionProvider for testing.
type mockSessionProvider struct {
	servers     []domain.Server
	sessions    map[int][]domain.Session // keyed by server port
	installed   bool
	version     string
	discoverErr error
}

func newMockSessionProvider() *mockSessionProvider {
	return &mockSessionProvider{
		sessions:  make(map[int][]domain.Session),
		installed: true,
		version:   "1.0.0-test",
	}
}

func (m *mockSessionProvider) DiscoverServers(_ context.Context) ([]domain.Server, error) {
	if m.discoverErr != nil {
		return nil, m.discoverErr
	}
	return m.servers, nil
}

func (m *mockSessionProvider) ListSessions(_ context.Context, node domain.Node) ([]domain.Session, error) {
	sessions, ok := m.sessions[node.Port]
	if !ok {
		return nil, nil
	}
	return sessions, nil
}

func (m *mockSessionProvider) GetSession(_ context.Context, node domain.Node, sessionID string) (*domain.Session, error) {
	for _, s := range m.sessions[node.Port] {
		if s.ID == sessionID {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("session %q not found", sessionID)
}

func (m *mockSessionProvider) CreateSession(_ context.Context, node domain.Node, title string) (*domain.Session, error) {
	s := domain.Session{
		ID:         fmt.Sprintf("sess-%d", len(m.sessions[node.Port])+1),
		Title:      title,
		NodeID:     node.ID,
		ServerDir:  node.Directory,
		ServerPort: node.Port,
		Status:     "idle",
	}
	m.sessions[node.Port] = append(m.sessions[node.Port], s)
	return &s, nil
}

func (m *mockSessionProvider) AbortSession(_ context.Context, _ domain.Node, _ string) error {
	return nil
}

func (m *mockSessionProvider) DeleteSession(_ context.Context, node domain.Node, sessionID string) error {
	sessions := m.sessions[node.Port]
	for i, s := range sessions {
		if s.ID == sessionID {
			m.sessions[node.Port] = append(sessions[:i], sessions[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("session %q not found", sessionID)
}

func (m *mockSessionProvider) IsInstalled() bool {
	return m.installed
}

func (m *mockSessionProvider) Version(_ context.Context) (string, error) {
	if m.version == "" {
		return "", fmt.Errorf("version unknown")
	}
	return m.version, nil
}

// mockProfileRepo implements port.ProfileRepository for testing.
type mockProfileRepo struct {
	profiles []*domain.Profile
	dir      string
}

func newMockProfileRepo(dir string) *mockProfileRepo {
	return &mockProfileRepo{dir: dir}
}

func (m *mockProfileRepo) List() ([]*domain.Profile, error) {
	return m.profiles, nil
}

func (m *mockProfileRepo) Get(name string) (*domain.Profile, error) {
	for _, p := range m.profiles {
		if p.Name == name {
			return p, nil
		}
	}
	return nil, fmt.Errorf("profile %q: %w", name, domain.ErrNotFound)
}

func (m *mockProfileRepo) Create(name, description string) (*domain.Profile, error) {
	p := &domain.Profile{Name: name, Description: description}
	m.profiles = append(m.profiles, p)
	return p, nil
}

func (m *mockProfileRepo) Delete(name string) error {
	for i, p := range m.profiles {
		if p.Name == name {
			m.profiles = append(m.profiles[:i], m.profiles[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("profile %q: %w", name, domain.ErrNotFound)
}

func (m *mockProfileRepo) Apply(_, _ string) error {
	return nil
}

func (m *mockProfileRepo) Path(name string) string {
	return m.dir + "/" + name
}

func (m *mockProfileRepo) Dir() string {
	return m.dir
}

// mockConfigWorkspace implements port.ConfigWorkspace for testing.
type mockConfigWorkspace struct {
	root string
}

func (m *mockConfigWorkspace) Root() string        { return m.root }
func (m *mockConfigWorkspace) ProfilesDir() string { return m.root + "/profiles" }
func (m *mockConfigWorkspace) SkillsDir() string   { return m.root + "/skills" }
func (m *mockConfigWorkspace) AgentsDir() string   { return m.root + "/agents" }
func (m *mockConfigWorkspace) PlansDir() string    { return m.root + "/plans" }
func (m *mockConfigWorkspace) MCPDir() string      { return m.root + "/mcp" }
func (m *mockConfigWorkspace) CommandsDir() string { return m.root + "/commands" }
func (m *mockConfigWorkspace) StateDir() string    { return m.root + "/state" }
func (m *mockConfigWorkspace) EnsureDirs() error   { return nil }

// sortCheckResults sorts by name for deterministic test assertions.
func sortCheckResults(results []domain.CheckResult) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})
}

// ---------------------------------------------------------------------------
// Mock implementations for Agent 6 service tests.
// ---------------------------------------------------------------------------

// mockGitHubProvider implements the githubProvider interface for testing.
type mockGitHubProvider struct {
	status *domain.GitHubAuthStatus
	repos  []domain.GitHubRepo
	issues []domain.GitHubIssue
	err    error
}

func (m *mockGitHubProvider) AuthStatus(ctx context.Context) (*domain.GitHubAuthStatus, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.status != nil {
		return m.status, nil
	}
	return &domain.GitHubAuthStatus{Authenticated: false}, nil
}
func (m *mockGitHubProvider) ListRepos(ctx context.Context) ([]domain.GitHubRepo, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.repos, nil
}
func (m *mockGitHubProvider) ListIssues(ctx context.Context, owner, repo string) ([]domain.GitHubIssue, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.issues, nil
}
func (m *mockGitHubProvider) RepoURL(owner, repo string) string {
	return fmt.Sprintf("https://github.com/%s/%s", owner, repo)
}
func (m *mockGitHubProvider) IssueURL(owner, repo string, number int) string {
	return fmt.Sprintf("https://github.com/%s/%s/issues/%d", owner, repo, number)
}

// ---------------------------------------------------------------------------
// Mock implementations for Node service tests.
// ---------------------------------------------------------------------------

// mockNodeStore is an in-memory NodeStore for testing.
type mockNodeStore struct {
	nodes   []*domain.Node
	saveErr error
	delErr  error
}

func newMockNodeStore(nodes ...*domain.Node) *mockNodeStore {
	cp := make([]*domain.Node, len(nodes))
	copy(cp, nodes)
	return &mockNodeStore{nodes: cp}
}

func (m *mockNodeStore) List(_ context.Context) ([]*domain.Node, error) {
	if m.nodes == nil {
		return []*domain.Node{}, nil
	}
	// Return copies so callers cannot mutate the store inadvertently.
	out := make([]*domain.Node, len(m.nodes))
	for i, n := range m.nodes {
		cp := *n
		out[i] = &cp
	}
	return out, nil
}

func (m *mockNodeStore) Get(_ context.Context, id string) (*domain.Node, error) {
	for _, n := range m.nodes {
		if n.ID == id {
			cp := *n
			return &cp, nil
		}
	}
	return nil, domain.ErrNodeNotFound
}

func (m *mockNodeStore) GetByHostPort(_ context.Context, hostname string, port int) (*domain.Node, error) {
	for _, n := range m.nodes {
		if n.Hostname == hostname && n.Port == port {
			cp := *n
			return &cp, nil
		}
	}
	return nil, domain.ErrNodeNotFound
}

func (m *mockNodeStore) Save(_ context.Context, node *domain.Node) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	// Upsert: replace existing by ID, or append.
	for i, n := range m.nodes {
		if n.ID == node.ID {
			cp := *node
			m.nodes[i] = &cp
			return nil
		}
	}
	cp := *node
	m.nodes = append(m.nodes, &cp)
	return nil
}

func (m *mockNodeStore) Delete(_ context.Context, id string) error {
	if m.delErr != nil {
		return m.delErr
	}
	for i, n := range m.nodes {
		if n.ID == id {
			m.nodes = append(m.nodes[:i], m.nodes[i+1:]...)
			return nil
		}
	}
	return domain.ErrNodeNotFound
}

// mockMarkdownRenderer implements the markdownRenderer interface for testing.
type mockMarkdownRenderer struct {
	fallback bool
}

func (m *mockMarkdownRenderer) Render(_ context.Context, source string) (*domain.RenderedMarkdown, error) {
	return &domain.RenderedMarkdown{
		Source:   source,
		HTML:     "<p>" + source + "</p>",
		Fallback: m.fallback,
	}, nil
}
