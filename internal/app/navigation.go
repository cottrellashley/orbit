package app

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/cottrellashley/orbit/internal/domain"
)

// NavigationService provides reusable jump-link helpers for outbound
// navigation. All methods are pure functions on domain data — no
// network calls are made. This service is usable by CLI, TUI, and
// web drivers alike.
type NavigationService struct{}

// NewNavigationService creates a NavigationService.
func NewNavigationService() *NavigationService {
	return &NavigationService{}
}

// RepoWebURL returns a JumpLink to the web view of a git repository.
// It parses both HTTPS and SSH remote URLs. Returns nil if the remote
// URL cannot be parsed into a web URL.
func (n *NavigationService) RepoWebURL(repo domain.RepoInfo) *domain.JumpLink {
	webURL := remoteToWebURL(repo.RemoteURL)
	if webURL == "" {
		return nil
	}
	return &domain.JumpLink{
		Label: repoLabel(webURL),
		URL:   webURL,
		Kind:  domain.LinkRepo,
	}
}

// BranchURL returns a JumpLink to view a specific branch on the web.
// Returns nil if the remote URL cannot be parsed.
func (n *NavigationService) BranchURL(repo domain.RepoInfo) *domain.JumpLink {
	webURL := remoteToWebURL(repo.RemoteURL)
	if webURL == "" || repo.CurrentBranch == "" {
		return nil
	}
	return &domain.JumpLink{
		Label: repo.CurrentBranch,
		URL:   webURL + "/tree/" + repo.CurrentBranch,
		Kind:  domain.LinkBranch,
	}
}

// IssueListURL returns a JumpLink to the issue list for a repository.
// Returns nil if the remote URL cannot be parsed.
func (n *NavigationService) IssueListURL(repo domain.RepoInfo) *domain.JumpLink {
	webURL := remoteToWebURL(repo.RemoteURL)
	if webURL == "" {
		return nil
	}
	return &domain.JumpLink{
		Label: "Issues",
		URL:   webURL + "/issues",
		Kind:  domain.LinkIssue,
	}
}

// PRListURL returns a JumpLink to the pull request list for a repository.
// Returns nil if the remote URL cannot be parsed.
func (n *NavigationService) PRListURL(repo domain.RepoInfo) *domain.JumpLink {
	webURL := remoteToWebURL(repo.RemoteURL)
	if webURL == "" {
		return nil
	}
	return &domain.JumpLink{
		Label: "Pull Requests",
		URL:   webURL + "/pulls",
		Kind:  domain.LinkPR,
	}
}

// SessionWebURL returns a JumpLink to the Orbit web UI for a session.
// orbitBaseURL is the base URL of the running Orbit server (e.g.
// "http://127.0.0.1:3000").
func (n *NavigationService) SessionWebURL(sess domain.Session, orbitBaseURL string) *domain.JumpLink {
	u := strings.TrimRight(orbitBaseURL, "/") + fmt.Sprintf("/#/sessions/%s", sess.ID)
	return &domain.JumpLink{
		Label: sess.Title,
		URL:   u,
		Kind:  domain.LinkSession,
	}
}

// ProjectLinks returns all available JumpLinks for a project by
// examining its repos. Returns an empty slice if no links can be
// derived.
func (n *NavigationService) ProjectLinks(proj domain.Project) []domain.JumpLink {
	var links []domain.JumpLink
	for _, repo := range proj.Repos {
		if link := n.RepoWebURL(repo); link != nil {
			links = append(links, *link)
		}
		if link := n.BranchURL(repo); link != nil {
			links = append(links, *link)
		}
		if link := n.IssueListURL(repo); link != nil {
			links = append(links, *link)
		}
		if link := n.PRListURL(repo); link != nil {
			links = append(links, *link)
		}
	}
	return links
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// remoteToWebURL converts a git remote URL (HTTPS or SSH) to a web URL.
// Returns empty string on failure.
func remoteToWebURL(remote string) string {
	if remote == "" {
		return ""
	}

	// SSH format: git@github.com:owner/repo.git
	if strings.HasPrefix(remote, "git@") {
		remote = strings.TrimPrefix(remote, "git@")
		remote = strings.Replace(remote, ":", "/", 1)
		remote = strings.TrimSuffix(remote, ".git")
		return "https://" + remote
	}

	// HTTPS format: https://github.com/owner/repo.git
	u, err := url.Parse(remote)
	if err != nil {
		return ""
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return ""
	}
	path := strings.TrimSuffix(u.Path, ".git")
	return fmt.Sprintf("https://%s%s", u.Host, path)
}

// repoLabel extracts a short label from a web URL (e.g. "owner/repo").
func repoLabel(webURL string) string {
	u, err := url.Parse(webURL)
	if err != nil {
		return webURL
	}
	return strings.TrimPrefix(u.Path, "/")
}
