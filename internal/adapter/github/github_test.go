package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/cottrellashley/orbit/internal/domain"
)

func TestRepoURL(t *testing.T) {
	c := NewClient("")
	got := c.RepoURL("octocat", "hello")
	want := "https://github.com/octocat/hello"
	if got != want {
		t.Fatalf("RepoURL() = %q, want %q", got, want)
	}
}

func TestIssueURL(t *testing.T) {
	c := NewClient("")
	got := c.IssueURL("octocat", "hello", 42)
	want := "https://github.com/octocat/hello/issues/42"
	if got != want {
		t.Fatalf("IssueURL() = %q, want %q", got, want)
	}
}

func TestResolveToken_EnvVar(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "gh_test_token_123")
	t.Setenv("GH_TOKEN", "")
	c := NewClient("")
	tok := c.resolveToken()
	if tok != "gh_test_token_123" {
		t.Fatalf("resolveToken() = %q, want 'gh_test_token_123'", tok)
	}
}

func TestResolveToken_GHToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "gho_secondary_456")
	c := NewClient("")
	tok := c.resolveToken()
	if tok != "gho_secondary_456" {
		t.Fatalf("resolveToken() = %q, want 'gho_secondary_456'", tok)
	}
}

func TestResolveToken_ConfigFile(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")

	dir := t.TempDir()
	tokenFile := filepath.Join(dir, "github-token")
	if err := os.WriteFile(tokenFile, []byte("file_token_789\n"), 0600); err != nil {
		t.Fatal(err)
	}

	c := NewClient(dir, WithCmdRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return nil, os.ErrNotExist // gh not available
	}))

	tok := c.resolveToken()
	if tok != "file_token_789" {
		t.Fatalf("resolveToken() = %q, want 'file_token_789'", tok)
	}
}

func TestResolveToken_GHCLIFallback(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")

	c := NewClient("", WithCmdRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name == "gh" && len(args) >= 2 && args[0] == "auth" && args[1] == "token" {
			return []byte("cli_token_abc\n"), nil
		}
		return nil, os.ErrNotExist
	}))

	tok := c.resolveToken()
	if tok != "cli_token_abc" {
		t.Fatalf("resolveToken() = %q, want 'cli_token_abc'", tok)
	}
}

func TestResolveToken_NoToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")

	c := NewClient("", WithCmdRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return nil, os.ErrNotExist
	}))

	tok := c.resolveToken()
	if tok != "" {
		t.Fatalf("resolveToken() = %q, want empty", tok)
	}
}

func TestResolveToken_Injected(t *testing.T) {
	c := NewClient("", WithTokenFunc(func() string { return "injected_token" }))
	tok := c.resolveToken()
	if tok != "injected_token" {
		t.Fatalf("resolveToken() = %q, want 'injected_token'", tok)
	}
}

func TestAuthStatus_Unauthenticated(t *testing.T) {
	c := NewClient("", WithTokenFunc(func() string { return "" }))
	status, err := c.AuthStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Authenticated {
		t.Fatal("expected not authenticated")
	}
}

func TestAuthStatus_Authenticated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test_token" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"message": "Bad credentials"})
			return
		}
		w.Header().Set("X-OAuth-Scopes", "repo, read:org")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"login": "testuser"})
	}))
	defer srv.Close()

	c := NewClient("",
		WithBaseURL(srv.URL),
		WithTokenFunc(func() string { return "test_token" }),
	)

	status, err := c.AuthStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Authenticated {
		t.Fatal("expected authenticated")
	}
	if status.User != "testuser" {
		t.Fatalf("expected user 'testuser', got %q", status.User)
	}
	if len(status.Scopes) != 2 {
		t.Fatalf("expected 2 scopes, got %d", len(status.Scopes))
	}
}

func TestAuthStatus_InvalidToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "Bad credentials"})
	}))
	defer srv.Close()

	c := NewClient("",
		WithBaseURL(srv.URL),
		WithTokenFunc(func() string { return "bad_token" }),
	)

	status, err := c.AuthStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Authenticated {
		t.Fatal("expected not authenticated for invalid token")
	}
}

