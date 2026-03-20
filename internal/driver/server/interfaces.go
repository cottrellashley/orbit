package server

import (
	"context"
	"io"

	"github.com/cottrellashley/orbit/internal/domain"
	"github.com/cottrellashley/orbit/internal/port"
)

// ---------------------------------------------------------------------------
// Service interfaces — consumer-defined, decoupled from concrete app types
// ---------------------------------------------------------------------------

// EnvironmentService is the subset of app.EnvironmentService used by the HTTP driver.
type EnvironmentService interface {
	List() ([]*domain.Environment, error)
	Register(name, path, description string) (*domain.Environment, error)
	Delete(name string) error
}

// ProjectService is the subset of app.ProjectService used by the HTTP driver.
type ProjectService interface {
	List() ([]*domain.Project, error)
	Get(name string) (*domain.Project, error)
	Register(name, path, description string) (*domain.Project, error)
	Delete(name string) error
}

// SessionService is the subset of app.SessionService used by the HTTP driver.
type SessionService interface {
	DiscoverServers(ctx context.Context) ([]domain.Server, error)
	ListAll(ctx context.Context) ([]domain.Session, error)
	GetSession(ctx context.Context, sessionID string) (*domain.Session, error)
	AbortSession(ctx context.Context, sessionID string) error
	DeleteSession(ctx context.Context, sessionID string) error
}

// DoctorService is the subset of app.DoctorService used by the HTTP driver.
type DoctorService interface {
	Run(ctx context.Context) *domain.Report
}

// OpenService is the subset of app.OpenService used by the HTTP driver.
type OpenService interface {
	Resolve(ctx context.Context, envName string) (*domain.OpenPlan, error)
}

// GitHubService is the subset of app.GitHubService used by the HTTP driver.
type GitHubService interface {
	AuthStatus(ctx context.Context) (*domain.GitHubAuthStatus, error)
	ListRepos(ctx context.Context) ([]domain.GitHubRepo, error)
	ListIssues(ctx context.Context, owner, repo string) ([]domain.GitHubIssue, error)
	RepoLink(owner, repo string) domain.JumpLink
	IssueLink(owner, repo string, number int) domain.JumpLink
	CapabilitySummary(ctx context.Context) string
}

// NavigationService is the subset of app.NavigationService used by the HTTP driver.
type NavigationService interface {
	RepoWebURL(repo domain.RepoInfo) *domain.JumpLink
	ProjectLinks(proj domain.Project) []domain.JumpLink
}

// MarkdownService is the subset of app.MarkdownService used by the HTTP driver.
type MarkdownService interface {
	Render(ctx context.Context, source string) (*domain.RenderedMarkdown, error)
}

// NodeService is the subset of app.NodeService used by the HTTP driver.
type NodeService interface {
	ListNodes(ctx context.Context) ([]*domain.Node, error)
	GetNode(ctx context.Context, id string) (*domain.Node, error)
	RegisterNode(ctx context.Context, hostname string, port int, provider domain.NodeProvider, name string) (*domain.Node, error)
	RemoveNode(ctx context.Context, id string) error
	SyncDiscoveredNodes(ctx context.Context) ([]*domain.Node, error)
}

// TerminalService is the subset of app.TerminalService used by the HTTP driver.
type TerminalService interface {
	Spawn(ctx context.Context, opts domain.TerminalSpawnOpts) (*domain.Terminal, error)
	List(ctx context.Context) []domain.Terminal
	Kill(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*domain.Terminal, error)
	Attach(ctx context.Context, id string) (port.TerminalConn, error)
}

// InstallService is the subset of app.InstallService used by the HTTP driver.
type InstallService interface {
	ListAll(ctx context.Context) ([]domain.ToolInfo, error)
	Install(ctx context.Context, name string) (domain.InstallResult, error)
}

// CopilotService is the subset of app.CopilotService used by the HTTP driver.
type CopilotService interface {
	IsAvailable() bool
	ListTasks(ctx context.Context, owner, repo string) ([]domain.CopilotTask, error)
	GetTask(ctx context.Context, sessionID string) (*domain.CopilotTask, error)
	CreateTask(ctx context.Context, opts domain.CopilotTaskCreateOpts) (*domain.CopilotTask, error)
	StopTask(ctx context.Context, sessionID string) error
	TaskLogs(ctx context.Context, sessionID string) (io.ReadCloser, error)
}
