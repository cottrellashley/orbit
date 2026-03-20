package jsonstore_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cottrellashley/orbit/internal/adapter/jsonstore"
	"github.com/cottrellashley/orbit/internal/domain"
	"github.com/cottrellashley/orbit/internal/port"
)

// Verify ProjectStore satisfies port.ProjectRepository at compile time.
var _ port.ProjectRepository = (*jsonstore.ProjectStore)(nil)

// ---------------------------------------------------------------------------
// Basic CRUD operations
// ---------------------------------------------------------------------------

func TestProjectStore_ListEmpty(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.NewProjectStore(dir)

	projects, err := s.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("List() returned %d projects, want 0", len(projects))
	}
}

func TestProjectStore_SaveAndList(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.NewProjectStore(dir)

	now := time.Date(2026, 3, 9, 12, 0, 0, 0, time.UTC)
	projects := []*domain.Project{
		{
			Name:         "myproject",
			Path:         "/home/user/code/myproject",
			Description:  "A test project",
			ProfileName:  "default",
			Topology:     domain.TopologySingleRepo,
			Integrations: []domain.IntegrationTag{domain.TagGit, domain.TagPython},
			Repos: []domain.RepoInfo{
				{Path: "/home/user/code/myproject", RemoteURL: "git@github.com:user/myproject.git", CurrentBranch: "main"},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	if err := s.Save(projects); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := s.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("List() returned %d projects, want 1", len(loaded))
	}

	p := loaded[0]
	if p.Name != "myproject" {
		t.Errorf("Name = %q, want %q", p.Name, "myproject")
	}
	if p.Path != "/home/user/code/myproject" {
		t.Errorf("Path = %q, want %q", p.Path, "/home/user/code/myproject")
	}
	if p.Topology != domain.TopologySingleRepo {
		t.Errorf("Topology = %v, want TopologySingleRepo", p.Topology)
	}
	if len(p.Integrations) != 2 {
		t.Errorf("len(Integrations) = %d, want 2", len(p.Integrations))
	}
	if len(p.Repos) != 1 {
		t.Errorf("len(Repos) = %d, want 1", len(p.Repos))
	}
	if p.Repos[0].CurrentBranch != "main" {
		t.Errorf("Repos[0].CurrentBranch = %q, want %q", p.Repos[0].CurrentBranch, "main")
	}
}

func TestProjectStore_Get(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.NewProjectStore(dir)

	now := time.Now().Truncate(time.Second)
	projects := []*domain.Project{
		{Name: "alpha", Path: "/a", CreatedAt: now, UpdatedAt: now},
		{Name: "beta", Path: "/b", CreatedAt: now, UpdatedAt: now},
	}
	if err := s.Save(projects); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	p, err := s.Get("beta")
	if err != nil {
		t.Fatalf("Get(beta) error: %v", err)
	}
	if p.Name != "beta" {
		t.Errorf("Get(beta).Name = %q, want %q", p.Name, "beta")
	}

	_, err = s.Get("nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get(nonexistent) error = %v, want ErrNotFound", err)
	}
}

func TestProjectStore_GetByPath(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.NewProjectStore(dir)

	now := time.Now().Truncate(time.Second)
	projects := []*domain.Project{
		{Name: "alpha", Path: "/a", CreatedAt: now, UpdatedAt: now},
	}
	if err := s.Save(projects); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	p, err := s.GetByPath("/a")
	if err != nil {
		t.Fatalf("GetByPath(/a) error: %v", err)
	}
	if p.Name != "alpha" {
		t.Errorf("Name = %q, want %q", p.Name, "alpha")
	}

	_, err = s.GetByPath("/nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByPath(nonexistent) error = %v, want ErrNotFound", err)
	}
}

func TestProjectStore_Delete(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.NewProjectStore(dir)

	now := time.Now().Truncate(time.Second)
	projects := []*domain.Project{
		{Name: "alpha", Path: "/a", CreatedAt: now, UpdatedAt: now},
		{Name: "beta", Path: "/b", CreatedAt: now, UpdatedAt: now},
	}
	if err := s.Save(projects); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	if err := s.Delete("alpha"); err != nil {
		t.Fatalf("Delete(alpha) error: %v", err)
	}

	remaining, err := s.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(remaining) != 1 {
		t.Fatalf("List() returned %d projects, want 1", len(remaining))
	}
	if remaining[0].Name != "beta" {
		t.Errorf("remaining[0].Name = %q, want %q", remaining[0].Name, "beta")
	}

	err = s.Delete("nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Delete(nonexistent) error = %v, want ErrNotFound", err)
	}
}

// ---------------------------------------------------------------------------
// Migration from environments.json
// ---------------------------------------------------------------------------

