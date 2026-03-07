package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cottrellashley/orbit/internal/adapter"
	"github.com/cottrellashley/orbit/internal/role"
)

func TestLoadAndSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &Config{
		ArchivePath: "/tmp/archive",
		Adapters: []*adapter.Adapter{
			{Name: "opencode", Command: "opencode", Default: true},
		},
		Roles: []*role.Role{
			{Name: "exec", Type: role.Environment, Path: "/tmp/exec", Tags: []string{"life"}},
			{Name: "eng", Type: role.Workspace, Path: "/tmp/eng", Tags: []string{"code"}},
		},
	}

	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.ArchivePath != "/tmp/archive" {
		t.Errorf("expected archive_path %q, got %q", "/tmp/archive", loaded.ArchivePath)
	}

	if len(loaded.Adapters) != 1 {
		t.Fatalf("expected 1 adapter, got %d", len(loaded.Adapters))
	}

	if len(loaded.Roles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(loaded.Roles))
	}

	if loaded.Roles[0].Name != "exec" {
		t.Errorf("expected first role name %q, got %q", "exec", loaded.Roles[0].Name)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error loading missing file")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte("{{invalid yaml"), 0644)

	_, err := Load(path)
	if err == nil {
		t.Error("expected error loading invalid YAML")
	}
}

func TestFindRole(t *testing.T) {
	cfg := &Config{
		Roles: []*role.Role{
			{Name: "exec", Type: role.Environment, Path: "/tmp/exec"},
			{Name: "eng", Type: role.Workspace, Path: "/tmp/eng"},
		},
	}

	r, err := cfg.FindRole("eng")
	if err != nil {
		t.Fatalf("FindRole failed: %v", err)
	}
	if r.Name != "eng" {
		t.Errorf("expected %q, got %q", "eng", r.Name)
	}

	_, err = cfg.FindRole("missing")
	if err == nil {
		t.Error("expected error for missing role")
	}
}

func TestAddRole(t *testing.T) {
	cfg := &Config{}

	r := &role.Role{Name: "test", Type: role.Environment, Path: "/tmp/test"}
	if err := cfg.AddRole(r); err != nil {
		t.Fatalf("AddRole failed: %v", err)
	}

	if len(cfg.Roles) != 1 {
		t.Fatalf("expected 1 role, got %d", len(cfg.Roles))
	}

	// Adding duplicate should fail
	err := cfg.AddRole(r)
	if err == nil {
		t.Error("expected error adding duplicate role")
	}
}

func TestRemoveRole(t *testing.T) {
	cfg := &Config{
		Roles: []*role.Role{
			{Name: "a", Type: role.Environment, Path: "/tmp/a"},
			{Name: "b", Type: role.Workspace, Path: "/tmp/b"},
		},
	}

	removed, err := cfg.RemoveRole("a")
	if err != nil {
		t.Fatalf("RemoveRole failed: %v", err)
	}
	if removed.Name != "a" {
		t.Errorf("expected removed role %q, got %q", "a", removed.Name)
	}
	if len(cfg.Roles) != 1 {
		t.Errorf("expected 1 role remaining, got %d", len(cfg.Roles))
	}

	_, err = cfg.RemoveRole("missing")
	if err == nil {
		t.Error("expected error removing missing role")
	}
}

func TestResolveAdapter(t *testing.T) {
	cfg := &Config{
		Adapters: []*adapter.Adapter{
			{Name: "opencode", Command: "opencode", Default: true},
			{Name: "cursor", Command: "cursor"},
		},
		Roles: []*role.Role{
			{Name: "exec", Type: role.Environment, Path: "/tmp/exec"},
			{Name: "special", Type: role.Environment, Path: "/tmp/special", Adapter: "cursor"},
		},
	}

	// Role without explicit adapter gets default
	a, err := cfg.ResolveAdapter(cfg.Roles[0])
	if err != nil {
		t.Fatalf("ResolveAdapter failed: %v", err)
	}
	if a.Name != "opencode" {
		t.Errorf("expected %q, got %q", "opencode", a.Name)
	}

	// Role with explicit adapter gets that one
	a, err = cfg.ResolveAdapter(cfg.Roles[1])
	if err != nil {
		t.Fatalf("ResolveAdapter failed: %v", err)
	}
	if a.Name != "cursor" {
		t.Errorf("expected %q, got %q", "cursor", a.Name)
	}
}

func TestExpandAndUnexpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}

	expanded := expandHome("~/test/path")
	expected := filepath.Join(home, "test/path")
	if expanded != expected {
		t.Errorf("expandHome(~/test/path) = %q, expected %q", expanded, expected)
	}

	unexpanded := unexpandHome(expected)
	if unexpanded != "~/test/path" {
		t.Errorf("unexpandHome(%q) = %q, expected %q", expected, unexpanded, "~/test/path")
	}

	// Non-home paths should pass through
	if expandHome("/absolute/path") != "/absolute/path" {
		t.Error("expandHome should not modify absolute paths")
	}
	if unexpandHome("/other/path") != "/other/path" {
		t.Error("unexpandHome should not modify non-home paths")
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "deep", "config.yaml")

	cfg := &Config{ArchivePath: "/tmp/archive"}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save failed to create nested directory: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected config file to exist after save")
	}
}

func TestSavePreservesTildePaths(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &Config{
		ArchivePath: filepath.Join(home, "Archive"),
		Roles: []*role.Role{
			{Name: "test", Type: role.Environment, Path: filepath.Join(home, "Test")},
		},
	}

	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Read raw file and verify ~ is used
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read saved file: %v", err)
	}

	content := string(data)
	if !contains(content, "~/Archive") {
		t.Errorf("expected ~/Archive in saved config, got:\n%s", content)
	}
	if !contains(content, "~/Test") {
		t.Errorf("expected ~/Test in saved config, got:\n%s", content)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
