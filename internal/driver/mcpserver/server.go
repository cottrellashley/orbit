package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Version is embedded in the MCP server implementation metadata.
const Version = "0.1.0"

// Server wraps an MCP server and the Orbit services it delegates to.
type Server struct {
	mcp *mcp.Server

	projects     ProjectService
	sessions     SessionService
	nodes        NodeService
	doctor       DoctorService
	github       GitHubService
	nav          NavigationService
	markdown     MarkdownService
	environments EnvironmentService
}

// New creates a new MCP Server with all Orbit tools registered.
// Nil service arguments are permitted — the corresponding tools will
// return an "unavailable" error at call time.
func New(
	projects ProjectService,
	sessions SessionService,
	nodes NodeService,
	doctor DoctorService,
	github GitHubService,
	nav NavigationService,
	markdown MarkdownService,
	environments EnvironmentService,
) *Server {
	s := &Server{
		projects:     projects,
		sessions:     sessions,
		nodes:        nodes,
		doctor:       doctor,
		github:       github,
		nav:          nav,
		markdown:     markdown,
		environments: environments,
	}

	s.mcp = mcp.NewServer(
		&mcp.Implementation{
			Name:    "orbit",
			Version: Version,
		},
		nil, // default options
	)

	s.registerTools()
	return s
}

// Run starts the MCP server on the given transport (typically stdio).
// It blocks until the client disconnects or the context is cancelled.
func (s *Server) Run(ctx context.Context, t mcp.Transport) error {
	return s.mcp.Run(ctx, t)
}

// ---------------------------------------------------------------------------
// Tool registration
// ---------------------------------------------------------------------------

func (s *Server) registerTools() {
	// --- Projects -----------------------------------------------------------
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "list_projects",
		Description: "List all registered projects.",
	}, s.handleListProjects)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_project",
		Description: "Get a project by name.",
	}, s.handleGetProject)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "register_project",
		Description: "Register a new project.",
	}, s.handleRegisterProject)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "delete_project",
		Description: "Delete a project by name.",
	}, s.handleDeleteProject)

	// --- Sessions -----------------------------------------------------------
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "list_sessions",
		Description: "List all sessions across all nodes.",
	}, s.handleListSessions)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_session",
		Description: "Get a session by ID.",
	}, s.handleGetSession)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "abort_session",
		Description: "Abort a running session.",
	}, s.handleAbortSession)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "delete_session",
		Description: "Delete a session by ID.",
	}, s.handleDeleteSession)

	// --- Servers ------------------------------------------------------------
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "list_servers",
		Description: "Discover running OpenCode servers.",
	}, s.handleListServers)

	// --- Nodes --------------------------------------------------------------
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "list_nodes",
		Description: "List all registered nodes (AI agent servers).",
	}, s.handleListNodes)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_node",
		Description: "Get a node by ID.",
	}, s.handleGetNode)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "register_node",
		Description: "Register a new node (AI agent server).",
	}, s.handleRegisterNode)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "remove_node",
		Description: "Remove a node by ID.",
	}, s.handleRemoveNode)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "sync_nodes",
		Description: "Trigger node discovery — scans for running servers and syncs the registry.",
	}, s.handleSyncNodes)

	// --- Doctor -------------------------------------------------------------
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "run_doctor",
		Description: "Run diagnostic checks on the Orbit installation.",
	}, s.handleRunDoctor)

	// --- GitHub -------------------------------------------------------------
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "github_status",
		Description: "Check GitHub authentication status.",
	}, s.handleGitHubStatus)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "list_github_repos",
		Description: "List GitHub repositories accessible to the authenticated user.",
	}, s.handleListGitHubRepos)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "list_github_issues",
		Description: "List issues for a GitHub repository.",
	}, s.handleListGitHubIssues)

	// --- Navigation ---------------------------------------------------------
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "project_links",
		Description: "Get navigation links for a project.",
	}, s.handleProjectLinks)

	// --- Markdown -----------------------------------------------------------
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "render_markdown",
		Description: "Render markdown source to HTML.",
	}, s.handleRenderMarkdown)

	// --- Environments (legacy) ----------------------------------------------
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "list_environments",
		Description: "List all registered environments (legacy — use projects instead).",
	}, s.handleListEnvironments)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "register_environment",
		Description: "Register a new environment (legacy — use projects instead).",
	}, s.handleRegisterEnvironment)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "delete_environment",
		Description: "Delete an environment by name (legacy — use projects instead).",
	}, s.handleDeleteEnvironment)
}

