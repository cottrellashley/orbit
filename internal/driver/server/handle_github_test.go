package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
)

func TestHandleGitHubStatus_Unauthenticated(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	w := doRequest(s, "GET", "/api/github/status", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var status githubAuthStatusJSON
	decodeJSON(t, w, &status)
	if status.Authenticated {
		t.Fatal("expected unauthenticated")
	}
}

func TestHandleGitHubStatus_Authenticated(t *testing.T) {
	ghSvc := &mockGitHubService{
		status: &domain.GitHubAuthStatus{
			Authenticated: true,
			User:          "testuser",
			TokenSource:   "GITHUB_TOKEN",
			Scopes:        []string{"repo", "read:org"},
		},
	}
	srv := New("127.0.0.1:0", &mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{},
		&mockDoctorService{}, &mockOpenService{}, ghSvc, &mockNavigationService{}, &mockMarkdownService{})

	w := doRequest(srv, "GET", "/api/github/status", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var status githubAuthStatusJSON
	decodeJSON(t, w, &status)
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

func TestHandleGitHubRepos(t *testing.T) {
	now := time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC)
	ghSvc := &mockGitHubService{
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
	srv := New("127.0.0.1:0", &mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{},
		&mockDoctorService{}, &mockOpenService{}, ghSvc, &mockNavigationService{}, &mockMarkdownService{})

	w := doRequest(srv, "GET", "/api/github/repos", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var repos []githubRepoJSON
	decodeJSON(t, w, &repos)
	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}
	if repos[0].FullName != "octocat/hello-world" {
		t.Fatalf("expected full_name 'octocat/hello-world', got %q", repos[0].FullName)
	}
}

func TestHandleGitHubIssues(t *testing.T) {
	now := time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC)
	ghSvc := &mockGitHubService{
		issues: []domain.GitHubIssue{
			{
				Number:    42,
				Title:     "Fix the thing",
				State:     "open",
				HTMLURL:   "https://github.com/octocat/hello-world/issues/42",
				User:      "reporter",
				Labels:    []string{"bug"},
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}
	srv := New("127.0.0.1:0", &mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{},
		&mockDoctorService{}, &mockOpenService{}, ghSvc, &mockNavigationService{}, &mockMarkdownService{})

	w := doRequest(srv, "GET", "/api/github/repos/octocat/hello-world/issues", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var issues []githubIssueJSON
	decodeJSON(t, w, &issues)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Number != 42 {
		t.Fatalf("expected issue number 42, got %d", issues[0].Number)
	}
	if issues[0].Title != "Fix the thing" {
		t.Fatalf("expected title 'Fix the thing', got %q", issues[0].Title)
	}
}
