package port

import (
	"github.com/cottrellashley/orbit/internal/domain"
)

// EnvironmentRepository persists and retrieves environment registry entries.
type EnvironmentRepository interface {
	// List returns all registered environments.
	// Returns an empty slice (not nil) if none exist.
	List() ([]*domain.Environment, error)

	// Get returns a single environment by name.
	// Returns domain.ErrNotFound if not found.
	Get(name string) (*domain.Environment, error)

	// GetByPath returns the environment whose registered path matches.
	// Returns domain.ErrNotFound if no match.
	GetByPath(path string) (*domain.Environment, error)

	// Save persists the full environment list (create or update).
	Save(envs []*domain.Environment) error

	// Delete removes an environment by name.
	// Returns domain.ErrNotFound if not found.
	Delete(name string) error
}

// ProfileRepository manages profile starter kits on disk.
type ProfileRepository interface {
	// List returns all available profiles.
	List() ([]*domain.Profile, error)

	// Get returns a single profile by name.
	// Returns domain.ErrNotFound if not found.
	Get(name string) (*domain.Profile, error)

	// Create scaffolds a new profile directory with metadata.
	Create(name, description string) (*domain.Profile, error)

	// Delete removes a profile directory.
	Delete(name string) error

	// Apply copies profile contents into a target directory.
	// Existing files in the target are NOT overwritten.
	Apply(profileName, targetDir string) error

	// Path returns the absolute path to a profile directory.
	Path(name string) string

	// Dir returns the base profiles directory.
	Dir() string
}
