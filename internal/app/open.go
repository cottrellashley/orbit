package app

import (
	"context"

	"github.com/cottrellashley/orbit/internal/domain"
	"github.com/cottrellashley/orbit/internal/port"
)

// sessionQuerier is the subset of SessionService that OpenService needs.
// Defined here (consumer-side) so OpenService depends on an interface,
// not a concrete type.
type sessionQuerier interface {
	ListForEnvironment(ctx context.Context, envName string) ([]domain.Session, error)
	DiscoverServers(ctx context.Context) ([]domain.Server, error)
}

// OpenService orchestrates the `orbit open` flow: resolve an environment,
// discover its sessions, and produce a plan for the driver to execute.
type OpenService struct {
	envs     port.EnvironmentRepository
	sessions sessionQuerier
}

// NewOpenService creates an OpenService.
func NewOpenService(envs port.EnvironmentRepository, sessions sessionQuerier) *OpenService {
	return &OpenService{envs: envs, sessions: sessions}
}

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
