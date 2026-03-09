package app

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
	"github.com/cottrellashley/orbit/internal/port"
)

// EnvironmentService manages the registry of orbit environments.
type EnvironmentService struct {
	repo     port.EnvironmentRepository
	profiles port.ProfileRepository
}

// NewEnvironmentService creates an EnvironmentService.
func NewEnvironmentService(repo port.EnvironmentRepository, profiles port.ProfileRepository) *EnvironmentService {
	return &EnvironmentService{repo: repo, profiles: profiles}
}

// List returns all registered environments.
func (s *EnvironmentService) List() ([]*domain.Environment, error) {
	return s.repo.List()
}

// Get returns a single environment by name.
func (s *EnvironmentService) Get(name string) (*domain.Environment, error) {
	return s.repo.Get(name)
}

// GetByPath returns the environment whose registered path matches.
func (s *EnvironmentService) GetByPath(path string) (*domain.Environment, error) {
	return s.repo.GetByPath(path)
}

// Register adds a new environment to the registry.
func (s *EnvironmentService) Register(name, path, description string) (*domain.Environment, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	info, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("path %q: %w", abs, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path %q is not a directory", abs)
	}

	// Check for name collision
	existing, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	for _, e := range existing {
		if e.Name == name {
			return nil, fmt.Errorf("environment %q: %w", name, domain.ErrAlreadyExists)
		}
	}

	now := time.Now().Truncate(time.Second)
	env := &domain.Environment{
		Name:        name,
		Path:        abs,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	existing = append(existing, env)
	if err := s.repo.Save(existing); err != nil {
		return nil, err
	}

	return env, nil
}

// CreateFromProfile creates a new environment by copying a profile's
// contents into the target path and registering it.
func (s *EnvironmentService) CreateFromProfile(name, profileName, path, description string, createDir bool) (*domain.Environment, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	// Verify profile exists
	if _, err := s.profiles.Get(profileName); err != nil {
		return nil, err
	}

	// Check for name collision
	existing, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	for _, e := range existing {
		if e.Name == name {
			return nil, fmt.Errorf("environment %q: %w", name, domain.ErrAlreadyExists)
		}
	}

	// Create directory if requested
	if createDir {
		if err := os.MkdirAll(abs, 0755); err != nil {
			return nil, fmt.Errorf("create directory %q: %w", abs, err)
		}
	} else {
		info, err := os.Stat(abs)
		if err != nil {
			return nil, fmt.Errorf("path %q: %w", abs, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("path %q is not a directory", abs)
		}
	}

	// Apply profile contents
	if err := s.profiles.Apply(profileName, abs); err != nil {
		return nil, fmt.Errorf("apply profile %q: %w", profileName, err)
	}

	now := time.Now().Truncate(time.Second)
	env := &domain.Environment{
		Name:        name,
		Path:        abs,
		ProfileName: profileName,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	existing = append(existing, env)
	if err := s.repo.Save(existing); err != nil {
		return nil, err
	}

	return env, nil
}

// Delete removes an environment from the registry.
func (s *EnvironmentService) Delete(name string) error {
	return s.repo.Delete(name)
}
