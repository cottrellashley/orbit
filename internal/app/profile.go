package app

import (
	"github.com/cottrellashley/orbit/internal/domain"
	"github.com/cottrellashley/orbit/internal/port"
)

// ProfileService manages reusable starter profiles.
type ProfileService struct {
	repo port.ProfileRepository
}

// NewProfileService creates a ProfileService.
func NewProfileService(repo port.ProfileRepository) *ProfileService {
	return &ProfileService{repo: repo}
}

// List returns all available profiles.
func (s *ProfileService) List() ([]*domain.Profile, error) {
	return s.repo.List()
}

// Get returns a single profile by name.
func (s *ProfileService) Get(name string) (*domain.Profile, error) {
	return s.repo.Get(name)
}

// Create scaffolds a new profile.
func (s *ProfileService) Create(name, description string) (*domain.Profile, error) {
	return s.repo.Create(name, description)
}

// Delete removes a profile.
func (s *ProfileService) Delete(name string) error {
	return s.repo.Delete(name)
}

// Apply copies profile contents into a target directory.
func (s *ProfileService) Apply(name, targetDir string) error {
	return s.repo.Apply(name, targetDir)
}

// Path returns the absolute path to a profile directory.
func (s *ProfileService) Path(name string) string {
	return s.repo.Path(name)
}

// Dir returns the base profiles directory.
func (s *ProfileService) Dir() string {
	return s.repo.Dir()
}
