package app

import (
	"context"

	"github.com/cottrellashley/orbit/internal/domain"
	"github.com/cottrellashley/orbit/internal/port"
)

// TerminalService orchestrates PTY-backed terminal sessions. It is a
// thin app service that delegates to port.TerminalManager.
type TerminalService struct {
	mgr port.TerminalManager
}

// NewTerminalService creates a TerminalService.
func NewTerminalService(mgr port.TerminalManager) *TerminalService {
	return &TerminalService{mgr: mgr}
}

// Spawn creates a new PTY-backed terminal.
func (s *TerminalService) Spawn(ctx context.Context, opts domain.TerminalSpawnOpts) (*domain.Terminal, error) {
	return s.mgr.Spawn(ctx, opts)
}

// List returns all terminals.
func (s *TerminalService) List(ctx context.Context) []domain.Terminal {
	return s.mgr.List(ctx)
}

// Get returns a single terminal by ID.
func (s *TerminalService) Get(ctx context.Context, id string) (*domain.Terminal, error) {
	return s.mgr.Get(ctx, id)
}

// Kill terminates a running terminal.
func (s *TerminalService) Kill(ctx context.Context, id string) error {
	return s.mgr.Kill(ctx, id)
}

// Attach returns a bidirectional connection to the terminal's PTY.
func (s *TerminalService) Attach(ctx context.Context, id string) (port.TerminalConn, error) {
	return s.mgr.Attach(ctx, id)
}
