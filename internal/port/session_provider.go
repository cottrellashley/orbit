package port

import (
	"context"

	"github.com/cottrellashley/orbit/internal/domain"
)

// SessionProvider discovers running coding-agent servers and manages
// their sessions. This is the adapter boundary — the domain does not
// know whether the provider is OpenCode, Cursor, or anything else.
type SessionProvider interface {
	// DiscoverServers scans for running coding-agent servers.
	DiscoverServers(ctx context.Context) ([]domain.Server, error)

	// ListSessions returns all sessions from a specific node.
	ListSessions(ctx context.Context, node domain.Node) ([]domain.Session, error)

	// GetSession fetches a single session by ID from a specific node.
	GetSession(ctx context.Context, node domain.Node, sessionID string) (*domain.Session, error)

	// CreateSession creates a new session on a specific node.
	CreateSession(ctx context.Context, node domain.Node, title string) (*domain.Session, error)

	// AbortSession stops a running session on the given node.
	AbortSession(ctx context.Context, node domain.Node, sessionID string) error

	// DeleteSession removes a session from the given node.
	DeleteSession(ctx context.Context, node domain.Node, sessionID string) error

	// IsInstalled reports whether the coding agent binary is available.
	IsInstalled() bool

	// Version returns the coding agent's version string.
	Version(ctx context.Context) (string, error)
}
