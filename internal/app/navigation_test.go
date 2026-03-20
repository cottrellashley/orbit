package app

import (
	"testing"

	"github.com/cottrellashley/orbit/internal/domain"
)

func TestNavigationService_RepoWebURL_SSH(t *testing.T) {
	nav := NewNavigationService()
	repo := domain.RepoInfo{
		Path:      "/home/user/myrepo",
		RemoteURL: "git@github.com:octocat/hello-world.git",
	}
	link := nav.RepoWebURL(repo)
	if link == nil {
		t.Fatal("expected non-nil link")
	}
	if link.URL != "https://github.com/octocat/hello-world" {
		t.Fatalf("expected URL 'https://github.com/octocat/hello-world', got %q", link.URL)
	}
	if link.Kind != domain.LinkRepo {
		t.Fatalf("expected kind LinkRepo, got %q", link.Kind)
	}
}

func TestNavigationService_RepoWebURL_HTTPS(t *testing.T) {
	nav := NewNavigationService()
	repo := domain.RepoInfo{
		Path:      "/home/user/myrepo",
		RemoteURL: "https://github.com/octocat/hello-world.git",
	}
	link := nav.RepoWebURL(repo)
	if link == nil {
		t.Fatal("expected non-nil link")
	}
	if link.URL != "https://github.com/octocat/hello-world" {
		t.Fatalf("expected URL without .git, got %q", link.URL)
	}
}

func TestNavigationService_RepoWebURL_Empty(t *testing.T) {
	nav := NewNavigationService()
	link := nav.RepoWebURL(domain.RepoInfo{})
	if link != nil {
		t.Fatal("expected nil for empty remote URL")
	}
}

func TestNavigationService_BranchURL(t *testing.T) {
	nav := NewNavigationService()
	repo := domain.RepoInfo{
		RemoteURL:     "git@github.com:octocat/hello-world.git",
		CurrentBranch: "feature/cool",
	}
	link := nav.BranchURL(repo)
	if link == nil {
		t.Fatal("expected non-nil link")
	}
	if link.URL != "https://github.com/octocat/hello-world/tree/feature/cool" {
		t.Fatalf("expected branch URL, got %q", link.URL)
	}
	if link.Kind != domain.LinkBranch {
		t.Fatalf("expected kind LinkBranch, got %q", link.Kind)
	}
}

func TestNavigationService_BranchURL_NoBranch(t *testing.T) {
	nav := NewNavigationService()
	repo := domain.RepoInfo{
		RemoteURL: "git@github.com:octocat/hello-world.git",
	}
	link := nav.BranchURL(repo)
	if link != nil {
		t.Fatal("expected nil when no branch")
	}
}

func TestNavigationService_IssueListURL(t *testing.T) {
	nav := NewNavigationService()
	repo := domain.RepoInfo{
		RemoteURL: "git@github.com:octocat/hello-world.git",
	}
	link := nav.IssueListURL(repo)
	if link == nil {
		t.Fatal("expected non-nil link")
	}
	if link.URL != "https://github.com/octocat/hello-world/issues" {
		t.Fatalf("expected issues URL, got %q", link.URL)
	}
}

func TestNavigationService_PRListURL(t *testing.T) {
	nav := NewNavigationService()
	repo := domain.RepoInfo{
		RemoteURL: "git@github.com:octocat/hello-world.git",
	}
	link := nav.PRListURL(repo)
	if link == nil {
		t.Fatal("expected non-nil link")
	}
	if link.URL != "https://github.com/octocat/hello-world/pulls" {
		t.Fatalf("expected pulls URL, got %q", link.URL)
	}
	if link.Kind != domain.LinkPR {
		t.Fatalf("expected kind LinkPR, got %q", link.Kind)
	}
}

func TestNavigationService_SessionWebURL(t *testing.T) {
	nav := NewNavigationService()
	sess := domain.Session{ID: "abc123", Title: "My Session"}
	link := nav.SessionWebURL(sess, "http://127.0.0.1:3000")
	if link == nil {
		t.Fatal("expected non-nil link")
	}
	if link.URL != "http://127.0.0.1:3000/#/sessions/abc123" {
		t.Fatalf("expected session URL, got %q", link.URL)
	}
	if link.Kind != domain.LinkSession {
		t.Fatalf("expected kind LinkSession, got %q", link.Kind)
	}
}

func TestNavigationService_ProjectLinks(t *testing.T) {
	nav := NewNavigationService()
	proj := domain.Project{
		Name: "test",
		Repos: []domain.RepoInfo{
			{
				Path:          "/home/user/test",
				RemoteURL:     "git@github.com:octocat/hello-world.git",
				CurrentBranch: "main",
			},
		},
	}
	links := nav.ProjectLinks(proj)
	// Should have 4 links: repo, branch, issues, PRs
	if len(links) != 4 {
		t.Fatalf("expected 4 links, got %d", len(links))
	}
	kinds := make(map[domain.LinkKind]bool)
	for _, l := range links {
		kinds[l.Kind] = true
	}
	if !kinds[domain.LinkRepo] {
		t.Error("missing repo link")
	}
	if !kinds[domain.LinkBranch] {
		t.Error("missing branch link")
	}
	if !kinds[domain.LinkIssue] {
		t.Error("missing issue link")
	}
	if !kinds[domain.LinkPR] {
		t.Error("missing PR link")
	}
}

func TestNavigationService_ProjectLinks_NoRepos(t *testing.T) {
	nav := NewNavigationService()
	proj := domain.Project{Name: "empty"}
	links := nav.ProjectLinks(proj)
	if len(links) != 0 {
		t.Fatalf("expected 0 links, got %d", len(links))
	}
}

// Test the remoteToWebURL helper directly via RepoWebURL
func TestNavigationService_RepoWebURL_VariousFormats(t *testing.T) {
	nav := NewNavigationService()

	tests := []struct {
		name     string
		remote   string
		expected string
	}{
		{"SSH with .git", "git@github.com:octocat/hello.git", "https://github.com/octocat/hello"},
		{"SSH without .git", "git@github.com:octocat/hello", "https://github.com/octocat/hello"},
		{"HTTPS with .git", "https://github.com/octocat/hello.git", "https://github.com/octocat/hello"},
		{"HTTPS without .git", "https://github.com/octocat/hello", "https://github.com/octocat/hello"},
		{"HTTP with .git", "http://github.com/octocat/hello.git", "https://github.com/octocat/hello"},
		{"GitLab SSH", "git@gitlab.com:group/project.git", "https://gitlab.com/group/project"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			link := nav.RepoWebURL(domain.RepoInfo{RemoteURL: tt.remote})
			if link == nil {
				t.Fatal("expected non-nil link")
			}
			if link.URL != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, link.URL)
			}
		})
	}
}
