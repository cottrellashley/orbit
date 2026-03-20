package app

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
	"github.com/cottrellashley/orbit/internal/port"
)

// ProjectService manages the registry of Orbit projects.
//
// This is the successor to EnvironmentService. During the migration
// period both services may coexist — ProjectService operates on the
// new domain.Project type while EnvironmentService continues to serve
// existing callers unchanged.
type ProjectService struct {
	repo     port.ProjectRepository
	profiles port.ProfileRepository
}

// NewProjectService creates a ProjectService.
// The profiles parameter may be nil if profile support is not yet wired.
func NewProjectService(repo port.ProjectRepository, profiles port.ProfileRepository) *ProjectService {
	return &ProjectService{repo: repo, profiles: profiles}
}

// List returns all registered projects.
func (s *ProjectService) List() ([]*domain.Project, error) {
	return s.repo.List()
}

// Get returns a single project by name.
func (s *ProjectService) Get(name string) (*domain.Project, error) {
	return s.repo.Get(name)
}

// GetByPath returns the project whose registered path matches.
func (s *ProjectService) GetByPath(path string) (*domain.Project, error) {
	return s.repo.GetByPath(path)
}

// Register adds a new project to the registry.
// Topology and integrations are left at their zero values — a separate
// detection pass populates them later.
func (s *ProjectService) Register(name, path, description string) (*domain.Project, error) {
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

	existing, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	for _, p := range existing {
		if p.Name == name {
			return nil, fmt.Errorf("project %q: %w", name, domain.ErrAlreadyExists)
		}
	}

	now := time.Now().Truncate(time.Second)
	project := &domain.Project{
		Name:        name,
		Path:        abs,
		Description: description,
		Topology:    domain.TopologyUnknown,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	existing = append(existing, project)
	if err := s.repo.Save(existing); err != nil {
		return nil, err
	}

	return project, nil
}

// CreateFromProfile creates a new project by applying a profile's
// contents into the target path and registering the project.
func (s *ProjectService) CreateFromProfile(name, profileName, path, description string, createDir bool) (*domain.Project, error) {
	if s.profiles == nil {
		return nil, fmt.Errorf("profile repository not configured")
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	// Verify profile exists.
	if _, err := s.profiles.Get(profileName); err != nil {
		return nil, err
	}

	// Check for name collision.
	existing, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	for _, p := range existing {
		if p.Name == name {
			return nil, fmt.Errorf("project %q: %w", name, domain.ErrAlreadyExists)
		}
	}

	// Create directory if requested.
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

	// Apply profile contents.
	if err := s.profiles.Apply(profileName, abs); err != nil {
		return nil, fmt.Errorf("apply profile %q: %w", profileName, err)
	}

	now := time.Now().Truncate(time.Second)
	project := &domain.Project{
		Name:        name,
		Path:        abs,
		Description: description,
		ProfileName: profileName,
		Topology:    domain.TopologyUnknown,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	existing = append(existing, project)
	if err := s.repo.Save(existing); err != nil {
		return nil, err
	}

	return project, nil
}

// Update replaces a project's mutable metadata (topology, integrations,
// repos, description) and bumps UpdatedAt. The project is identified by
// name. Fields that callers do not wish to change should carry their
// existing values — Update is a full-replace on the mutable fields.
func (s *ProjectService) Update(name string, topology domain.ProjectTopology, integrations []domain.IntegrationTag, repos []domain.RepoInfo, description string) (*domain.Project, error) {
	projects, err := s.repo.List()
	if err != nil {
		return nil, err
	}

	var target *domain.Project
	for _, p := range projects {
		if p.Name == name {
			target = p
			break
		}
	}
	if target == nil {
		return nil, fmt.Errorf("project %q: %w", name, domain.ErrNotFound)
	}

	target.Topology = topology
	target.Integrations = integrations
	target.Repos = repos
	target.Description = description
	target.UpdatedAt = time.Now().Truncate(time.Second)

	if err := s.repo.Save(projects); err != nil {
		return nil, err
	}
	return target, nil
}

// Delete removes a project from the registry.
func (s *ProjectService) Delete(name string) error {
	return s.repo.Delete(name)
}

// ---------------------------------------------------------------------------
// Migration compatibility shims
// ---------------------------------------------------------------------------

// ListAsEnvironments returns all projects converted to legacy Environment
// values. This allows existing environment-centric drivers and services to
// consume project data without changes.
func (s *ProjectService) ListAsEnvironments() ([]*domain.Environment, error) {
	projects, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	envs := make([]*domain.Environment, len(projects))
	for i, p := range projects {
		envs[i] = domain.EnvironmentFromProject(p)
	}
	return envs, nil
}

// GetAsEnvironment returns a single project converted to a legacy Environment.
func (s *ProjectService) GetAsEnvironment(name string) (*domain.Environment, error) {
	p, err := s.repo.Get(name)
	if err != nil {
		return nil, err
	}
	return domain.EnvironmentFromProject(p), nil
}
