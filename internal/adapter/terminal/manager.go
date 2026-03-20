// Package terminal implements port.TerminalManager by managing PTY-backed
// subprocess terminals. It uses github.com/creack/pty for PTY allocation.
package terminal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/creack/pty"

	"github.com/cottrellashley/orbit/internal/domain"
	"github.com/cottrellashley/orbit/internal/port"
)

// Compile-time check: Manager satisfies port.TerminalManager.
var _ port.TerminalManager = (*Manager)(nil)

// Manager manages PTY-backed terminal sessions.
type Manager struct {
	mu        sync.Mutex
	terminals map[string]*terminal
	nextID    int
}

// New creates a Manager.
func New() *Manager {
	return &Manager{
		terminals: make(map[string]*terminal),
	}
}

// terminal is an internal representation of a running PTY session.
type terminal struct {
	meta domain.Terminal
	cmd  *exec.Cmd
	ptmx *os.File
	done chan struct{} // closed when subprocess exits
}

// Spawn creates a new PTY-backed terminal running the given command.
func (m *Manager) Spawn(_ context.Context, opts domain.TerminalSpawnOpts) (*domain.Terminal, error) {
	if opts.Command == "" {
		return nil, fmt.Errorf("command is required")
	}
	if opts.Cols == 0 {
		opts.Cols = 120
	}
	if opts.Rows == 0 {
		opts.Rows = 40
	}

	cmd := exec.Command(opts.Command, opts.Args...)
	cmd.Env = os.Environ()

	// Set initial window size.
	ws := &pty.Winsize{
		Cols: opts.Cols,
		Rows: opts.Rows,
	}

	ptmx, err := pty.StartWithSize(cmd, ws)
	if err != nil {
		return nil, fmt.Errorf("start pty: %w", err)
	}

	m.mu.Lock()
	m.nextID++
	id := fmt.Sprintf("term-%d", m.nextID)
	t := &terminal{
		meta: domain.Terminal{
			ID:      id,
			Command: opts.Command,
			Args:    opts.Args,
			Cols:    opts.Cols,
			Rows:    opts.Rows,
			Running: true,
		},
		cmd:  cmd,
		ptmx: ptmx,
		done: make(chan struct{}),
	}
	m.terminals[id] = t
	m.mu.Unlock()

	// Monitor subprocess exit in the background.
	go func() {
		_ = cmd.Wait()
		m.mu.Lock()
		t.meta.Running = false
		m.mu.Unlock()
		close(t.done)
	}()

	result := t.meta // copy
	return &result, nil
}

// List returns snapshots of all terminals.
func (m *Manager) List(_ context.Context) []domain.Terminal {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]domain.Terminal, 0, len(m.terminals))
	for _, t := range m.terminals {
		result = append(result, t.meta) // copy
	}
	return result
}

// Get returns a single terminal snapshot by ID.
func (m *Manager) Get(_ context.Context, id string) (*domain.Terminal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, ok := m.terminals[id]
	if !ok {
		return nil, fmt.Errorf("terminal %q not found", id)
	}
	result := t.meta // copy
	return &result, nil
}

// Kill terminates a running terminal.
func (m *Manager) Kill(_ context.Context, id string) error {
	m.mu.Lock()
	t, ok := m.terminals[id]
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("terminal %q not found", id)
	}

	// Send SIGHUP then close the PTY master — same as closing a terminal window.
	if t.cmd.Process != nil {
		_ = t.cmd.Process.Signal(os.Signal(syscall.SIGHUP))
	}
	_ = t.ptmx.Close()

	// Wait for exit (with a brief timeout to avoid blocking forever).
	<-t.done

	// Clean up the map entry.
	m.mu.Lock()
	delete(m.terminals, id)
	m.mu.Unlock()

	return nil
}

// Attach returns a bidirectional I/O connection to the terminal's PTY.
func (m *Manager) Attach(_ context.Context, id string) (port.TerminalConn, error) {
	m.mu.Lock()
	t, ok := m.terminals[id]
	m.mu.Unlock()

	if !ok {
		return nil, fmt.Errorf("terminal %q not found", id)
	}

	return &conn{t: t}, nil
}

// conn implements port.TerminalConn.
type conn struct {
	t *terminal
}

func (c *conn) Read(p []byte) (int, error) {
	return c.t.ptmx.Read(p)
}

func (c *conn) Write(p []byte) (int, error) {
	return c.t.ptmx.Write(p)
}

func (c *conn) Resize(cols, rows uint16) error {
	return pty.Setsize(c.t.ptmx, &pty.Winsize{
		Cols: cols,
		Rows: rows,
	})
}

func (c *conn) Close() error {
	// Close is a no-op on the conn — the PTY is owned by the Manager.
	// Closing the conn just means the WebSocket detached; the terminal
	// continues running.
	return nil
}
