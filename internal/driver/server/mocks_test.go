package server

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
)

// ---------------------------------------------------------------------------
// Mock implementations of consumer-defined interfaces
// ---------------------------------------------------------------------------

type mockEnvironmentService struct {
	envs []*domain.Environment
	err  error
}

func (m *mockEnvironmentService) List() ([]*domain.Environment, error) {
	return m.envs, m.err
}
func (m *mockEnvironmentService) Register(name, path, description string) (*domain.Environment, error) {
	if m.err != nil {
		return nil, m.err
	}
	env := &domain.Environment{
		Name:        name,
		Path:        path,
		Description: description,
		CreatedAt:   time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC),
	}
	m.envs = append(m.envs, env)
	return env, nil
}
func (m *mockEnvironmentService) Delete(name string) error {
	if m.err != nil {
		return m.err
	}
	for i, e := range m.envs {
		if e.Name == name {
			m.envs = append(m.envs[:i], m.envs[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("environment %q: %w", name, domain.ErrNotFound)
}

type mockProjectService struct {
	projects []*domain.Project
	err      error
}

func (m *mockProjectService) List() ([]*domain.Project, error) {
	return m.projects, m.err
}
func (m *mockProjectService) Get(name string) (*domain.Project, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, p := range m.projects {
		if p.Name == name {
			return p, nil
		}
	}
	return nil, fmt.Errorf("project %q: %w", name, domain.ErrNotFound)
}
func (m *mockProjectService) Register(name, path, description string) (*domain.Project, error) {
	if m.err != nil {
		return nil, m.err
	}
	p := &domain.Project{
		Name:        name,
		Path:        path,
		Description: description,
		Topology:    domain.TopologyUnknown,
		CreatedAt:   time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC),
	}
	m.projects = append(m.projects, p)
	return p, nil
}
func (m *mockProjectService) Delete(name string) error {
	if m.err != nil {
		return m.err
	}
	for i, p := range m.projects {
		if p.Name == name {
			m.projects = append(m.projects[:i], m.projects[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("project %q: %w", name, domain.ErrNotFound)
}

type mockSessionService struct {
	servers  []domain.Server
	sessions []domain.Session
	err      error
}

func (m *mockSessionService) DiscoverServers(ctx context.Context) ([]domain.Server, error) {
	return m.servers, m.err
}
func (m *mockSessionService) ListAll(ctx context.Context) ([]domain.Session, error) {
	return m.sessions, m.err
}
func (m *mockSessionService) GetSession(ctx context.Context, id string) (*domain.Session, error) {
	for i := range m.sessions {
		if m.sessions[i].ID == id {
			return &m.sessions[i], nil
		}
	}
	return nil, fmt.Errorf("session %q not found", id)
}
func (m *mockSessionService) AbortSession(ctx context.Context, id string) error  { return m.err }
func (m *mockSessionService) DeleteSession(ctx context.Context, id string) error { return m.err }

type mockDoctorService struct{}

func (m *mockDoctorService) Run(ctx context.Context) *domain.Report {
	return &domain.Report{
		Results: []domain.CheckResult{
			{Name: "test-check", Status: domain.CheckPass, Message: "ok"},
		},
	}
}

type mockOpenService struct{}

func (m *mockOpenService) Resolve(ctx context.Context, envName string) (*domain.OpenPlan, error) {
	return &domain.OpenPlan{
		Environment: &domain.Environment{Name: envName, Path: "/tmp/" + envName},
		Action:      domain.OpenActionCreate,
	}, nil
}

type mockGitHubService struct {
	status *domain.GitHubAuthStatus
	repos  []domain.GitHubRepo
	issues []domain.GitHubIssue
	err    error
}

func (m *mockGitHubService) AuthStatus(ctx context.Context) (*domain.GitHubAuthStatus, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.status != nil {
		return m.status, nil
	}
	return &domain.GitHubAuthStatus{Authenticated: false}, nil
}
func (m *mockGitHubService) ListRepos(ctx context.Context) ([]domain.GitHubRepo, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.repos, nil
}
func (m *mockGitHubService) ListIssues(ctx context.Context, owner, repo string) ([]domain.GitHubIssue, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.issues, nil
}
func (m *mockGitHubService) RepoLink(owner, repo string) domain.JumpLink {
	return domain.JumpLink{
		Label: owner + "/" + repo,
		URL:   "https://github.com/" + owner + "/" + repo,
		Kind:  domain.LinkRepo,
	}
}
func (m *mockGitHubService) IssueLink(owner, repo string, number int) domain.JumpLink {
	return domain.JumpLink{
		Label: fmt.Sprintf("%s/%s#%d", owner, repo, number),
		URL:   fmt.Sprintf("https://github.com/%s/%s/issues/%d", owner, repo, number),
		Kind:  domain.LinkIssue,
	}
}
func (m *mockGitHubService) CapabilitySummary(ctx context.Context) string {
	return "GitHub: mock"
}

type mockNavigationService struct{}

func (m *mockNavigationService) RepoWebURL(repo domain.RepoInfo) *domain.JumpLink {
	return nil
}
func (m *mockNavigationService) ProjectLinks(proj domain.Project) []domain.JumpLink {
	var links []domain.JumpLink
	for _, repo := range proj.Repos {
		if repo.RemoteURL != "" {
			links = append(links, domain.JumpLink{
				Label: repo.RemoteURL,
				URL:   "https://github.com/mock",
				Kind:  domain.LinkRepo,
			})
		}
	}
	return links
}

type mockMarkdownService struct{}

func (m *mockMarkdownService) Render(ctx context.Context, source string) (*domain.RenderedMarkdown, error) {
	return &domain.RenderedMarkdown{
		Source:   source,
		HTML:     "<pre><code>" + source + "</code></pre>",
		Fallback: true,
	}, nil
}

type mockNodeService struct {
	nodes []*domain.Node
	err   error
}

func (m *mockNodeService) ListNodes(ctx context.Context) ([]*domain.Node, error) {
	return m.nodes, m.err
}
func (m *mockNodeService) GetNode(ctx context.Context, id string) (*domain.Node, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, n := range m.nodes {
		if n.ID == id {
			return n, nil
		}
	}
	return nil, fmt.Errorf("node %q: %w", id, domain.ErrNodeNotFound)
}
func (m *mockNodeService) RegisterNode(ctx context.Context, hostname string, port int, provider domain.NodeProvider, name string) (*domain.Node, error) {
	if m.err != nil {
		return nil, m.err
	}
	n := &domain.Node{
		ID:       "mock-node-id",
		Hostname: hostname,
		Port:     port,
		Provider: provider,
		Name:     name,
		Healthy:  true,
		Origin:   domain.OriginRegistered,
	}
	m.nodes = append(m.nodes, n)
	return n, nil
}
func (m *mockNodeService) RemoveNode(ctx context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	for i, n := range m.nodes {
		if n.ID == id {
			m.nodes = append(m.nodes[:i], m.nodes[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("node %q: %w", id, domain.ErrNodeNotFound)
}
func (m *mockNodeService) SyncDiscoveredNodes(ctx context.Context) ([]*domain.Node, error) {
	return m.nodes, m.err
}

type mockInstallService struct {
	tools      []domain.ToolInfo
	listErr    error
	result     domain.InstallResult
	installErr error
}

func (m *mockInstallService) ListAll(ctx context.Context) ([]domain.ToolInfo, error) {
	return m.tools, m.listErr
}

func (m *mockInstallService) Install(ctx context.Context, name string) (domain.InstallResult, error) {
	return m.result, m.installErr
}

type mockCopilotService struct {
	available bool
	tasks     []domain.CopilotTask
	task      *domain.CopilotTask
	logs      string
	listErr   error
	getErr    error
	createErr error
	stopErr   error
	logsErr   error
}

func (m *mockCopilotService) IsAvailable() bool { return m.available }

func (m *mockCopilotService) ListTasks(_ context.Context, _, _ string) ([]domain.CopilotTask, error) {
	return m.tasks, m.listErr
}

func (m *mockCopilotService) GetTask(_ context.Context, _ string) (*domain.CopilotTask, error) {
	return m.task, m.getErr
}

func (m *mockCopilotService) CreateTask(_ context.Context, _ domain.CopilotTaskCreateOpts) (*domain.CopilotTask, error) {
	return m.task, m.createErr
}

func (m *mockCopilotService) StopTask(_ context.Context, _ string) error {
	return m.stopErr
}

func (m *mockCopilotService) TaskLogs(_ context.Context, _ string) (io.ReadCloser, error) {
	if m.logsErr != nil {
		return nil, m.logsErr
	}
	return io.NopCloser(strings.NewReader(m.logs)), nil
}
