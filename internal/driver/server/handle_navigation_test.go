package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
)

func TestHandleProjectLinks(t *testing.T) {
	projSvc := &mockProjectService{
		projects: []*domain.Project{
			{
				Name:      "myproj",
				Path:      "/home/user/myproj",
				Topology:  domain.TopologySingleRepo,
				Repos:     []domain.RepoInfo{{Path: "/home/user/myproj", RemoteURL: "git@github.com:x/y.git"}},
				CreatedAt: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	s := New("127.0.0.1:0", &mockEnvironmentService{}, projSvc, &mockSessionService{},
		&mockDoctorService{}, &mockOpenService{}, &mockGitHubService{}, &mockNavigationService{}, &mockMarkdownService{})

	w := doRequest(s, "GET", "/api/projects/myproj/links", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var links []jumpLinkJSON
	decodeJSON(t, w, &links)
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].Kind != "repo" {
		t.Fatalf("expected kind 'repo', got %q", links[0].Kind)
	}
}

func TestHandleProjectLinks_NotFound(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	w := doRequest(s, "GET", "/api/projects/ghost/links", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
