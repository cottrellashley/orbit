package port

import (
	"context"

	"github.com/cottrellashley/orbit/internal/domain"
)

// GitHubProvider abstracts access to the GitHub API. Adapters implement
// this interface using REST API calls authenticated via a personal access
// token. The token resolution strategy (env var, config file, gh CLI) is
// an adapter concern — consumers only see the domain types.
//
// All methods accept a context for cancellation and timeout control.
type GitHubProvider interface {
	// AuthStatus returns the current authentication state. If no token
	// can be resolved, the returned status has Authenticated == false
	// and a nil error. A non-nil error indicates an unexpected failure
	// (e.g. network error while validating the token).
	AuthStatus(ctx context.Context) (*domain.GitHubAuthStatus, error)

	// ListRepos returns repositories visible to the authenticated user.
	// Returns domain.ErrNotAuthenticated when no token is available.
	ListRepos(ctx context.Context) ([]domain.GitHubRepo, error)

	// ListIssues returns open issues for the given "owner/repo".
	// Returns domain.ErrNotAuthenticated when no token is available.
	ListIssues(ctx context.Context, owner, repo string) ([]domain.GitHubIssue, error)

	// RepoURL returns the canonical web URL for the given "owner/repo".
	// This is a pure derivation — no network call required.
	RepoURL(owner, repo string) string

	// IssueURL returns the canonical web URL for an issue/PR number.
	// This is a pure derivation — no network call required.
	IssueURL(owner, repo string, number int) string
}
