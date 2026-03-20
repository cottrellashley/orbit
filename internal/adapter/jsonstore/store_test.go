package jsonstore_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cottrellashley/orbit/internal/adapter/jsonstore"
	"github.com/cottrellashley/orbit/internal/domain"
	"github.com/cottrellashley/orbit/internal/port"
)

// Verify Store satisfies port.EnvironmentRepository at compile time.
var _ port.EnvironmentRepository = (*jsonstore.Store)(nil)

func TestStore_ListEmpty(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.New(dir)

	envs, err := s.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(envs) != 0 {
		t.Errorf("List() returned %d envs, want 0", len(envs))
	}
}

func TestStore_SaveAndList(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.New(dir)

	now := time.Date(2026, 3, 9, 12, 0, 0, 0, time.UTC)
	envs := []*domain.Environment{
		{Name: "test-env", Path: "/home/user/test", Description: "test", ProfileName: "default", CreatedAt: now, UpdatedAt: now},
	}

	if err := s.Save(envs); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := s.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("List() returned %d envs, want 1", len(loaded))
	}
	if loaded[0].Name != "test-env" {
		t.Errorf("Name = %q, want %q", loaded[0].Name, "test-env")
	}
	if loaded[0].ProfileName != "default" {
		t.Errorf("ProfileName = %q, want %q", loaded[0].ProfileName, "default")
	}
}

func TestStore_Get(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.New(dir)

	now := time.Now().Truncate(time.Second)
	envs := []*domain.Environment{
		{Name: "alpha", Path: "/a", CreatedAt: now, UpdatedAt: now},
		{Name: "beta", Path: "/b", CreatedAt: now, UpdatedAt: now},
	}
	if err := s.Save(envs); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	e, err := s.Get("alpha")
	if err != nil {
		t.Fatalf("Get(alpha) error: %v", err)
	}
	if e.Name != "alpha" {
		t.Errorf("Name = %q, want %q", e.Name, "alpha")
	}

	_, err = s.Get("nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get(nonexistent) error = %v, want ErrNotFound", err)
	}
}

func TestStore_GetByPath(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.New(dir)

	now := time.Now().Truncate(time.Second)
	envs := []*domain.Environment{
		{Name: "alpha", Path: "/a", CreatedAt: now, UpdatedAt: now},
	}
	if err := s.Save(envs); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	e, err := s.GetByPath("/a")
	if err != nil {
		t.Fatalf("GetByPath(/a) error: %v", err)
	}
	if e.Name != "alpha" {
		t.Errorf("Name = %q, want %q", e.Name, "alpha")
	}

	_, err = s.GetByPath("/nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByPath(nonexistent) error = %v, want ErrNotFound", err)
	}
}

func TestStore_Delete(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.New(dir)

	now := time.Now().Truncate(time.Second)
	envs := []*domain.Environment{
		{Name: "alpha", Path: "/a", CreatedAt: now, UpdatedAt: now},
		{Name: "beta", Path: "/b", CreatedAt: now, UpdatedAt: now},
	}
	if err := s.Save(envs); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	if err := s.Delete("alpha"); err != nil {
		t.Fatalf("Delete(alpha) error: %v", err)
	}

	remaining, err := s.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(remaining) != 1 || remaining[0].Name != "beta" {
		t.Errorf("after delete: got %v, want [beta]", remaining)
	}

	err = s.Delete("nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Delete(nonexistent) error = %v, want ErrNotFound", err)
	}
}

func TestStore_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.New(dir)

	now := time.Now().Truncate(time.Second)
	if err := s.Save([]*domain.Environment{
		{Name: "test", Path: "/test", CreatedAt: now, UpdatedAt: now},
	}); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify no leftover temp files.
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("leftover temp file: %s", e.Name())
		}
	}
}

// Verify that the existing Environment store and new Project store
// coexist in the same directory without interference.
func TestStore_CoexistWithProjectStore(t *testing.T) {
	dir := t.TempDir()
	envStore := jsonstore.New(dir)
	projStore := jsonstore.NewProjectStore(dir)

	now := time.Now().Truncate(time.Second)

	// Save environments.
	if err := envStore.Save([]*domain.Environment{
		{Name: "env-one", Path: "/env1", CreatedAt: now, UpdatedAt: now},
	}); err != nil {
		t.Fatalf("envStore.Save() error: %v", err)
	}

	// Save projects.
	if err := projStore.Save([]*domain.Project{
		{Name: "proj-one", Path: "/proj1", Topology: domain.TopologySingleRepo, CreatedAt: now, UpdatedAt: now},
	}); err != nil {
		t.Fatalf("projStore.Save() error: %v", err)
	}

	// Both stores should return their own data independently.
	envs, err := envStore.List()
	if err != nil {
		t.Fatalf("envStore.List() error: %v", err)
	}
	if len(envs) != 1 || envs[0].Name != "env-one" {
		t.Errorf("envStore.List() = %v, want [env-one]", envs)
	}

	projs, err := projStore.List()
	if err != nil {
		t.Fatalf("projStore.List() error: %v", err)
	}
	if len(projs) != 1 || projs[0].Name != "proj-one" {
		t.Errorf("projStore.List() = %v, want [proj-one]", projs)
	}
}
