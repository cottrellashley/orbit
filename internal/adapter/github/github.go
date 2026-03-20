package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	gogithub "github.com/google/go-github/v84/github"

	"github.com/cottrellashley/orbit/internal/domain"
)

const requestTimeout = 15 * time.Second

// Client implements port.GitHubProvider using the google/go-github SDK.
type Client struct {
	baseURL   string
	configDir string
	http      *http.Client

	// tokenFunc allows injecting a custom token resolver for tests.
	// When nil the default cascade is used.
	tokenFunc func() string

	// cmdRunner allows injecting a custom command executor for tests.
	// Signature matches exec.CommandContext return.
	cmdRunner func(ctx context.Context, name string, args ...string) ([]byte, error)
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the GitHub API base URL (useful for GitHub
// Enterprise or test servers).
func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(u, "/") }
}

// WithHTTPClient overrides the default HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.http = hc }
}

// WithTokenFunc overrides the token resolution function (for tests).
func WithTokenFunc(fn func() string) Option {
	return func(c *Client) { c.tokenFunc = fn }
}

// WithCmdRunner overrides the command runner used for `gh auth token`
// fallback (for tests).
func WithCmdRunner(fn func(ctx context.Context, name string, args ...string) ([]byte, error)) Option {
	return func(c *Client) { c.cmdRunner = fn }
}

