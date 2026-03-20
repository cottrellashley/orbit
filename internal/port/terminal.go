package port

import (
	"context"
	"io"

	"github.com/cottrellashley/orbit/internal/domain"
)

// TerminalManager manages PTY-backed terminal sessions. It owns the
// lifecycle of each terminal subprocess and provides I/O access for
// bridging to WebSocket connections.
type TerminalManager interface {
	// Spawn creates a new PTY-backed terminal running the given command.
	// Returns the terminal's metadata.
	Spawn(ctx context.Context, opts domain.TerminalSpawnOpts) (*domain.Terminal, error)

	// List returns snapshots of all terminals (running and recently exited).
	List(ctx context.Context) []domain.Terminal

	// Get returns a single terminal snapshot by ID.
	Get(ctx context.Context, id string) (*domain.Terminal, error)

	// Kill terminates a running terminal.
	Kill(ctx context.Context, id string) error

	// Attach returns a bidirectional I/O connection to the terminal's PTY.
	// The caller owns the returned TerminalConn and must close it when done.
	Attach(ctx context.Context, id string) (TerminalConn, error)
}

// TerminalConn provides bidirectional access to a terminal's PTY.
// It combines io.ReadWriteCloser with resize capability.
type TerminalConn interface {
	io.ReadWriteCloser

	// Resize changes the PTY window size.
	Resize(cols, rows uint16) error
}