func TestListRepos(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user/repos" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		repos := []map[string]any{
			{
				"full_name":      "octocat/hello-world",
				"name":           "hello-world",
				"html_url":       "https://github.com/octocat/hello-world",
				"clone_url":      "https://github.com/octocat/hello-world.git",
				"ssh_url":        "git@github.com:octocat/hello-world.git",
				"default_branch": "main",
				"private":        false,
				"fork":           false,
				"archived":       false,
				"updated_at":     "2026-03-09T00:00:00Z",
				"owner":          map[string]string{"login": "octocat"},
			},
		}
		json.NewEncoder(w).Encode(repos)
	}))
	defer srv.Close()

	c := NewClient("",
		WithBaseURL(srv.URL),
		WithTokenFunc(func() string { return "token" }),
	)

	repos, err := c.ListRepos(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}
	if repos[0].FullName != "octocat/hello-world" {
		t.Fatalf("expected 'octocat/hello-world', got %q", repos[0].FullName)
	}
}

func TestListRepos_NoAuth(t *testing.T) {
	c := NewClient("", WithTokenFunc(func() string { return "" }))
	_, err := c.ListRepos(context.Background())
	if err != domain.ErrNotAuthenticated {
		t.Fatalf("expected ErrNotAuthenticated, got %v", err)
	}
}

func TestListIssues(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/octocat/hello/issues" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		issues := []map[string]any{
			{
				"number":     42,
				"title":      "Bug report",
				"state":      "open",
				"html_url":   "https://github.com/octocat/hello/issues/42",
				"created_at": "2026-03-09T00:00:00Z",
				"updated_at": "2026-03-09T00:00:00Z",
				"user":       map[string]string{"login": "reporter"},
				"labels":     []map[string]string{{"name": "bug"}},
			},
		}
		json.NewEncoder(w).Encode(issues)
	}))
	defer srv.Close()

	c := NewClient("",
		WithBaseURL(srv.URL),
		WithTokenFunc(func() string { return "token" }),
	)

	issues, err := c.ListIssues(context.Background(), "octocat", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Number != 42 {
		t.Fatalf("expected issue #42, got #%d", issues[0].Number)
	}
	if len(issues[0].Labels) != 1 || issues[0].Labels[0] != "bug" {
		t.Fatalf("expected label 'bug', got %v", issues[0].Labels)
	}
}

func TestListIssues_NoAuth(t *testing.T) {
	c := NewClient("", WithTokenFunc(func() string { return "" }))
	_, err := c.ListIssues(context.Background(), "octocat", "hello")
	if err != domain.ErrNotAuthenticated {
		t.Fatalf("expected ErrNotAuthenticated, got %v", err)
	}
}

func TestRateLimiting(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Limit", "60")
		w.Header().Set("X-RateLimit-Reset", "1893456000") // far future
		w.WriteHeader(http.StatusForbidden)
		// The go-github SDK expects a JSON body with rate limit info.
		json.NewEncoder(w).Encode(map[string]any{
			"message":           "API rate limit exceeded",
			"documentation_url": "https://docs.github.com/rest/overview/resources-in-the-rest-api#rate-limiting",
		})
	}))
	defer srv.Close()

	c := NewClient("",
		WithBaseURL(srv.URL),
		WithTokenFunc(func() string { return "token" }),
	)

	_, err := c.ListRepos(context.Background())
	if err != domain.ErrRateLimited {
		t.Fatalf("expected ErrRateLimited, got %v", err)
	}
}

func TestTokenSource(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "tok")
	t.Setenv("GH_TOKEN", "")
	c := NewClient("")
	src := c.tokenSource()
	if src != "GITHUB_TOKEN" {
		t.Fatalf("tokenSource() = %q, want 'GITHUB_TOKEN'", src)
	}
}