func TestProjectStore_MigrateFromEnvironments(t *testing.T) {
	dir := t.TempDir()

	// Write a legacy environments.json (same format as jsonstore.Store).
	now := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	envData := []map[string]any{
		{
			"name":         "legacy-env",
			"path":         "/home/user/legacy",
			"profile_name": "starter",
			"description":  "A legacy environment",
			"created_at":   now.Format(time.RFC3339),
			"updated_at":   now.Format(time.RFC3339),
		},
		{
			"name":        "another-env",
			"path":        "/home/user/another",
			"description": "Another environment",
			"created_at":  now.Format(time.RFC3339),
			"updated_at":  now.Format(time.RFC3339),
		},
	}
	b, _ := json.MarshalIndent(envData, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, "environments.json"), b, 0644); err != nil {
		t.Fatalf("write environments.json: %v", err)
	}

	// ProjectStore should auto-migrate when projects.json doesn't exist.
	s := jsonstore.NewProjectStore(dir)
	projects, err := s.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("List() returned %d projects, want 2", len(projects))
	}

	p := projects[0]
	if p.Name != "legacy-env" {
		t.Errorf("projects[0].Name = %q, want %q", p.Name, "legacy-env")
	}
	if p.ProfileName != "starter" {
		t.Errorf("projects[0].ProfileName = %q, want %q", p.ProfileName, "starter")
	}
	if p.Topology != domain.TopologyUnknown {
		t.Errorf("Topology = %v, want TopologyUnknown", p.Topology)
	}
	if len(p.Integrations) != 0 {
		t.Errorf("len(Integrations) = %d, want 0", len(p.Integrations))
	}

	// Verify environments.json is untouched.
	origData, err := os.ReadFile(filepath.Join(dir, "environments.json"))
	if err != nil {
		t.Fatalf("read environments.json after migration: %v", err)
	}
	if string(origData) != string(b) {
		t.Error("environments.json was modified during migration")
	}
}

func TestProjectStore_MigrateDoesNotOverwrite(t *testing.T) {
	dir := t.TempDir()

	// Write both files — projects.json should take precedence.
	envData := []map[string]any{
		{"name": "env-only", "path": "/env", "description": "", "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"},
	}
	b, _ := json.MarshalIndent(envData, "", "  ")
	os.WriteFile(filepath.Join(dir, "environments.json"), b, 0644)

	projData := map[string]any{
		"version": 1,
		"projects": []map[string]any{
			{"name": "proj-only", "path": "/proj", "description": "", "topology": "unknown", "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"},
		},
	}
	b2, _ := json.MarshalIndent(projData, "", "  ")
	os.WriteFile(filepath.Join(dir, "projects.json"), b2, 0644)

	s := jsonstore.NewProjectStore(dir)
	projects, err := s.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("List() returned %d projects, want 1", len(projects))
	}
	if projects[0].Name != "proj-only" {
		t.Errorf("Name = %q, want %q (projects.json should take precedence)", projects[0].Name, "proj-only")
	}
}

// ---------------------------------------------------------------------------
// Versioned format
// ---------------------------------------------------------------------------

func TestProjectStore_VersionedFormat(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.NewProjectStore(dir)

	now := time.Now().Truncate(time.Second)
	if err := s.Save([]*domain.Project{
		{Name: "test", Path: "/test", CreatedAt: now, UpdatedAt: now},
	}); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Read raw JSON and verify version field exists.
	raw, err := os.ReadFile(filepath.Join(dir, "projects.json"))
	if err != nil {
		t.Fatalf("read projects.json: %v", err)
	}

	var envelope struct {
		Version int `json:"version"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("parse projects.json: %v", err)
	}
	if envelope.Version != 1 {
		t.Errorf("version = %d, want 1", envelope.Version)
	}
}

func TestProjectStore_RejectsFutureVersion(t *testing.T) {
	dir := t.TempDir()

	// Write a projects.json with version 999.
	data := `{"version": 999, "projects": []}`
	os.WriteFile(filepath.Join(dir, "projects.json"), []byte(data), 0644)

	s := jsonstore.NewProjectStore(dir)
	_, err := s.List()
	if err == nil {
		t.Fatal("List() should error on future version")
	}
}

// ---------------------------------------------------------------------------
// Atomic write safety
// ---------------------------------------------------------------------------

func TestProjectStore_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.NewProjectStore(dir)

	now := time.Now().Truncate(time.Second)
	projects := []*domain.Project{
		{Name: "first", Path: "/first", CreatedAt: now, UpdatedAt: now},
	}
	if err := s.Save(projects); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify no leftover temp files.
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("leftover temp file: %s", e.Name())
		}
	}

	// Overwrite with new data — old data should be cleanly replaced.
	projects = []*domain.Project{
		{Name: "second", Path: "/second", CreatedAt: now, UpdatedAt: now},
	}
	if err := s.Save(projects); err != nil {
		t.Fatalf("Save() overwrite error: %v", err)
	}

	loaded, err := s.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(loaded) != 1 || loaded[0].Name != "second" {
		t.Errorf("after overwrite: got %v, want [second]", loaded)
	}
}

// ---------------------------------------------------------------------------
// Topology round-trip
// ---------------------------------------------------------------------------

func TestProjectStore_TopologyRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.NewProjectStore(dir)

	now := time.Now().Truncate(time.Second)
	tests := []struct {
		topology domain.ProjectTopology
	}{
		{domain.TopologyUnknown},
		{domain.TopologySingleRepo},
		{domain.TopologyMultiRepo},
	}

	for _, tt := range tests {
		t.Run(tt.topology.String(), func(t *testing.T) {
			if err := s.Save([]*domain.Project{
				{Name: "rt", Path: "/rt", Topology: tt.topology, CreatedAt: now, UpdatedAt: now},
			}); err != nil {
				t.Fatalf("Save() error: %v", err)
			}

			loaded, err := s.List()
			if err != nil {
				t.Fatalf("List() error: %v", err)
			}
			if loaded[0].Topology != tt.topology {
				t.Errorf("Topology = %v, want %v", loaded[0].Topology, tt.topology)
			}
		})
	}
}
