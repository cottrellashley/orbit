package app

import (
	"context"
	"testing"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
)

func TestGitHubService_AuthStatus(t *testing.T) {
	gh := &mockGitHubProvider{
		status: &domain.GitHubAuthStatus{
			Authenticated: true,
			User:          "testuser",
			TokenSource:   "GITHUB_TOKEN",
			Scopes:        []string{"repo"},
		},
	}
	svc := NewGitHubService(gh)

	status, err := svc.AuthStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Authenticated {
		t.Fatal("expected authenticated")
	}
	if status.User != "testuser" {
		t.Fatalf("expected user 'testuser', got %q", status.User)
	}
}

func TestGitHubService_AuthStatus_Unauthenticated(t *testing.T) {
	gh := &mockGitHubProvider{}
	svc := NewGitHubService(gh)

	status, err := svc.AuthStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Authenticated {
		t.Fatal("expected not authenticated")
	}
}

func TestGitHubService_ListRepos(t *testing.T) {
	now := time.Now().Add(-2 * time.Second)
	gh := &mockGitHubProvider{
		repos: []domain.GitHubRepo{
			{
				Owner:         "octocat",
				Name:          "hello-world",
				FullName:      "octocat/hello-world",
				HTMLURL:       "https://github.com/octocat/hello-world",
				DefaultBranch: "main",
				UpdatedAt:     now,
			},
		},
	}
	svc := NewGitHubService(gh)

	repos, err := svc.ListRepos(context.Background())
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

func TestGitHubService_ListIssues(t *testing.T) {
	now := time.Now().Add(-2 * time.Second)
	gh := &mockGitHubProvider{
		issues: []domain.GitHubIssue{
			{Number: 1, Title: "Bug", State: "open", CreatedAt: now, UpdatedAt: now},
			{Number: 2, Title: "Feature", State: "open", CreatedAt: now, UpdatedAt: now},
		},
	}
	svc := NewGitHubService(gh)

	issues, err := svc.ListIssues(context.Background(), "octocat", "hello-world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 2 {
		t.Fatalf("expected 2 issues, got %d", len(issues))
	}
}

func TestGitHubService_RepoLink(t *testing.T) {
	gh := &mockGitHubProvider{}
	svc := NewGitHubService(gh)

	link := svc.RepoLink("octocat", "hello-world")
	if link.Label != "octocat/hello-world" {
		t.Fatalf("expected label 'octocat/hello-world', got %q", link.Label)
	}
	if link.URL != "https://github.com/octocat/hello-world" {
		t.Fatalf("expected URL 'https://github.com/octocat/hello-world', got %q", link.URL)
	}
	if link.Kind != domain.LinkRepo {
		t.Fatalf("expected kind LinkRepo, got %q", link.Kind)
	}
}

func TestGitHubService_IssueLink(t *testing.T) {
	gh := &mockGitHubProvider{}
	svc := NewGitHubService(gh)

	link := svc.IssueLink("octocat", "hello-world", 42)
	if link.Label != "octocat/hello-world#42" {
		t.Fatalf("expected label 'octocat/hello-world#42', got %q", link.Label)
	}
	if link.URL != "https://github.com/octocat/hello-world/issues/42" {
		t.Fatalf("expected issue URL, got %q", link.URL)
	}
	if link.Kind != domain.LinkIssue {
		t.Fatalf("expected kind LinkIssue, got %q", link.Kind)
	}
}

func TestGitHubService_CapabilitySummary_Authenticated(t *testing.T) {
	gh := &mockGitHubProvider{
		status: &domain.GitHubAuthStatus{
			Authenticated: true,
			User:          "testuser",
			TokenSource:   "GH_TOKEN",
		},
	}
	svc := NewGitHubService(gh)

	summary := svc.CapabilitySummary(context.Background())
	if summary != "GitHub: authenticated as testuser (via GH_TOKEN)" {
		t.Fatalf("unexpected summary: %q", summary)
	}
}

func TestGitHubService_CapabilitySummary_Unauthenticated(t *testing.T) {
	gh := &mockGitHubProvider{}
	svc := NewGitHubService(gh)

	summary := svc.CapabilitySummary(context.Background())
	if summary != "GitHub: not authenticated" {
		t.Fatalf("unexpected summary: %q", summary)
	}
}
