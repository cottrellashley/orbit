// Package mcpserver implements an MCP (Model Context Protocol) driver that
// exposes Orbit's capabilities as MCP tools. It runs over stdio transport
// so the chatbot's OpenCode server can invoke it as a subprocess.
package mcpserver

import (
	"context"

	"github.com/cottrellashley/orbit/internal/domain"
)

// ---------------------------------------------------------------------------
// Service interfaces — consumer-defined, decoupled from concrete app types.
// Each interface declares only the methods the MCP driver actually calls.
// ---------------------------------------------------------------------------

// ProjectService is the subset of app.ProjectService used by the MCP driver.
type ProjectService interface {
	List() ([]*domain.Project, error)
	Get(name string) (*domain.Project, error)
	Register(name, path, description string) (*domain.Project, error)
	Delete(name string) error
}

// SessionService is the subset of app.SessionService used by the MCP driver.
type SessionService interface {
	DiscoverServers(ctx context.Context) ([]domain.Server, error)
	ListAll(ctx context.Context) ([]domain.Session, error)
	GetSession(ctx context.Context, sessionID string) (*domain.Session, error)
	AbortSession(ctx context.Context, sessionID string) error
	DeleteSession(ctx context.Context, sessionID string) error
}

// NodeService is the subset of app.NodeService used by the MCP driver.
type NodeService interface {
	ListNodes(ctx context.Context) ([]*domain.Node, error)
	GetNode(ctx context.Context, id string) (*domain.Node, error)
	RegisterNode(ctx context.Context, hostname string, port int, provider domain.NodeProvider, name string) (*domain.Node, error)
	RemoveNode(ctx context.Context, id string) error
	SyncDiscoveredNodes(ctx context.Context) ([]*domain.Node, error)
}

// DoctorService is the subset of app.DoctorService used by the MCP driver.
type DoctorService interface {
	Run(ctx context.Context) *domain.Report
}

// GitHubService is the subset of app.GitHubService used by the MCP driver.
type GitHubService interface {
	AuthStatus(ctx context.Context) (*domain.GitHubAuthStatus, error)
	ListRepos(ctx context.Context) ([]domain.GitHubRepo, error)
	ListIssues(ctx context.Context, owner, repo string) ([]domain.GitHubIssue, error)
}

// NavigationService is the subset of app.NavigationService used by the MCP driver.
type NavigationService interface {
	ProjectLinks(proj domain.Project) []domain.JumpLink
}

// MarkdownService is the subset of app.MarkdownService used by the MCP driver.
type MarkdownService interface {
	Render(ctx context.Context, source string) (*domain.RenderedMarkdown, error)
}

// EnvironmentService is the subset of app.EnvironmentService used by the MCP driver.
type EnvironmentService interface {
	List() ([]*domain.Environment, error)
	Register(name, path, description string) (*domain.Environment, error)
	Delete(name string) error
}
