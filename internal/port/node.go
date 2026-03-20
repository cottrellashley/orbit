package port

import (
	"context"

	"github.com/cottrellashley/orbit/internal/domain"
)

// NodeStore persists and retrieves node registry entries. Nodes are
// stored with a stable UUID so their identity survives hostname:port
// changes across restarts.
//
// Today this is backed by a JSON file (jsonstore adapter); tomorrow
// it may be SQLite or another backend — callers depend only on this
// interface.
type NodeStore interface {
	// List returns all registered nodes.
	// Returns an empty slice (not nil) if none exist.
	List(ctx context.Context) ([]*domain.Node, error)

	// Get returns a single node by its stable ID.
	// Returns domain.ErrNodeNotFound if not found.
	Get(ctx context.Context, id string) (*domain.Node, error)

	// GetByHostPort returns the node matching the given hostname and port.
	// Returns domain.ErrNodeNotFound if no match.
	GetByHostPort(ctx context.Context, hostname string, port int) (*domain.Node, error)

	// Save persists a node (create or update). The node's ID field must
	// be set by the caller.
	Save(ctx context.Context, node *domain.Node) error

	// Delete removes a node by its stable ID.
	// Returns domain.ErrNodeNotFound if not found.
	Delete(ctx context.Context, id string) error
}
