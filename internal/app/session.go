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
// and maps them back to Orbit environments (or projects) by filesystem path.
type SessionService struct {
	envs      port.EnvironmentRepository
	projects  port.ProjectRepository // nil until project migration is wired
	provider  port.SessionProvider
	lifecycle port.ServerLifecycle // nil when no managed server
	nodeStore port.NodeStore       // nil until node migration is wired
}

// NewSessionService creates a SessionService.
func NewSessionService(envs port.EnvironmentRepository, provider port.SessionProvider) *SessionService {
	return &SessionService{envs: envs, provider: provider}
}

// SetProjects attaches a ProjectRepository. When set, the project-aware
// session methods (ListForProject, CreateSessionForProject) become available.
// This is additive — environment-based methods remain unchanged.
func (s *SessionService) SetProjects(repo port.ProjectRepository) {
	s.projects = repo
}

// SetLifecycle attaches a server lifecycle manager. When set,
// DiscoverServers prefers the managed server over process-table scanning.
func (s *SessionService) SetLifecycle(lc port.ServerLifecycle) {
	s.lifecycle = lc
}

// SetNodeStore attaches a node store. When set, the node-aware session
// methods (ListAllByNode) become available for iterating over registered
// nodes instead of ephemeral server discovery.
func (s *SessionService) SetNodeStore(ns port.NodeStore) {
	s.nodeStore = ns
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
		node := serverToNode(srv)
		sessions, err := s.provider.ListSessions(ctx, node)
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
		node := serverToNode(srv)
		session, err := s.provider.GetSession(ctx, node, sessionID)
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
			node := serverToNode(srv)
			session, err := s.provider.CreateSession(ctx, node, title)
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
	return s.provider.AbortSession(ctx, serverToNode(srv), sessionID)
}

// DeleteSession removes a session.
func (s *SessionService) DeleteSession(ctx context.Context, sessionID string) error {
	srv, err := s.findServerForSession(ctx, sessionID)
	if err != nil {
		return err
	}
	return s.provider.DeleteSession(ctx, serverToNode(srv), sessionID)
}

func (s *SessionService) findServerForSession(ctx context.Context, sessionID string) (domain.Server, error) {
	servers, err := s.DiscoverServers(ctx)
	if err != nil {
		return domain.Server{}, err
	}

	for _, srv := range servers {
		if _, err := s.provider.GetSession(ctx, serverToNode(srv), sessionID); err == nil {
			return srv, nil
		}
	}

	return domain.Server{}, fmt.Errorf("session %q not found on any server", sessionID)
}

// ---------------------------------------------------------------------------
// Node-aware session methods
// ---------------------------------------------------------------------------

// ListAllByNode iterates over all registered nodes and returns every session
// across all of them, enriched with environment mapping. Unlike ListAll
// (which relies on ephemeral server discovery), this method uses the
// persistent node registry — meaning remote/registered nodes are included.
//
// Requires SetNodeStore to have been called; falls back to ListAll if not.
func (s *SessionService) ListAllByNode(ctx context.Context) ([]domain.Session, error) {
	if s.nodeStore == nil {
		return s.ListAll(ctx)
	}

	nodes, err := s.nodeStore.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}

	envs, err := s.envs.List()
	if err != nil {
		return nil, fmt.Errorf("list environments: %w", err)
	}

	seen := make(map[string]bool) // dedup by nodeID
	var all []domain.Session
	for _, node := range nodes {
		if !node.Healthy {
			continue
		}
		if seen[node.ID] {
			continue
		}
		seen[node.ID] = true

		sessions, err := s.provider.ListSessions(ctx, *node)
		if err != nil {
			continue // skip unreachable nodes
		}

		envName, envPath := matchEnvironment(node.Directory, envs)
		for i := range sessions {
			sessions[i].EnvironmentName = envName
			sessions[i].EnvironmentPath = envPath
		}

		all = append(all, sessions...)
	}

	return all, nil
}

// GetSessionByNode looks up a session by ID across all healthy registered
// nodes. Returns the session and the node that hosts it.
//
// Requires SetNodeStore; falls back to GetSession if not configured.
func (s *SessionService) GetSessionByNode(ctx context.Context, sessionID string) (*domain.Session, *domain.Node, error) {
	if s.nodeStore == nil {
		session, err := s.GetSession(ctx, sessionID)
		return session, nil, err
	}

	nodes, err := s.nodeStore.List(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("list nodes: %w", err)
	}

	envs, err := s.envs.List()
	if err != nil {
		return nil, nil, fmt.Errorf("list environments: %w", err)
	}

	for _, node := range nodes {
		if !node.Healthy {
			continue
		}
		session, err := s.provider.GetSession(ctx, *node, sessionID)
		if err != nil {
			continue
		}

		envName, envPath := matchEnvironment(node.Directory, envs)
		session.EnvironmentName = envName
		session.EnvironmentPath = envPath
		return session, node, nil
	}

	return nil, nil, fmt.Errorf("session %q not found on any node", sessionID)
}

