package jsonstore_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cottrellashley/orbit/internal/adapter/jsonstore"
	"github.com/cottrellashley/orbit/internal/domain"
)

func newStore(t *testing.T) *jsonstore.Store {
	t.Helper()
	dir := t.TempDir()
	return jsonstore.New(dir)
}

func sampleEnv(name, path string) *domain.Environment {
	now := time.Now().Truncate(time.Second)
	return &domain.Environment{
		Name:        name,
		Path:        path,
		Description: "test environment",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func TestList_Empty(t *testing.T) {
	s := newStore(t)
	envs, err := s.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(envs) != 0 {
		t.Errorf("expected 0 environments, got %d", len(envs))
	}
}

func TestSaveAndList(t *testing.T) {
	s := newStore(t)
	dir := t.TempDir()

	env := sampleEnv("myenv", filepath.Join(dir, "project"))
	if err := s.Save([]*domain.Environment{env}); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	envs, err := s.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(envs) != 1 {
		t.Fatalf("expected 1 environment, got %d", len(envs))
	}
	if envs[0].Name != env.Name {
		t.Errorf("Name: got %q, want %q", envs[0].Name, env.Name)
	}
	if envs[0].Path != env.Path {
		t.Errorf("Path: got %q, want %q", envs[0].Path, env.Path)
	}
}

func TestGet_Found(t *testing.T) {
	s := newStore(t)
	dir := t.TempDir()

	env := sampleEnv("alpha", filepath.Join(dir, "alpha"))
	if err := s.Save([]*domain.Environment{env}); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	got, err := s.Get("alpha")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got.Name != "alpha" {
		t.Errorf("Name: got %q, want %q", got.Name, "alpha")
	}
}

func TestGet_NotFound(t *testing.T) {
	s := newStore(t)

	_, err := s.Get("missing")
	if err == nil {
		t.Fatal("expected error for missing environment, got nil")
	}
	if !isNotFound(err) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestGetByPath_Found(t *testing.T) {
	s := newStore(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "proj")

	env := sampleEnv("proj", path)
	if err := s.Save([]*domain.Environment{env}); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	got, err := s.GetByPath(path)
	if err != nil {
		t.Fatalf("GetByPath() error: %v", err)
	}
	if got.Path != path {
		t.Errorf("Path: got %q, want %q", got.Path, path)
	}
}

func TestGetByPath_NotFound(t *testing.T) {
	s := newStore(t)

	_, err := s.GetByPath("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for missing path, got nil")
	}
	if !isNotFound(err) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestDelete_Found(t *testing.T) {
	s := newStore(t)
	dir := t.TempDir()

	envA := sampleEnv("a", filepath.Join(dir, "a"))
	envB := sampleEnv("b", filepath.Join(dir, "b"))
	if err := s.Save([]*domain.Environment{envA, envB}); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	if err := s.Delete("a"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	envs, err := s.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(envs) != 1 {
		t.Fatalf("expected 1 environment after delete, got %d", len(envs))
	}
	if envs[0].Name != "b" {
		t.Errorf("expected remaining env to be %q, got %q", "b", envs[0].Name)
	}
}

func TestDelete_NotFound(t *testing.T) {
	s := newStore(t)

	err := s.Delete("ghost")
	if err == nil {
		t.Fatal("expected error when deleting non-existent environment")
	}
	if !isNotFound(err) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestSave_Overwrites(t *testing.T) {
	s := newStore(t)
	dir := t.TempDir()

	env := sampleEnv("env", filepath.Join(dir, "env"))
	if err := s.Save([]*domain.Environment{env}); err != nil {
		t.Fatalf("first Save() error: %v", err)
	}

	// Save an empty list — should replace the previous data.
	if err := s.Save([]*domain.Environment{}); err != nil {
		t.Fatalf("second Save() error: %v", err)
	}

	envs, err := s.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(envs) != 0 {
		t.Errorf("expected 0 environments after overwrite, got %d", len(envs))
	}
}

func TestDir(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.New(dir)
	if got := s.Dir(); got != dir {
		t.Errorf("Dir() = %q, want %q", got, dir)
	}
}

func TestList_FileDoesNotExist(t *testing.T) {
	// Ensure a brand-new store with no file returns empty slice, not error.
	dir := t.TempDir()
	s := jsonstore.New(filepath.Join(dir, "nested", "subdir"))

	envs, err := s.List()
	if err != nil {
		t.Fatalf("List() on missing file: %v", err)
	}
	if envs == nil {
		t.Error("expected non-nil slice, got nil")
	}
	if len(envs) != 0 {
		t.Errorf("expected 0 environments, got %d", len(envs))
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	nestedDir := filepath.Join(dir, "level1", "level2")

	s := jsonstore.New(nestedDir)
	env := sampleEnv("x", filepath.Join(dir, "x"))
	if err := s.Save([]*domain.Environment{env}); err != nil {
		t.Fatalf("Save() to nested dir: %v", err)
	}

	if _, err := os.Stat(nestedDir); err != nil {
		t.Errorf("directory was not created: %v", err)
	}
}

// isNotFound checks whether the error wraps domain.ErrNotFound.
func isNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}
