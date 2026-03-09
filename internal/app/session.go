package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cottrellashley/orbit/internal/domain"
	"github.com/cottrellashley/orbit/internal/port"
)

// SessionService discovers coding-agent servers, lists their sessions,
// and maps them back to Orbit environments by filesystem path.
type SessionService struct {
	envs      port.EnvironmentRepository
	provider  port.SessionProvider
	lifecycle port.ServerLifecycle // nil when no managed server
}

// NewSessionService creates a SessionService.
func NewSessionService(envs port.EnvironmentRepository, provider port.SessionProvider) *SessionService {
	return &SessionService{envs: envs, provider: provider}
}

// SetLifecycle attaches a server lifecycle manager. When set,
// DiscoverServers prefers the managed server over process-table scanning.
func (s *SessionService) SetLifecycle(lc port.ServerLifecycle) {
	s.lifecycle = lc
}

// DiscoverServers returns all running coding-agent servers.
// If a managed server is running (via ServerLifecycle), it is included.
// Additionally, any servers found via process-table scanning are included.
func (s *SessionService) DiscoverServers(ctx context.Context) ([]domain.Server, error) {
	var servers []domain.Server

	// Include the managed server if available.
	if s.lifecycle != nil {
		if srv := s.lifecycle.Server(ctx); srv != nil {
			servers = append(servers, *srv)
		}
	}

	// Also include any discovered servers (may overlap with managed).
	discovered, err := s.provider.DiscoverServers(ctx)
	if err != nil {
		// If we already have the managed server, don't fail.
		if len(servers) > 0 {
			return servers, nil
		}
		return nil, fmt.Errorf("discover servers: %w", err)
	}

	// Deduplicate by port — the managed server takes precedence.
	seen := make(map[int]bool, len(servers))
	for _, srv := range servers {
		seen[srv.Port] = true
	}
	for _, srv := range discovered {
		if !seen[srv.Port] {
			servers = append(servers, srv)
			seen[srv.Port] = true
		}
	}

	return servers, nil
}

// ListAll discovers all servers and returns every session across all of them,
// enriched with environment mapping.
func (s *SessionService) ListAll(ctx context.Context) ([]domain.Session, error) {
	servers, err := s.DiscoverServers(ctx)
	if err != nil {
		return nil, fmt.Errorf("discover servers: %w", err)
	}

	envs, err := s.envs.List()
	if err != nil {
		return nil, fmt.Errorf("list environments: %w", err)
	}

	var all []domain.Session
	for _, srv := range servers {
		sessions, err := s.provider.ListSessions(ctx, srv)
		if err != nil {
			continue // skip unreachable servers
		}

		envName, envPath := matchEnvironment(srv.Directory, envs)
		for i := range sessions {
			sessions[i].EnvironmentName = envName
			sessions[i].EnvironmentPath = envPath
		}

		all = append(all, sessions...)
	}

	return all, nil
}

// ListForEnvironment returns sessions running in a specific environment.
func (s *SessionService) ListForEnvironment(ctx context.Context, envName string) ([]domain.Session, error) {
	env, err := s.envs.Get(envName)
	if err != nil {
		return nil, err
	}

	all, err := s.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	var matched []domain.Session
	for _, si := range all {
		if si.EnvironmentName == env.Name {
			matched = append(matched, si)
		}
	}
	return matched, nil
}

// GetSession fetches a single session by ID from any discovered server.
func (s *SessionService) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	servers, err := s.DiscoverServers(ctx)
	if err != nil {
		return nil, fmt.Errorf("discover servers: %w", err)
	}

	envs, err := s.envs.List()
	if err != nil {
		return nil, fmt.Errorf("list environments: %w", err)
	}

	for _, srv := range servers {
		session, err := s.provider.GetSession(ctx, srv, sessionID)
		if err != nil {
			continue
		}

		envName, envPath := matchEnvironment(srv.Directory, envs)
		session.EnvironmentName = envName
		session.EnvironmentPath = envPath
		return session, nil
	}

	return nil, fmt.Errorf("session %q not found on any server", sessionID)
}

// CreateSession creates a new session on the server running at the
// given environment's directory.
func (s *SessionService) CreateSession(ctx context.Context, envName, title string) (*domain.Session, error) {
	env, err := s.envs.Get(envName)
	if err != nil {
		return nil, err
	}

	servers, err := s.DiscoverServers(ctx)
	if err != nil {
		return nil, err
	}

	cleanTarget := filepath.Clean(env.Path)
	for _, srv := range servers {
		if filepath.Clean(srv.Directory) == cleanTarget {
			session, err := s.provider.CreateSession(ctx, srv, title)
			if err != nil {
				return nil, fmt.Errorf("create session: %w", err)
			}
			session.EnvironmentName = envName
			session.EnvironmentPath = env.Path
			return session, nil
		}
	}

	return nil, fmt.Errorf("no running server for environment %q at %s", envName, env.Path)
}

// AbortSession stops a running session.
func (s *SessionService) AbortSession(ctx context.Context, sessionID string) error {
	srv, err := s.findServerForSession(ctx, sessionID)
	if err != nil {
		return err
	}
	return s.provider.AbortSession(ctx, srv, sessionID)
}

// DeleteSession removes a session.
func (s *SessionService) DeleteSession(ctx context.Context, sessionID string) error {
	srv, err := s.findServerForSession(ctx, sessionID)
	if err != nil {
		return err
	}
	return s.provider.DeleteSession(ctx, srv, sessionID)
}

func (s *SessionService) findServerForSession(ctx context.Context, sessionID string) (domain.Server, error) {
	servers, err := s.DiscoverServers(ctx)
	if err != nil {
		return domain.Server{}, err
	}

	for _, srv := range servers {
		if _, err := s.provider.GetSession(ctx, srv, sessionID); err == nil {
			return srv, nil
		}
	}

	return domain.Server{}, fmt.Errorf("session %q not found on any server", sessionID)
}

// matchEnvironment finds the best-matching environment for a server directory
// using longest-prefix matching.
func matchEnvironment(serverDir string, envs []*domain.Environment) (name, path string) {
	if serverDir == "" {
		return "", ""
	}

	cleanDir := filepath.Clean(serverDir)
	bestLen := 0

	for _, env := range envs {
		envDir := filepath.Clean(env.Path)
		if !pathContains(envDir, cleanDir) {
			continue
		}
		if len(envDir) > bestLen {
			bestLen = len(envDir)
			name = env.Name
			path = env.Path
		}
	}
	return name, path
}

func pathContains(dir, target string) bool {
	if dir == target {
		return true
	}
	prefix := dir
	if !strings.HasSuffix(prefix, string(filepath.Separator)) {
		prefix += string(filepath.Separator)
	}
	return strings.HasPrefix(target, prefix)
}
