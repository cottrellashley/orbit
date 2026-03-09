// Package jsonstore implements port.EnvironmentRepository using atomic
// JSON file persistence. Each write goes to a temp file first, then is
// renamed into place — a crash mid-write cannot corrupt existing data.
//
// The store is single-user (no file locking). The data file lives at
// <dir>/environments.json.
package jsonstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
)

const collectionName = "environments"

// ---------------------------------------------------------------------------
// JSON DTO — keeps serialization details out of the domain
// ---------------------------------------------------------------------------

type envDTO struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	ProfileName string    `json:"profile_name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toDTO(e *domain.Environment) envDTO {
	return envDTO{
		Name:        e.Name,
		Path:        e.Path,
		ProfileName: e.ProfileName,
		Description: e.Description,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func fromDTO(d envDTO) *domain.Environment {
	return &domain.Environment{
		Name:        d.Name,
		Path:        d.Path,
		ProfileName: d.ProfileName,
		Description: d.Description,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

// Store persists environment registry entries as a single JSON file.
type Store struct {
	dir string
}

// New creates a Store rooted at dir. The directory is created on the
// first call to Save.
func New(dir string) *Store {
	return &Store{dir: dir}
}

// Dir returns the store's base directory.
func (s *Store) Dir() string {
	return s.dir
}

// ---------------------------------------------------------------------------
// port.EnvironmentRepository implementation
// ---------------------------------------------------------------------------

// List returns all registered environments.
// Returns an empty slice (not nil) if none exist or the file is missing.
func (s *Store) List() ([]*domain.Environment, error) {
	envs, err := s.load()
	if err != nil {
		return nil, err
	}
	return envs, nil
}

// Get returns a single environment by name.
// Returns domain.ErrNotFound if not found.
func (s *Store) Get(name string) (*domain.Environment, error) {
	envs, err := s.load()
	if err != nil {
		return nil, err
	}
	for _, e := range envs {
		if e.Name == name {
			return e, nil
		}
	}
	return nil, fmt.Errorf("environment %q: %w", name, domain.ErrNotFound)
}

// GetByPath returns the environment whose registered path matches.
// Returns domain.ErrNotFound if no match.
func (s *Store) GetByPath(path string) (*domain.Environment, error) {
	envs, err := s.load()
	if err != nil {
		return nil, err
	}
	for _, e := range envs {
		if e.Path == path {
			return e, nil
		}
	}
	return nil, fmt.Errorf("environment at path %q: %w", path, domain.ErrNotFound)
}

// Save persists the full environment list (create or update).
func (s *Store) Save(envs []*domain.Environment) error {
	return s.save(envs)
}

// Delete removes an environment by name.
// Returns domain.ErrNotFound if not found.
func (s *Store) Delete(name string) error {
	envs, err := s.load()
	if err != nil {
		return err
	}

	idx := -1
	for i, e := range envs {
		if e.Name == name {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("environment %q: %w", name, domain.ErrNotFound)
	}

	envs = append(envs[:idx], envs[idx+1:]...)
	return s.save(envs)
}

// ---------------------------------------------------------------------------
// Internal persistence
// ---------------------------------------------------------------------------

func (s *Store) filePath() string {
	return filepath.Join(s.dir, collectionName+".json")
}

// load reads the environments file and returns the slice.
// Returns an empty (non-nil) slice if the file does not exist.
func (s *Store) load() ([]*domain.Environment, error) {
	p := s.filePath()

	b, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []*domain.Environment{}, nil
		}
		return nil, fmt.Errorf("jsonstore: read %s: %w", p, err)
	}

	var dtos []envDTO
	if err := json.Unmarshal(b, &dtos); err != nil {
		return nil, fmt.Errorf("jsonstore: parse %s: %w", p, err)
	}

	envs := make([]*domain.Environment, len(dtos))
	for i, d := range dtos {
		envs[i] = fromDTO(d)
	}
	return envs, nil
}

// save writes the environments slice atomically.
func (s *Store) save(envs []*domain.Environment) error {
	if err := os.MkdirAll(s.dir, 0755); err != nil {
		return fmt.Errorf("jsonstore: create directory %s: %w", s.dir, err)
	}

	dtos := make([]envDTO, len(envs))
	for i, e := range envs {
		dtos[i] = toDTO(e)
	}

	b, err := json.MarshalIndent(dtos, "", "  ")
	if err != nil {
		return fmt.Errorf("jsonstore: marshal: %w", err)
	}
	b = append(b, '\n')

	// Atomic write: temp file + rename.
	tmp, err := os.CreateTemp(s.dir, collectionName+"-*.tmp")
	if err != nil {
		return fmt.Errorf("jsonstore: create temp: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("jsonstore: write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("jsonstore: close temp: %w", err)
	}

	dest := s.filePath()
	if err := os.Rename(tmpPath, dest); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("jsonstore: rename %s -> %s: %w", tmpPath, dest, err)
	}

	return nil
}
