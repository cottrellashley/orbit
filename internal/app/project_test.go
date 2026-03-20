package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
)

func TestProjectService_List(t *testing.T) {
	repo := newMockProjectRepo(
		&domain.Project{Name: "alpha", Path: "/a"},
		&domain.Project{Name: "beta", Path: "/b"},
	)
	svc := NewProjectService(repo, nil)

	list, err := svc.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(list))
	}
}

func TestProjectService_Get(t *testing.T) {
	repo := newMockProjectRepo(
		&domain.Project{Name: "alpha", Path: "/a"},
	)
	svc := NewProjectService(repo, nil)

	p, err := svc.Get("alpha")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if p.Name != "alpha" {
		t.Fatalf("expected name 'alpha', got %q", p.Name)
	}

	_, err = svc.Get("missing")
	if err == nil {
		t.Fatal("expected error for missing project")
	}
}

func TestProjectService_GetByPath(t *testing.T) {
	repo := newMockProjectRepo(
		&domain.Project{Name: "alpha", Path: "/tmp/alpha"},
	)
	svc := NewProjectService(repo, nil)

	p, err := svc.GetByPath("/tmp/alpha")
	if err != nil {
		t.Fatalf("GetByPath() error: %v", err)
	}
	if p.Name != "alpha" {
		t.Fatalf("expected name 'alpha', got %q", p.Name)
	}
}

func TestProjectService_Register(t *testing.T) {
	dir := t.TempDir()
	repo := newMockProjectRepo()
	svc := NewProjectService(repo, nil)

	p, err := svc.Register("myproject", dir, "test project")
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}
	if p.Name != "myproject" {
		t.Fatalf("expected name 'myproject', got %q", p.Name)
	}
	abs, _ := filepath.Abs(dir)
	if p.Path != abs {
		t.Fatalf("expected path %q, got %q", abs, p.Path)
	}
	if p.Topology != domain.TopologyUnknown {
		t.Fatalf("expected TopologyUnknown, got %v", p.Topology)
	}
	if p.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}

	// Verify persisted.
	list, _ := svc.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 project in repo, got %d", len(list))
	}
}

func TestProjectService_Register_DuplicateName(t *testing.T) {
	dir := t.TempDir()
	repo := newMockProjectRepo(
		&domain.Project{Name: "dup", Path: "/existing"},
	)
	svc := NewProjectService(repo, nil)

	_, err := svc.Register("dup", dir, "duplicate")
	if err == nil {
		t.Fatal("expected error for duplicate name")
	}
}

func TestProjectService_Register_InvalidPath(t *testing.T) {
	repo := newMockProjectRepo()
	svc := NewProjectService(repo, nil)

	_, err := svc.Register("bad", "/nonexistent/path/xyz", "")
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func TestProjectService_Delete(t *testing.T) {
	repo := newMockProjectRepo(
		&domain.Project{Name: "alpha", Path: "/a"},
	)
	svc := NewProjectService(repo, nil)

	if err := svc.Delete("alpha"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	list, _ := svc.List()
	if len(list) != 0 {
		t.Fatalf("expected 0 projects, got %d", len(list))
	}
}

func TestProjectService_Delete_NotFound(t *testing.T) {
	repo := newMockProjectRepo()
	svc := NewProjectService(repo, nil)

	err := svc.Delete("ghost")
	if err == nil {
		t.Fatal("expected error for missing project")
	}
}

func TestProjectService_Update(t *testing.T) {
	past := time.Now().Add(-2 * time.Second).Truncate(time.Second)
	repo := newMockProjectRepo(
		&domain.Project{
			Name:      "alpha",
			Path:      "/a",
			Topology:  domain.TopologyUnknown,
			UpdatedAt: past,
		},
	)
	svc := NewProjectService(repo, nil)

	tags := []domain.IntegrationTag{domain.TagGit, domain.TagPython}
	repos := []domain.RepoInfo{{Path: "/a", RemoteURL: "git@github.com:x/y.git", CurrentBranch: "main"}}

	p, err := svc.Update("alpha", domain.TopologySingleRepo, tags, repos, "updated desc")
	if err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	if p.Topology != domain.TopologySingleRepo {
		t.Fatalf("expected TopologySingleRepo, got %v", p.Topology)
	}
	if len(p.Integrations) != 2 {
		t.Fatalf("expected 2 integrations, got %d", len(p.Integrations))
	}
	if len(p.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(p.Repos))
	}
	if p.Description != "updated desc" {
		t.Fatalf("expected description 'updated desc', got %q", p.Description)
	}
	if !p.UpdatedAt.After(past) {
		t.Fatalf("expected UpdatedAt (%v) to be after past (%v)", p.UpdatedAt, past)
	}
}

func TestProjectService_Update_NotFound(t *testing.T) {
	repo := newMockProjectRepo()
	svc := NewProjectService(repo, nil)

	_, err := svc.Update("ghost", domain.TopologyUnknown, nil, nil, "")
	if err == nil {
		t.Fatal("expected error for missing project")
	}
}

func TestProjectService_CreateFromProfile(t *testing.T) {
	dir := t.TempDir()
	targetPath := filepath.Join(dir, "newproj")

	profiles := newMockProfileRepo(dir)
	profiles.profiles = append(profiles.profiles, &domain.Profile{Name: "starter", Description: "test"})

	repo := newMockProjectRepo()
	svc := NewProjectService(repo, profiles)

	p, err := svc.CreateFromProfile("proj1", "starter", targetPath, "from profile", true)
	if err != nil {
		t.Fatalf("CreateFromProfile() error: %v", err)
	}
	if p.ProfileName != "starter" {
		t.Fatalf("expected profile 'starter', got %q", p.ProfileName)
	}

	// Verify dir was created.
	info, err := os.Stat(targetPath)
	if err != nil {
		t.Fatalf("expected directory to be created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected a directory")
	}
}

func TestProjectService_CreateFromProfile_NilProfiles(t *testing.T) {
	repo := newMockProjectRepo()
	svc := NewProjectService(repo, nil)

	_, err := svc.CreateFromProfile("p", "starter", "/tmp", "", false)
	if err == nil {
		t.Fatal("expected error when profiles repo is nil")
	}
}

func TestProjectService_ListAsEnvironments(t *testing.T) {
	repo := newMockProjectRepo(
		&domain.Project{Name: "alpha", Path: "/a", Description: "desc"},
	)
	svc := NewProjectService(repo, nil)

	envs, err := svc.ListAsEnvironments()
	if err != nil {
		t.Fatalf("ListAsEnvironments() error: %v", err)
	}
	if len(envs) != 1 {
		t.Fatalf("expected 1 env, got %d", len(envs))
	}
	if envs[0].Name != "alpha" {
		t.Fatalf("expected name 'alpha', got %q", envs[0].Name)
	}
}

func TestProjectService_GetAsEnvironment(t *testing.T) {
	repo := newMockProjectRepo(
		&domain.Project{Name: "alpha", Path: "/a"},
	)
	svc := NewProjectService(repo, nil)

	env, err := svc.GetAsEnvironment("alpha")
	if err != nil {
		t.Fatalf("GetAsEnvironment() error: %v", err)
	}
	if env.Name != "alpha" {
		t.Fatalf("expected 'alpha', got %q", env.Name)
	}
}
