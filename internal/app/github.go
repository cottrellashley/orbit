package app

import (
	"context"
	"fmt"

	"github.com/cottrellashley/orbit/internal/domain"
)

// githubProvider is the consumer-defined interface for the GitHub port.
// It mirrors port.GitHubProvider but is defined here so the app layer
// does not import port types directly (hexagonal consumer-defined rule).
type githubProvider interface {
	AuthStatus(ctx context.Context) (*domain.GitHubAuthStatus, error)
	ListRepos(ctx context.Context) ([]domain.GitHubRepo, error)
	ListIssues(ctx context.Context, owner, repo string) ([]domain.GitHubIssue, error)
	RepoURL(owner, repo string) string
	IssueURL(owner, repo string, number int) string
}

// GitHubService orchestrates GitHub integration operations for drivers.
// It delegates all API work to the GitHubProvider port.
type GitHubService struct {
	gh githubProvider
}

// NewGitHubService creates a GitHubService.
func NewGitHubService(gh githubProvider) *GitHubService {
	return &GitHubService{gh: gh}
}

// AuthStatus returns the current GitHub authentication state.
func (s *GitHubService) AuthStatus(ctx context.Context) (*domain.GitHubAuthStatus, error) {
	return s.gh.AuthStatus(ctx)
}

// ListRepos returns repositories visible to the authenticated user.
func (s *GitHubService) ListRepos(ctx context.Context) ([]domain.GitHubRepo, error) {
	return s.gh.ListRepos(ctx)
}

// ListIssues returns open issues for the given owner/repo.
func (s *GitHubService) ListIssues(ctx context.Context, owner, repo string) ([]domain.GitHubIssue, error) {
	return s.gh.ListIssues(ctx, owner, repo)
}

// RepoLink returns a JumpLink for the given repository.
func (s *GitHubService) RepoLink(owner, repo string) domain.JumpLink {
	return domain.JumpLink{
		Label: fmt.Sprintf("%s/%s", owner, repo),
		URL:   s.gh.RepoURL(owner, repo),
		Kind:  domain.LinkRepo,
	}
}

// IssueLink returns a JumpLink for the given issue or PR.
func (s *GitHubService) IssueLink(owner, repo string, number int) domain.JumpLink {
	return domain.JumpLink{
		Label: fmt.Sprintf("%s/%s#%d", owner, repo, number),
		URL:   s.gh.IssueURL(owner, repo, number),
		Kind:  domain.LinkIssue,
	}
}

// CapabilitySummary returns a brief text summary of the GitHub
// integration status suitable for display in a dashboard or doctor
// report.
func (s *GitHubService) CapabilitySummary(ctx context.Context) string {
	status, err := s.gh.AuthStatus(ctx)
	if err != nil {
		return "GitHub: error checking auth"
	}
	if !status.Authenticated {
		return "GitHub: not authenticated"
	}
	return fmt.Sprintf("GitHub: authenticated as %s (via %s)", status.User, status.TokenSource)
}
