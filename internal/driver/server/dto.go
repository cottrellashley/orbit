package server

import (
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
)

// ---------------------------------------------------------------------------
// API response DTOs — keeps JSON serialization in the driver layer
// ---------------------------------------------------------------------------

type serverJSON struct {
	PID       int    `json:"pid"`
	Port      int    `json:"port"`
	Hostname  string `json:"hostname"`
	Directory string `json:"directory"`
	Version   string `json:"version,omitempty"`
	Healthy   bool   `json:"healthy"`
}

func toServerJSON(s domain.Server) serverJSON {
	return serverJSON{
		PID:       s.PID,
		Port:      s.Port,
		Hostname:  s.Hostname,
		Directory: s.Directory,
		Version:   s.Version,
		Healthy:   s.Healthy,
	}
}

type sessionJSON struct {
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

func toSessionJSON(s domain.Session) sessionJSON {
	return sessionJSON{
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

type environmentJSON struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	ProfileName string `json:"profile_name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func toEnvironmentJSON(e *domain.Environment) environmentJSON {
	return environmentJSON{
		Name:        e.Name,
		Path:        e.Path,
		ProfileName: e.ProfileName,
		Description: e.Description,
		CreatedAt:   e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   e.UpdatedAt.Format(time.RFC3339),
	}
}

type repoInfoJSON struct {
	Path          string `json:"path"`
	RemoteURL     string `json:"remote_url,omitempty"`
	CurrentBranch string `json:"current_branch,omitempty"`
}

type projectJSON struct {
	Name         string         `json:"name"`
	Path         string         `json:"path"`
	Description  string         `json:"description"`
	ProfileName  string         `json:"profile_name,omitempty"`
	Topology     string         `json:"topology"`
	Integrations []string       `json:"integrations"`
	Repos        []repoInfoJSON `json:"repos"`
	CreatedAt    string         `json:"created_at"`
	UpdatedAt    string         `json:"updated_at"`
}

func toProjectJSON(p *domain.Project) projectJSON {
	integrations := make([]string, len(p.Integrations))
	for i, t := range p.Integrations {
		integrations[i] = string(t)
	}
	repos := make([]repoInfoJSON, len(p.Repos))
	for i, r := range p.Repos {
		repos[i] = repoInfoJSON{
			Path:          r.Path,
			RemoteURL:     r.RemoteURL,
			CurrentBranch: r.CurrentBranch,
		}
	}
	return projectJSON{
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

// ---------------------------------------------------------------------------
// GitHub DTOs
// ---------------------------------------------------------------------------

type githubAuthStatusJSON struct {
	Authenticated bool     `json:"authenticated"`
	User          string   `json:"user,omitempty"`
	TokenSource   string   `json:"token_source,omitempty"`
	Scopes        []string `json:"scopes,omitempty"`
}

type githubRepoJSON struct {
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

func toGitHubRepoJSON(r domain.GitHubRepo) githubRepoJSON {
	return githubRepoJSON{
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

type githubIssueJSON struct {
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

func toGitHubIssueJSON(i domain.GitHubIssue) githubIssueJSON {
	return githubIssueJSON{
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

type jumpLinkJSON struct {
	Label string `json:"label"`
	URL   string `json:"url"`
	Kind  string `json:"kind"`
}

func toJumpLinkJSON(l domain.JumpLink) jumpLinkJSON {
	return jumpLinkJSON{
		Label: l.Label,
		URL:   l.URL,
		Kind:  string(l.Kind),
	}
}

// ---------------------------------------------------------------------------
// Node DTOs
// ---------------------------------------------------------------------------

type nodeJSON struct {
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

func toNodeJSON(n *domain.Node) nodeJSON {
	return nodeJSON{
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
