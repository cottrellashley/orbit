package app

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
	"github.com/cottrellashley/orbit/internal/port"
)

// generateUUID produces a version-4 UUID using crypto/rand.
// This avoids pulling in an external UUID library for a single use.
func generateUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// uuidFunc is the function used to generate new UUIDs.
// Tests can replace this to produce deterministic IDs.
var uuidFunc = generateUUID

// NodeService manages the node registry. It orchestrates two ports:
//   - NodeStore for persistence of node records
//   - SessionProvider for discovering running servers (local process scan)
//
// Business rules live here: UUID assignment, origin tracking, reconciliation
// of discovered servers against the persistent registry.
type NodeService struct {
	store    port.NodeStore
	provider port.SessionProvider
}

// NewNodeService creates a NodeService.
func NewNodeService(store port.NodeStore, provider port.SessionProvider) *NodeService {
	return &NodeService{store: store, provider: provider}
}

// RegisterNode adds a new node to the registry (manual registration).
// If a node with the same hostname:port already exists, it is updated
// in-place and the existing ID is preserved.
func (s *NodeService) RegisterNode(ctx context.Context, hostname string, port int, provider domain.NodeProvider, name string) (*domain.Node, error) {
	if hostname == "" {
		return nil, fmt.Errorf("hostname is required")
	}
	if port <= 0 || port > 65535 {
		return nil, fmt.Errorf("port must be between 1 and 65535")
	}

	// Check for existing node at this hostname:port.
	existing, err := s.store.GetByHostPort(ctx, hostname, port)
	if err == nil {
		// Update in-place — preserve ID and origin.
		existing.Provider = provider
		existing.Name = name
		existing.UpdatedAt = time.Now()
		if err := s.store.Save(ctx, existing); err != nil {
			return nil, fmt.Errorf("update existing node: %w", err)
		}
		return existing, nil
	}
	if !errors.Is(err, domain.ErrNodeNotFound) {
		return nil, fmt.Errorf("check existing node: %w", err)
	}

	now := time.Now()
	node := &domain.Node{
		ID:        uuidFunc(),
		Name:      name,
		Provider:  provider,
		Origin:    domain.OriginRegistered,
		Hostname:  hostname,
		Port:      port,
		Healthy:   false, // unknown until first health check
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.store.Save(ctx, node); err != nil {
		return nil, fmt.Errorf("save node: %w", err)
	}
	return node, nil
}

// RemoveNode deletes a node from the registry by its stable ID.
func (s *NodeService) RemoveNode(ctx context.Context, id string) error {
	if err := s.store.Delete(ctx, id); err != nil {
		return fmt.Errorf("remove node: %w", err)
	}
	return nil
}

// ListNodes returns all nodes in the registry.
func (s *NodeService) ListNodes(ctx context.Context) ([]*domain.Node, error) {
	nodes, err := s.store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}
	return nodes, nil
}

// GetNode returns a single node by its stable ID.
func (s *NodeService) GetNode(ctx context.Context, id string) (*domain.Node, error) {
	node, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get node: %w", err)
	}
	return node, nil
}

// SyncDiscoveredNodes calls DiscoverServers on the provider, then
// reconciles the results with the persistent node store:
//
//   - New discoveries (no matching hostname:port) are registered with
//     OriginDiscovered.
//   - Existing nodes that match a discovered server are updated with
//     fresh metadata (PID, version, healthy, directory) and UpdatedAt.
//   - Discovered-origin nodes that are NOT in the discovery results are
//     marked unhealthy (they may have been stopped). Registered-origin
//     nodes are left alone — the user explicitly added them.
//
// Returns the full list of nodes after reconciliation.
func (s *NodeService) SyncDiscoveredNodes(ctx context.Context) ([]*domain.Node, error) {
	servers, err := s.provider.DiscoverServers(ctx)
	if err != nil {
		return nil, fmt.Errorf("discover servers: %w", err)
	}

	// Build a set of discovered hostname:port for stale detection.
	type hostPort struct {
		hostname string
		port     int
	}
	discoveredSet := make(map[hostPort]domain.Server, len(servers))
	for _, srv := range servers {
		hp := hostPort{hostname: srv.Hostname, port: srv.Port}
		discoveredSet[hp] = srv
	}

	// Reconcile each discovered server against the store.
	now := time.Now()
	for _, srv := range servers {
		existing, err := s.store.GetByHostPort(ctx, srv.Hostname, srv.Port)
		if errors.Is(err, domain.ErrNodeNotFound) {
			// New discovery — register it.
			node := &domain.Node{
				ID:        uuidFunc(),
				Provider:  domain.ProviderOpenCode, // discovery is provider-specific today
				Origin:    domain.OriginDiscovered,
				Hostname:  srv.Hostname,
				Port:      srv.Port,
				Directory: srv.Directory,
				Version:   srv.Version,
				PID:       srv.PID,
				Healthy:   srv.Healthy,
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := s.store.Save(ctx, node); err != nil {
				return nil, fmt.Errorf("save discovered node: %w", err)
			}
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("lookup node by host:port: %w", err)
		}

		// Existing — refresh metadata.
		existing.PID = srv.PID
		existing.Directory = srv.Directory
		existing.Version = srv.Version
		existing.Healthy = srv.Healthy
		existing.UpdatedAt = now
		if err := s.store.Save(ctx, existing); err != nil {
			return nil, fmt.Errorf("update discovered node: %w", err)
		}
	}

	// Mark stale discovered-origin nodes as unhealthy.
	allNodes, err := s.store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list nodes after sync: %w", err)
	}

	for _, node := range allNodes {
		if node.Origin != domain.OriginDiscovered {
			continue // never touch manually-registered nodes
		}
		hp := hostPort{hostname: node.Hostname, port: node.Port}
		if _, found := discoveredSet[hp]; found {
			continue // still alive
		}
		if node.Healthy {
			node.Healthy = false
			node.UpdatedAt = now
			if err := s.store.Save(ctx, node); err != nil {
				return nil, fmt.Errorf("mark stale node unhealthy: %w", err)
			}
		}
	}

	// Return the final state.
	return s.store.List(ctx)
}
