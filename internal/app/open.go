package app

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/cottrellashley/orbit/internal/domain"
	"github.com/cottrellashley/orbit/internal/port"
)

// sessionQuerier is the subset of SessionService that OpenService needs
// for environment-based flows. Defined here (consumer-side) so
// OpenService depends on an interface, not a concrete type.
type sessionQuerier interface {
	ListForEnvironment(ctx context.Context, envName string) ([]domain.Session, error)
	DiscoverServers(ctx context.Context) ([]domain.Server, error)
}

// projectSessionQuerier is the subset of SessionService that the
// project-based open flow needs. Separate from sessionQuerier so that
// existing callers are unchanged.
type projectSessionQuerier interface {
	ListForProject(ctx context.Context, projectName string) ([]domain.Session, error)
	DiscoverServers(ctx context.Context) ([]domain.Server, error)
}

// OpenService orchestrates the `orbit open` flow: resolve an environment
// (or project), discover its sessions, and produce a plan for the driver
// to execute.
type OpenService struct {
	envs         port.EnvironmentRepository
	projects     port.ProjectRepository // nil until project migration is wired
	sessions     sessionQuerier
	projSessions projectSessionQuerier // nil until project migration is wired
}

// NewOpenService creates an OpenService.
func NewOpenService(envs port.EnvironmentRepository, sessions sessionQuerier) *OpenService {
	return &OpenService{envs: envs, sessions: sessions}
}

// SetProjects attaches the project-first dependencies. When set, the
// ResolveProject / ResolveProjectByPath / ServerForProject methods
// become available.
func (s *OpenService) SetProjects(repo port.ProjectRepository, psq projectSessionQuerier) {
	s.projects = repo
	s.projSessions = psq
}

// ---------------------------------------------------------------------------
// Environment-based flow (existing, unchanged)
// ---------------------------------------------------------------------------

// Resolve looks up an environment by name and produces an OpenPlan.
func (s *OpenService) Resolve(ctx context.Context, envName string) (*domain.OpenPlan, error) {
	env, err := s.envs.Get(envName)
	if err != nil {
		return nil, err
	}

	sessions, err := s.sessions.ListForEnvironment(ctx, envName)
	if err != nil {
		sessions = nil // discovery failure is not fatal
	}

	plan := &domain.OpenPlan{
		Environment: env,
		Sessions:    sessions,
	}

	switch len(sessions) {
	case 0:
		plan.Action = domain.OpenActionCreate
	case 1:
		plan.Action = domain.OpenActionResume
	default:
		plan.Action = domain.OpenActionSelect
	}

	return plan, nil
}

// ResolveByPath looks up an environment by path and produces an OpenPlan.
func (s *OpenService) ResolveByPath(ctx context.Context, path string) (*domain.OpenPlan, error) {
	env, err := s.envs.GetByPath(path)
	if err != nil {
		return nil, err
	}
	return s.Resolve(ctx, env.Name)
}

// ServerForEnvironment checks whether a coding-agent server is running
// for the given environment. Returns the server info if found, nil if not.
func (s *OpenService) ServerForEnvironment(ctx context.Context, envName string) (*domain.Server, error) {
	env, err := s.envs.Get(envName)
	if err != nil {
		return nil, err
	}

	servers, err := s.sessions.DiscoverServers(ctx)
	if err != nil {
		return nil, err
	}

	for _, srv := range servers {
		if srv.Directory == env.Path {
			return &srv, nil
		}
	}

	return nil, nil // no server running — not an error
}

// ---------------------------------------------------------------------------
// Project-first flow (Stage 3 migration)
// ---------------------------------------------------------------------------

// ResolveProject looks up a project by name and produces a ProjectOpenPlan.
// Requires SetProjects to have been called.
func (s *OpenService) ResolveProject(ctx context.Context, projectName string) (*domain.ProjectOpenPlan, error) {
	if s.projects == nil {
		return nil, fmt.Errorf("project repository not configured")
	}

	proj, err := s.projects.Get(projectName)
	if err != nil {
		return nil, err
	}

	return s.buildProjectPlan(ctx, proj)
}

// ResolveProjectByPath looks up a project by its filesystem path and
// produces a ProjectOpenPlan. Requires SetProjects to have been called.
func (s *OpenService) ResolveProjectByPath(ctx context.Context, path string) (*domain.ProjectOpenPlan, error) {
	if s.projects == nil {
		return nil, fmt.Errorf("project repository not configured")
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	proj, err := s.projects.GetByPath(abs)
	if err != nil {
		return nil, err
	}

	return s.buildProjectPlan(ctx, proj)
}

// ServerForProject checks whether a coding-agent server is running
// for the given project. Returns the server info if found, nil if not.
// Requires SetProjects to have been called.
func (s *OpenService) ServerForProject(ctx context.Context, projectName string) (*domain.Server, error) {
	if s.projects == nil {
		return nil, fmt.Errorf("project repository not configured")
	}

	proj, err := s.projects.Get(projectName)
	if err != nil {
		return nil, err
	}

	servers, err := s.sessions.DiscoverServers(ctx)
	if err != nil {
		return nil, err
	}

	cleanTarget := filepath.Clean(proj.Path)
	for _, srv := range servers {
		if filepath.Clean(srv.Directory) == cleanTarget {
			return &srv, nil
		}
	}

	return nil, nil // no server running — not an error
}

// buildProjectPlan creates a ProjectOpenPlan from a resolved project.
func (s *OpenService) buildProjectPlan(ctx context.Context, proj *domain.Project) (*domain.ProjectOpenPlan, error) {
	var sessions []domain.Session
	if s.projSessions != nil {
		var err error
		sessions, err = s.projSessions.ListForProject(ctx, proj.Name)
		if err != nil {
			sessions = nil // discovery failure is not fatal
		}
	}

	// Check server presence.
	serverOnline := false
	servers, err := s.sessions.DiscoverServers(ctx)
	if err == nil {
		cleanTarget := filepath.Clean(proj.Path)
		for _, srv := range servers {
			if filepath.Clean(srv.Directory) == cleanTarget {
				serverOnline = true
				break
			}
		}
	}

	plan := &domain.ProjectOpenPlan{
		Project:      proj,
		Sessions:     sessions,
		ServerOnline: serverOnline,
	}

	switch len(sessions) {
	case 0:
		plan.Action = domain.OpenActionCreate
	case 1:
		plan.Action = domain.OpenActionResume
	default:
		plan.Action = domain.OpenActionSelect
	}

	return plan, nil
}