// findNodeForSession locates the node that hosts the given session.
// Requires SetNodeStore; returns an error if not configured.
func (s *SessionService) findNodeForSession(ctx context.Context, sessionID string) (*domain.Node, error) {
	if s.nodeStore == nil {
		return nil, fmt.Errorf("node store not configured")
	}

	nodes, err := s.nodeStore.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}

	for _, node := range nodes {
		if !node.Healthy {
			continue
		}
		if _, err := s.provider.GetSession(ctx, *node, sessionID); err == nil {
			return node, nil
		}
	}

	return nil, fmt.Errorf("session %q not found on any node", sessionID)
}

// serverToNode converts a legacy domain.Server to a domain.Node for use
// with the updated SessionProvider interface. This is a transitional bridge
// used until Step 4 replaces server-based discovery with node-based discovery.
func serverToNode(srv domain.Server) domain.Node {
	return domain.Node{
		Hostname:  srv.Hostname,
		Port:      srv.Port,
		Directory: srv.Directory,
		Version:   srv.Version,
		PID:       srv.PID,
		Healthy:   srv.Healthy,
	}
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

// ---------------------------------------------------------------------------
// Project-aware session methods (Stage 3 migration)
// ---------------------------------------------------------------------------

// ListForProject returns sessions running in a specific project's directory.
// Requires SetProjects to have been called; returns an error if not.
func (s *SessionService) ListForProject(ctx context.Context, projectName string) ([]domain.Session, error) {
	if s.projects == nil {
		return nil, fmt.Errorf("project repository not configured")
	}

	proj, err := s.projects.Get(projectName)
	if err != nil {
		return nil, err
	}

	servers, err := s.DiscoverServers(ctx)
	if err != nil {
		return nil, fmt.Errorf("discover servers: %w", err)
	}

	cleanTarget := filepath.Clean(proj.Path)
	var matched []domain.Session
	for _, srv := range servers {
		if !pathContains(cleanTarget, filepath.Clean(srv.Directory)) {
			continue
		}
		sessions, err := s.provider.ListSessions(ctx, serverToNode(srv))
		if err != nil {
			continue // skip unreachable servers
		}
		for i := range sessions {
			sessions[i].EnvironmentName = proj.Name
			sessions[i].EnvironmentPath = proj.Path
		}
		matched = append(matched, sessions...)
	}
	return matched, nil
}

// CreateSessionForProject creates a new session on the server running at
// the given project's directory. Requires SetProjects to have been called.
func (s *SessionService) CreateSessionForProject(ctx context.Context, projectName, title string) (*domain.Session, error) {
	if s.projects == nil {
		return nil, fmt.Errorf("project repository not configured")
	}

	proj, err := s.projects.Get(projectName)
	if err != nil {
		return nil, err
	}

	servers, err := s.DiscoverServers(ctx)
	if err != nil {
		return nil, err
	}

	cleanTarget := filepath.Clean(proj.Path)
	for _, srv := range servers {
		if filepath.Clean(srv.Directory) == cleanTarget {
			node := serverToNode(srv)
			session, err := s.provider.CreateSession(ctx, node, title)
			if err != nil {
				return nil, fmt.Errorf("create session: %w", err)
			}
			session.EnvironmentName = proj.Name
			session.EnvironmentPath = proj.Path
			return session, nil
		}
	}

	return nil, fmt.Errorf("no running server for project %q at %s", projectName, proj.Path)
}

// matchProject finds the best-matching project for a server directory
// using longest-prefix matching.
func matchProject(serverDir string, projects []*domain.Project) (name, path string) {
	if serverDir == "" {
		return "", ""
	}

	cleanDir := filepath.Clean(serverDir)
	bestLen := 0

	for _, p := range projects {
		pDir := filepath.Clean(p.Path)
		if !pathContains(pDir, cleanDir) {
			continue
		}
		if len(pDir) > bestLen {
			bestLen = len(pDir)
			name = p.Name
			path = p.Path
		}
	}
	return name, path
}