// ---------------------------------------------------------------------------
// Output helpers — shared types returned as structured output from tools
// ---------------------------------------------------------------------------

type projectOut struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	Description  string    `json:"description"`
	ProfileName  string    `json:"profile_name,omitempty"`
	Topology     string    `json:"topology"`
	Integrations []string  `json:"integrations"`
	Repos        []repoOut `json:"repos"`
	CreatedAt    string    `json:"created_at"`
	UpdatedAt    string    `json:"updated_at"`
}

type repoOut struct {
	Path          string `json:"path"`
	RemoteURL     string `json:"remote_url,omitempty"`
	CurrentBranch string `json:"current_branch,omitempty"`
}

func toProjectOut(p *domain.Project) projectOut {
	integrations := make([]string, len(p.Integrations))
	for i, t := range p.Integrations {
		integrations[i] = string(t)
	}
	repos := make([]repoOut, len(p.Repos))
	for i, r := range p.Repos {
		repos[i] = repoOut{
			Path:          r.Path,
			RemoteURL:     r.RemoteURL,
			CurrentBranch: r.CurrentBranch,
		}
	}
	return projectOut{
		Name:         p.Name,
		Path:         p.Path,
		Description:  p.Description,
		ProfileName:  p.ProfileName,
		Topology:     p.Topology.String(),
		Integrations: integrations,
		Repos:        repos,
		CreatedAt:    p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    p.UpdatedAt.Format(time.RFC3339),
	}
}

type sessionOut struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	NodeID          string `json:"node_id,omitempty"`
	EnvironmentName string `json:"environment_name"`
	EnvironmentPath string `json:"environment_path"`
	ServerDir       string `json:"server_dir"`
	ServerPort      int    `json:"server_port"`
	Status          string `json:"status"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

func toSessionOut(s domain.Session) sessionOut {
	return sessionOut{
		ID:              s.ID,
		Title:           s.Title,
		NodeID:          s.NodeID,
		EnvironmentName: s.EnvironmentName,
		EnvironmentPath: s.EnvironmentPath,
		ServerDir:       s.ServerDir,
		ServerPort:      s.ServerPort,
		Status:          s.Status,
		CreatedAt:       s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       s.UpdatedAt.Format(time.RFC3339),
	}
}

type serverOut struct {
	PID       int    `json:"pid"`
	Port      int    `json:"port"`
	Hostname  string `json:"hostname"`
	Directory string `json:"directory"`
	Version   string `json:"version,omitempty"`
	Healthy   bool   `json:"healthy"`
}

func toServerOut(s domain.Server) serverOut {
	return serverOut{
		PID:       s.PID,
		Port:      s.Port,
		Hostname:  s.Hostname,
		Directory: s.Directory,
		Version:   s.Version,
		Healthy:   s.Healthy,
	}
}

type nodeOut struct {
	ID        string `json:"id"`
	Name      string `json:"name,omitempty"`
	Provider  string `json:"provider"`
	Origin    string `json:"origin"`
	Hostname  string `json:"hostname"`
	Port      int    `json:"port"`
	Directory string `json:"directory,omitempty"`
	Version   string `json:"version,omitempty"`
	PID       int    `json:"pid,omitempty"`
	Healthy   bool   `json:"healthy"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toNodeOut(n *domain.Node) nodeOut {
	return nodeOut{
		ID:        n.ID,
		Name:      n.Name,
		Provider:  n.Provider.String(),
		Origin:    n.Origin.String(),
		Hostname:  n.Hostname,
		Port:      n.Port,
		Directory: n.Directory,
		Version:   n.Version,
		PID:       n.PID,
		Healthy:   n.Healthy,
		CreatedAt: n.CreatedAt.Format(time.RFC3339),
		UpdatedAt: n.UpdatedAt.Format(time.RFC3339),
	}
}

