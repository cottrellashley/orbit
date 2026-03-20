package mcpserver

import (
	"context"
	"fmt"

	"github.com/cottrellashley/orbit/internal/domain"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---------------------------------------------------------------------------
// Tool input types — struct tags drive JSON Schema auto-generation
// ---------------------------------------------------------------------------

type emptyInput struct{}

type nameInput struct {
	Name string `json:"name" jsonschema:"required,description=The name of the resource."`
}

type idInput struct {
	ID string `json:"id" jsonschema:"required,description=The unique identifier."`
}

type registerProjectInput struct {
	Name        string `json:"name" jsonschema:"required,description=Project name."`
	Path        string `json:"path" jsonschema:"required,description=Filesystem path to the project."`
	Description string `json:"description" jsonschema:"description=Optional description of the project."`
}

type sessionIDInput struct {
	SessionID string `json:"session_id" jsonschema:"required,description=The session ID."`
}

type registerNodeInput struct {
	Hostname string `json:"hostname" jsonschema:"required,description=Network hostname or IP address."`
	Port     int    `json:"port" jsonschema:"required,description=TCP port the server listens on."`
	Provider string `json:"provider" jsonschema:"required,description=Server provider type (e.g. opencode)."`
	Name     string `json:"name" jsonschema:"description=Optional display name for the node."`
}

type githubIssuesInput struct {
	Owner string `json:"owner" jsonschema:"required,description=GitHub repository owner."`
	Repo  string `json:"repo" jsonschema:"required,description=GitHub repository name."`
}

type markdownInput struct {
	Source string `json:"source" jsonschema:"required,description=Markdown source text to render."`
}

type registerEnvironmentInput struct {
	Name        string `json:"name" jsonschema:"required,description=Environment name."`
	Path        string `json:"path" jsonschema:"required,description=Filesystem path."`
	Description string `json:"description" jsonschema:"description=Optional description."`
}

// ---------------------------------------------------------------------------
// Project handlers
// ---------------------------------------------------------------------------

func (s *Server) handleListProjects(ctx context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	if s.projects == nil {
		return nil, nil, fmt.Errorf("project service unavailable")
	}
	projects, err := s.projects.List()
	if err != nil {
		return nil, nil, err
	}
	out := make([]projectOut, len(projects))
	for i, p := range projects {
		out[i] = toProjectOut(p)
	}
	res, err := jsonResult(out)
	return res, nil, err
}

func (s *Server) handleGetProject(ctx context.Context, req *mcp.CallToolRequest, in nameInput) (*mcp.CallToolResult, any, error) {
	if s.projects == nil {
		return nil, nil, fmt.Errorf("project service unavailable")
	}
	p, err := s.projects.Get(in.Name)
	if err != nil {
		return nil, nil, err
	}
	res, err := jsonResult(toProjectOut(p))
	return res, nil, err
}

func (s *Server) handleRegisterProject(ctx context.Context, req *mcp.CallToolRequest, in registerProjectInput) (*mcp.CallToolResult, any, error) {
	if s.projects == nil {
		return nil, nil, fmt.Errorf("project service unavailable")
	}
	p, err := s.projects.Register(in.Name, in.Path, in.Description)
	if err != nil {
		return nil, nil, err
	}
	res, err := jsonResult(toProjectOut(p))
	return res, nil, err
}

func (s *Server) handleDeleteProject(ctx context.Context, req *mcp.CallToolRequest, in nameInput) (*mcp.CallToolResult, any, error) {
	if s.projects == nil {
		return nil, nil, fmt.Errorf("project service unavailable")
	}
	if err := s.projects.Delete(in.Name); err != nil {
		return nil, nil, err
	}
	return textResult(fmt.Sprintf("Project %q deleted.", in.Name)), nil, nil
}

// ---------------------------------------------------------------------------
// Session handlers
// ---------------------------------------------------------------------------

func (s *Server) handleListSessions(ctx context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	if s.sessions == nil {
		return nil, nil, fmt.Errorf("session service unavailable")
	}
	sessions, err := s.sessions.ListAll(ctx)
	if err != nil {
		return nil, nil, err
	}
	out := make([]sessionOut, len(sessions))
	for i, sess := range sessions {
		out[i] = toSessionOut(sess)
	}
	res, err := jsonResult(out)
	return res, nil, err
}

func (s *Server) handleGetSession(ctx context.Context, req *mcp.CallToolRequest, in sessionIDInput) (*mcp.CallToolResult, any, error) {
	if s.sessions == nil {
		return nil, nil, fmt.Errorf("session service unavailable")
	}
	sess, err := s.sessions.GetSession(ctx, in.SessionID)
	if err != nil {
		return nil, nil, err
	}
	res, err := jsonResult(toSessionOut(*sess))
	return res, nil, err
}

func (s *Server) handleAbortSession(ctx context.Context, req *mcp.CallToolRequest, in sessionIDInput) (*mcp.CallToolResult, any, error) {
	if s.sessions == nil {
		return nil, nil, fmt.Errorf("session service unavailable")
	}
	if err := s.sessions.AbortSession(ctx, in.SessionID); err != nil {
		return nil, nil, err
	}
	return textResult(fmt.Sprintf("Session %q aborted.", in.SessionID)), nil, nil
}

func (s *Server) handleDeleteSession(ctx context.Context, req *mcp.CallToolRequest, in sessionIDInput) (*mcp.CallToolResult, any, error) {
	if s.sessions == nil {
		return nil, nil, fmt.Errorf("session service unavailable")
	}
	if err := s.sessions.DeleteSession(ctx, in.SessionID); err != nil {
		return nil, nil, err
	}
	return textResult(fmt.Sprintf("Session %q deleted.", in.SessionID)), nil, nil
}

// ---------------------------------------------------------------------------
// Server handler
// ---------------------------------------------------------------------------

func (s *Server) handleListServers(ctx context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	if s.sessions == nil {
		return nil, nil, fmt.Errorf("session service unavailable")
	}
	servers, err := s.sessions.DiscoverServers(ctx)
	if err != nil {
		return nil, nil, err
	}
	out := make([]serverOut, len(servers))
	for i, srv := range servers {
		out[i] = toServerOut(srv)
	}
	res, err := jsonResult(out)
	return res, nil, err
}

// ---------------------------------------------------------------------------
// Node handlers
// ---------------------------------------------------------------------------

func (s *Server) handleListNodes(ctx context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	if s.nodes == nil {
		return nil, nil, fmt.Errorf("node service unavailable")
	}
	nodes, err := s.nodes.ListNodes(ctx)
	if err != nil {
		return nil, nil, err
	}
	out := make([]nodeOut, len(nodes))
	for i, n := range nodes {
		out[i] = toNodeOut(n)
	}
	res, err := jsonResult(out)
	return res, nil, err
}

func (s *Server) handleGetNode(ctx context.Context, req *mcp.CallToolRequest, in idInput) (*mcp.CallToolResult, any, error) {
	if s.nodes == nil {
		return nil, nil, fmt.Errorf("node service unavailable")
	}
	n, err := s.nodes.GetNode(ctx, in.ID)
	if err != nil {
		return nil, nil, err
	}
	res, err := jsonResult(toNodeOut(n))
	return res, nil, err
}

func (s *Server) handleRegisterNode(ctx context.Context, req *mcp.CallToolRequest, in registerNodeInput) (*mcp.CallToolResult, any, error) {
	if s.nodes == nil {
		return nil, nil, fmt.Errorf("node service unavailable")
	}
	provider := domain.ParseNodeProvider(in.Provider)
	n, err := s.nodes.RegisterNode(ctx, in.Hostname, in.Port, provider, in.Name)
	if err != nil {
		return nil, nil, err
	}
	res, err := jsonResult(toNodeOut(n))
	return res, nil, err
}

func (s *Server) handleRemoveNode(ctx context.Context, req *mcp.CallToolRequest, in idInput) (*mcp.CallToolResult, any, error) {
	if s.nodes == nil {
		return nil, nil, fmt.Errorf("node service unavailable")
	}
	if err := s.nodes.RemoveNode(ctx, in.ID); err != nil {
		return nil, nil, err
	}
	return textResult(fmt.Sprintf("Node %q removed.", in.ID)), nil, nil
}

func (s *Server) handleSyncNodes(ctx context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	if s.nodes == nil {
		return nil, nil, fmt.Errorf("node service unavailable")
	}
	nodes, err := s.nodes.SyncDiscoveredNodes(ctx)
	if err != nil {
		return nil, nil, err
	}
	out := make([]nodeOut, len(nodes))
	for i, n := range nodes {
		out[i] = toNodeOut(n)
	}
	res, err := jsonResult(out)
	return res, nil, err
}

// ---------------------------------------------------------------------------
// Doctor handler
// ---------------------------------------------------------------------------

func (s *Server) handleRunDoctor(ctx context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	if s.doctor == nil {
		return nil, nil, fmt.Errorf("doctor service unavailable")
	}
	report := s.doctor.Run(ctx)
	res, err := jsonResult(toReportOut(report))
	return res, nil, err
}

// ---------------------------------------------------------------------------
// GitHub handlers
// ---------------------------------------------------------------------------

func (s *Server) handleGitHubStatus(ctx context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	if s.github == nil {
		return nil, nil, fmt.Errorf("github service unavailable")
	}
	status, err := s.github.AuthStatus(ctx)
	if err != nil {
		return nil, nil, err
	}
	out := githubAuthStatusOut{
		Authenticated: status.Authenticated,
		User:          status.User,
		TokenSource:   status.TokenSource,
		Scopes:        status.Scopes,
	}
	res, err := jsonResult(out)
	return res, nil, err
}

func (s *Server) handleListGitHubRepos(ctx context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	if s.github == nil {
		return nil, nil, fmt.Errorf("github service unavailable")
	}
	repos, err := s.github.ListRepos(ctx)
	if err != nil {
		return nil, nil, err
	}
	out := make([]githubRepoOut, len(repos))
	for i, r := range repos {
		out[i] = toGitHubRepoOut(r)
	}
	res, err := jsonResult(out)
	return res, nil, err
}

func (s *Server) handleListGitHubIssues(ctx context.Context, req *mcp.CallToolRequest, in githubIssuesInput) (*mcp.CallToolResult, any, error) {
	if s.github == nil {
		return nil, nil, fmt.Errorf("github service unavailable")
	}
	issues, err := s.github.ListIssues(ctx, in.Owner, in.Repo)
	if err != nil {
		return nil, nil, err
	}
	out := make([]githubIssueOut, len(issues))
	for i, iss := range issues {
		out[i] = toGitHubIssueOut(iss)
	}
	res, err := jsonResult(out)
	return res, nil, err
}

// ---------------------------------------------------------------------------
// Navigation handler
// ---------------------------------------------------------------------------

func (s *Server) handleProjectLinks(ctx context.Context, req *mcp.CallToolRequest, in nameInput) (*mcp.CallToolResult, any, error) {
	if s.projects == nil || s.nav == nil {
		return nil, nil, fmt.Errorf("navigation service unavailable")
	}
	p, err := s.projects.Get(in.Name)
	if err != nil {
		return nil, nil, err
	}
	links := s.nav.ProjectLinks(*p)
	out := make([]jumpLinkOut, len(links))
	for i, l := range links {
		out[i] = toJumpLinkOut(l)
	}
	res, err := jsonResult(out)
	return res, nil, err
}

// ---------------------------------------------------------------------------
// Markdown handler
// ---------------------------------------------------------------------------

func (s *Server) handleRenderMarkdown(ctx context.Context, req *mcp.CallToolRequest, in markdownInput) (*mcp.CallToolResult, any, error) {
	if s.markdown == nil {
		return nil, nil, fmt.Errorf("markdown service unavailable")
	}
	rendered, err := s.markdown.Render(ctx, in.Source)
	if err != nil {
		return nil, nil, err
	}
	return textResult(rendered.HTML), nil, nil
}

// ---------------------------------------------------------------------------
// Environment handlers (legacy)
// ---------------------------------------------------------------------------

func (s *Server) handleListEnvironments(ctx context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	if s.environments == nil {
		return nil, nil, fmt.Errorf("environment service unavailable")
	}
	envs, err := s.environments.List()
	if err != nil {
		return nil, nil, err
	}
	out := make([]environmentOut, len(envs))
	for i, e := range envs {
		out[i] = toEnvironmentOut(e)
	}
	res, err := jsonResult(out)
	return res, nil, err
}

func (s *Server) handleRegisterEnvironment(ctx context.Context, req *mcp.CallToolRequest, in registerEnvironmentInput) (*mcp.CallToolResult, any, error) {
	if s.environments == nil {
		return nil, nil, fmt.Errorf("environment service unavailable")
	}
	e, err := s.environments.Register(in.Name, in.Path, in.Description)
	if err != nil {
		return nil, nil, err
	}
	res, err := jsonResult(toEnvironmentOut(e))
	return res, nil, err
}

func (s *Server) handleDeleteEnvironment(ctx context.Context, req *mcp.CallToolRequest, in nameInput) (*mcp.CallToolResult, any, error) {
	if s.environments == nil {
		return nil, nil, fmt.Errorf("environment service unavailable")
	}
	if err := s.environments.Delete(in.Name); err != nil {
		return nil, nil, err
	}
	return textResult(fmt.Sprintf("Environment %q deleted.", in.Name)), nil, nil
}
