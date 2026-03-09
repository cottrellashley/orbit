package port

import (
	"context"

	"github.com/cottrellashley/orbit/internal/domain"
)

// ServerLifecycle manages the lifecycle of a coding-agent server process.
// Orbit uses this to start a server on boot, monitor its health, and
// stop it on shutdown. The domain does not know which agent backs this.
type ServerLifecycle interface {
	// Start launches a managed server process and waits for it to become
	// healthy. It persists the server's connection details to the state
	// file. If a healthy managed server already exists, Start returns it
	// without launching a new one.
	Start(ctx context.Context) (*domain.ManagedServer, error)

	// Stop gracefully shuts down the managed server and removes its
	// state entry. It is safe to call Stop if no server is running.
	Stop(ctx context.Context) error

	// Status returns the current managed server info, or nil if no
	// managed server is running. It health-checks the process before
	// returning.
	Status(ctx context.Context) (*domain.ManagedServer, error)

	// Server returns the domain.Server for the managed process, suitable
	// for passing to SessionProvider methods. Returns nil if not running.
	Server(ctx context.Context) *domain.Server
}