type checkResultOut struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Fix     string `json:"fix,omitempty"`
}

type reportOut struct {
	Results []checkResultOut `json:"results"`
	OK      bool             `json:"ok"`
}

func toReportOut(r *domain.Report) reportOut {
	results := make([]checkResultOut, len(r.Results))
	for i, c := range r.Results {
		var status string
		switch c.Status {
		case domain.CheckPass:
			status = "pass"
		case domain.CheckWarn:
			status = "warn"
		case domain.CheckFail:
			status = "fail"
		default:
			status = "unknown"
		}
		results[i] = checkResultOut{
			Name:    c.Name,
			Status:  status,
			Message: c.Message,
			Fix:     c.Fix,
		}
	}
	return reportOut{Results: results, OK: r.OK()}
}

type githubAuthStatusOut struct {
	Authenticated bool     `json:"authenticated"`
	User          string   `json:"user,omitempty"`
	TokenSource   string   `json:"token_source,omitempty"`
	Scopes        []string `json:"scopes,omitempty"`
}

type githubRepoOut struct {
	Owner         string `json:"owner"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Description   string `json:"description,omitempty"`
	HTMLURL       string `json:"html_url"`
	CloneURL      string `json:"clone_url"`
	SSHURL        string `json:"ssh_url"`
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
	Fork          bool   `json:"fork"`
	Archived      bool   `json:"archived"`
	UpdatedAt     string `json:"updated_at"`
}

func toGitHubRepoOut(r domain.GitHubRepo) githubRepoOut {
	return githubRepoOut{
		Owner:         r.Owner,
		Name:          r.Name,
		FullName:      r.FullName,
		Description:   r.Description,
		HTMLURL:       r.HTMLURL,
		CloneURL:      r.CloneURL,
		SSHURL:        r.SSHURL,
		DefaultBranch: r.DefaultBranch,
		Private:       r.Private,
		Fork:          r.Fork,
		Archived:      r.Archived,
		UpdatedAt:     r.UpdatedAt.Format(time.RFC3339),
	}
}

type githubIssueOut struct {
	Number        int      `json:"number"`
	Title         string   `json:"title"`
	State         string   `json:"state"`
	HTMLURL       string   `json:"html_url"`
	User          string   `json:"user"`
	Labels        []string `json:"labels"`
	IsPullRequest bool     `json:"is_pull_request"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

func toGitHubIssueOut(i domain.GitHubIssue) githubIssueOut {
	return githubIssueOut{
		Number:        i.Number,
		Title:         i.Title,
		State:         i.State,
		HTMLURL:       i.HTMLURL,
		User:          i.User,
		Labels:        i.Labels,
		IsPullRequest: i.IsPullRequest,
		CreatedAt:     i.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     i.UpdatedAt.Format(time.RFC3339),
	}
}

type jumpLinkOut struct {
	Label string `json:"label"`
	URL   string `json:"url"`
	Kind  string `json:"kind"`
}

func toJumpLinkOut(l domain.JumpLink) jumpLinkOut {
	return jumpLinkOut{
		Label: l.Label,
		URL:   l.URL,
		Kind:  string(l.Kind),
	}
}

type environmentOut struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	ProfileName string `json:"profile_name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func toEnvironmentOut(e *domain.Environment) environmentOut {
	return environmentOut{
		Name:        e.Name,
		Path:        e.Path,
		ProfileName: e.ProfileName,
		Description: e.Description,
		CreatedAt:   e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   e.UpdatedAt.Format(time.RFC3339),
	}
}

// textResult is a convenience for returning a plain text tool result.
func textResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}
}

// jsonResult marshals v as indented JSON and returns it as text content.
func jsonResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal result: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil
}
