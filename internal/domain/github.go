package domain

import "time"

// GitHubAuthStatus describes the current state of GitHub authentication.
type GitHubAuthStatus struct {
	// Authenticated is true when a valid token was resolved.
	Authenticated bool
	// User is the GitHub login name (empty when not authenticated).
	User string
	// TokenSource describes where the token was loaded from, e.g.
	// "GITHUB_TOKEN", "GH_TOKEN", "config-file", "gh-cli".
	// Empty when not authenticated.
	TokenSource string
	// Scopes lists the OAuth scopes reported by GitHub for the token.
	// May be empty if the token is a fine-grained PAT (which has no
	// classic scopes header).
	Scopes []string
}

// GitHubRepo is a minimal representation of a GitHub repository.
// It carries only the fields Orbit needs — listing and linking.
type GitHubRepo struct {
	// Owner is the repository owner login (user or org).
	Owner string
	// Name is the repository name (without owner prefix).
	Name string
	// FullName is "owner/name".
	FullName string
	// Description is the repo description (may be empty).
	Description string
	// HTMLURL is the canonical web URL, e.g. "https://github.com/owner/name".
	HTMLURL string
	// CloneURL is the HTTPS clone URL.
	CloneURL string
	// SSHURL is the SSH clone URL.
	SSHURL string
	// DefaultBranch is the repository's default branch.
	DefaultBranch string
	// Private indicates whether the repository is private.
	Private bool
	// Fork indicates whether the repository is a fork.
	Fork bool
	// Archived indicates whether the repository is archived.
	Archived bool
	// UpdatedAt is when the repository was last pushed to or updated.
	UpdatedAt time.Time
}

// GitHubIssue is a minimal representation of a GitHub issue (or pull request).
type GitHubIssue struct {
	// Number is the issue/PR number within the repository.
	Number int
	// Title is the issue/PR title.
	Title string
	// State is "open" or "closed".
	State string
	// HTMLURL is the canonical web URL.
	HTMLURL string
	// User is the login of the user who created the issue.
	User string
	// Labels contains label names.
	Labels []string
	// IsPullRequest is true when the issue is actually a pull request.
	IsPullRequest bool
	// CreatedAt is when the issue was opened.
	CreatedAt time.Time
	// UpdatedAt is when the issue was last updated.
	UpdatedAt time.Time
}