// NewClient creates a GitHub adapter. configDir is the Orbit config
// directory (e.g. ~/.config/orbit) used for token file lookup.
func NewClient(configDir string, opts ...Option) *Client {
	c := &Client{
		configDir: configDir,
		http: &http.Client{
			Timeout: requestTimeout,
		},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// ---------------------------------------------------------------------------
// Token resolution
// ---------------------------------------------------------------------------

// resolveToken returns the first available token from the cascade.
// Returns empty string if no token is found. Tokens are never logged.
func (c *Client) resolveToken() string {
	if c.tokenFunc != nil {
		return c.tokenFunc()
	}

	// 1. GITHUB_TOKEN env var
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return strings.TrimSpace(t)
	}

	// 2. GH_TOKEN env var
	if t := os.Getenv("GH_TOKEN"); t != "" {
		return strings.TrimSpace(t)
	}

	// 3. Config file
	if c.configDir != "" {
		tokenFile := filepath.Join(c.configDir, "github-token")
		if data, err := os.ReadFile(tokenFile); err == nil {
			if t := strings.TrimSpace(string(data)); t != "" {
				return t
			}
		}
	}

	// 4. gh auth token command
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, err := c.runCmd(ctx, "gh", "auth", "token")
	if err == nil {
		if t := strings.TrimSpace(string(out)); t != "" {
			return t
		}
	}

	return ""
}

// runCmd executes an external command and returns its stdout.
func (c *Client) runCmd(ctx context.Context, name string, args ...string) ([]byte, error) {
	if c.cmdRunner != nil {
		return c.cmdRunner(ctx, name, args...)
	}
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Output()
}

// tokenSource returns a human-readable label for the token source.
// Used only in GitHubAuthStatus — never includes the token value.
func (c *Client) tokenSource() string {
	if c.tokenFunc != nil {
		t := c.tokenFunc()
		if t != "" {
			return "injected"
		}
		return ""
	}

	if os.Getenv("GITHUB_TOKEN") != "" {
		return "GITHUB_TOKEN"
	}
	if os.Getenv("GH_TOKEN") != "" {
		return "GH_TOKEN"
	}
	if c.configDir != "" {
		tokenFile := filepath.Join(c.configDir, "github-token")
		if data, err := os.ReadFile(tokenFile); err == nil {
			if strings.TrimSpace(string(data)) != "" {
				return "config-file"
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := c.runCmd(ctx, "gh", "auth", "token")
	if err == nil && strings.TrimSpace(string(out)) != "" {
		return "gh-cli"
	}

	return ""
}

// ---------------------------------------------------------------------------
// SDK client factory
// ---------------------------------------------------------------------------

// sdkClient creates a go-github client authenticated with the current
// token. The token is resolved fresh on every call so that rotations
// are picked up automatically. If baseURL is set (tests / GHE), the
// client's BaseURL is overridden.
func (c *Client) sdkClient(token string) *gogithub.Client {
	gh := gogithub.NewClient(c.http).WithAuthToken(token)

	if c.baseURL != "" {
		// go-github expects a trailing slash on BaseURL.
		u, err := url.Parse(strings.TrimRight(c.baseURL, "/") + "/")
		if err == nil {
			gh.BaseURL = u
		}
	}
	return gh
}

// ---------------------------------------------------------------------------
// Error mapping
// ---------------------------------------------------------------------------

// mapError converts go-github errors to domain sentinel errors where
// applicable. Unrecognised errors are returned as-is.
func mapError(err error) error {
	if err == nil {
		return nil
	}
	var rlErr *gogithub.RateLimitError
	if errors.As(err, &rlErr) {
		return domain.ErrRateLimited
	}
	return err
}

// ---------------------------------------------------------------------------
// GitHubProvider implementation
// ---------------------------------------------------------------------------

// AuthStatus checks the current GitHub authentication state.
func (c *Client) AuthStatus(ctx context.Context) (*domain.GitHubAuthStatus, error) {
	token := c.resolveToken()
	if token == "" {
		return &domain.GitHubAuthStatus{Authenticated: false}, nil
	}

	gh := c.sdkClient(token)
	user, resp, err := gh.Users.Get(ctx, "")
	if err != nil {
		// 401 means the token is invalid — not an error, just
		// unauthenticated.
		var errResp *gogithub.ErrorResponse
		if errors.As(err, &errResp) && errResp.Response != nil &&
			errResp.Response.StatusCode == http.StatusUnauthorized {
			return &domain.GitHubAuthStatus{Authenticated: false}, nil
		}
		return nil, mapError(err)
	}

	// Parse scopes from response header (classic PATs only).
	var scopes []string
	if resp != nil && resp.Response != nil {
		if scopeHeader := resp.Response.Header.Get("X-OAuth-Scopes"); scopeHeader != "" {
			for _, s := range strings.Split(scopeHeader, ",") {
				s = strings.TrimSpace(s)
				if s != "" {
					scopes = append(scopes, s)
				}
			}
		}
	}

	return &domain.GitHubAuthStatus{
		Authenticated: true,
		User:          user.GetLogin(),
		TokenSource:   c.tokenSource(),
		Scopes:        scopes,
	}, nil
}

// ListRepos returns repositories visible to the authenticated user.
func (c *Client) ListRepos(ctx context.Context) ([]domain.GitHubRepo, error) {
	token := c.resolveToken()
	if token == "" {
		return nil, domain.ErrNotAuthenticated
	}

	gh := c.sdkClient(token)
	sdkRepos, _, err := gh.Repositories.ListByAuthenticatedUser(ctx,
		&gogithub.RepositoryListByAuthenticatedUserOptions{
			Sort:      "updated",
			Direction: "desc",
			ListOptions: gogithub.ListOptions{
				PerPage: 100,
			},
		})
	if err != nil {
		return nil, mapError(err)
	}

	repos := make([]domain.GitHubRepo, len(sdkRepos))
	for i, r := range sdkRepos {
		repos[i] = repoToDomain(r)
	}
	return repos, nil
}

// ListIssues returns open issues for the given owner/repo.
func (c *Client) ListIssues(ctx context.Context, owner, repo string) ([]domain.GitHubIssue, error) {
	token := c.resolveToken()
	if token == "" {
		return nil, domain.ErrNotAuthenticated
	}

	gh := c.sdkClient(token)
	sdkIssues, _, err := gh.Issues.ListByRepo(ctx, owner, repo,
		&gogithub.IssueListByRepoOptions{
			State: "open",
			ListOptions: gogithub.ListOptions{
				PerPage: 100,
			},
		})
	if err != nil {
		return nil, mapError(err)
	}

	issues := make([]domain.GitHubIssue, len(sdkIssues))
	for i, a := range sdkIssues {
		issues[i] = issueToDomain(a)
	}
	return issues, nil
}

// RepoURL returns the canonical web URL for a repository.
func (c *Client) RepoURL(owner, repo string) string {
	return fmt.Sprintf("https://github.com/%s/%s", owner, repo)
}

// IssueURL returns the canonical web URL for an issue or PR.
func (c *Client) IssueURL(owner, repo string, number int) string {
	return fmt.Sprintf("https://github.com/%s/%s/issues/%d", owner, repo, number)
}

// ---------------------------------------------------------------------------
// SDK → domain conversions
// ---------------------------------------------------------------------------

// repoToDomain converts a go-github Repository to a domain.GitHubRepo.
func repoToDomain(r *gogithub.Repository) domain.GitHubRepo {
	var updatedAt time.Time
	if ts := r.GetUpdatedAt(); !ts.Time.IsZero() {
		updatedAt = ts.Time
	}

	return domain.GitHubRepo{
		Owner:         r.GetOwner().GetLogin(),
		Name:          r.GetName(),
		FullName:      r.GetFullName(),
		Description:   r.GetDescription(),
		HTMLURL:       r.GetHTMLURL(),
		CloneURL:      r.GetCloneURL(),
		SSHURL:        r.GetSSHURL(),
		DefaultBranch: r.GetDefaultBranch(),
		Private:       r.GetPrivate(),
		Fork:          r.GetFork(),
		Archived:      r.GetArchived(),
		UpdatedAt:     updatedAt,
	}
}

// issueToDomain converts a go-github Issue to a domain.GitHubIssue.
func issueToDomain(a *gogithub.Issue) domain.GitHubIssue {
	labels := make([]string, len(a.Labels))
	for i, l := range a.Labels {
		labels[i] = l.GetName()
	}

	var createdAt, updatedAt time.Time
	if ts := a.GetCreatedAt(); !ts.Time.IsZero() {
		createdAt = ts.Time
	}
	if ts := a.GetUpdatedAt(); !ts.Time.IsZero() {
		updatedAt = ts.Time
	}

	return domain.GitHubIssue{
		Number:        a.GetNumber(),
		Title:         a.GetTitle(),
		State:         a.GetState(),
		HTMLURL:       a.GetHTMLURL(),
		User:          a.GetUser().GetLogin(),
		Labels:        labels,
		IsPullRequest: a.PullRequestLinks != nil,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}
}
