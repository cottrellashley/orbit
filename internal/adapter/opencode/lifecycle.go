package opencode

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
	"github.com/cottrellashley/orbit/internal/port"
	"github.com/sst/opencode-sdk-go"
)

// Compile-time check: ServerManager satisfies port.ServerLifecycle.
var _ port.ServerLifecycle = (*ServerManager)(nil)

// DefaultStopTimeout is the default timeout for stopping a managed server.
const DefaultStopTimeout = 15 * time.Second

// ---------------------------------------------------------------------------
// ServerManager — implements port.ServerLifecycle for OpenCode
// ---------------------------------------------------------------------------

// ServerManagerOpts configures the ServerManager.
type ServerManagerOpts struct {
	// ProcessOpts is forwarded to StartServer when launching a new process.
	ProcessOpts ProcessOpts

	// StatePath overrides the default state file location.
	// Defaults to ~/.local/state/orbit/servers.json.
	StatePath string
}

// ServerManager manages the lifecycle of a single OpenCode server process.
// It implements port.ServerLifecycle.
type ServerManager struct {
	mu        sync.Mutex
	opts      ServerManagerOpts
	statePath string
	proc      *Process // nil when no managed server is running in-process
}

// NewServerManager creates a ServerManager.
func NewServerManager(opts ServerManagerOpts) *ServerManager {
	sp := opts.StatePath
	if sp == "" {
		sp = defaultStatePath()
	}
	return &ServerManager{
		opts:      opts,
		statePath: sp,
	}
}

// Start launches a managed OpenCode server. If a healthy server from
// a previous Orbit session is still running (found in state file),
// it is reused. Otherwise a new process is started.
func (m *ServerManager) Start(ctx context.Context) (*domain.ManagedServer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If we already have a live in-process handle, return it.
	if m.proc != nil && m.proc.Running() {
		return m.managedServerInfo(ctx)
	}

	// Reap stale entries and check for a surviving server from a prior run.
	sf, err := reapStale(m.statePath)
	if err != nil {
		sf = &stateFile{}
	}

	if existing := findServer(sf); existing != nil {
		// Verify it's actually healthy before trusting the state file.
		if m.probeExisting(ctx, existing) {
			return existing, nil
		}
		// Stale — remove it.
		_ = removeServer(m.statePath, existing.PID)
	}

	// Launch a new server.
	proc, err := StartServer(ctx, m.opts.ProcessOpts)
	if err != nil {
		return nil, fmt.Errorf("start managed server: %w", err)
	}
	m.proc = proc

	// Determine version via health check (already healthy at this point).
	version := m.probeVersion(ctx, proc)

	ms := domain.ManagedServer{
		PID:       proc.cmd.Process.Pid,
		Port:      proc.Port(),
		Hostname:  m.hostname(),
		Password:  m.opts.ProcessOpts.Password,
		Directory: m.opts.ProcessOpts.Directory,
		Version:   version,
		StartedAt: nowUTC(),
	}

	if err := addServer(m.statePath, ms); err != nil {
		// Non-fatal: server is running, just can't persist state.
		// Log would go here if we had a logger.
	}

	return &ms, nil
}

// Stop gracefully shuts down the managed server.
func (m *ServerManager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.proc != nil {
		pid := 0
		if m.proc.cmd.Process != nil {
			pid = m.proc.cmd.Process.Pid
		}

		if err := m.proc.Stop(ctx); err != nil {
			return fmt.Errorf("stop managed server: %w", err)
		}

		if pid > 0 {
			_ = removeServer(m.statePath, pid)
		}
		m.proc = nil
		return nil
	}

	// No in-process handle. Check state file for orphaned servers we own.
	sf, _ := readState(m.statePath)
	if existing := findServer(sf); existing != nil {
		// Best-effort kill via signal.
		if processAlive(existing.PID) {
			proc, err := connectToProcess(ctx, *existing, m.opts.ProcessOpts)
			if err == nil {
				_ = proc.Stop(ctx)
			}
		}
		_ = removeServer(m.statePath, existing.PID)
	}

	return nil
}

// Status returns the current managed server info, or nil if none is running.
func (m *ServerManager) Status(ctx context.Context) (*domain.ManagedServer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check in-process handle first.
	if m.proc != nil && m.proc.Running() {
		return m.managedServerInfo(ctx)
	}

	// Fall back to state file.
	sf, err := reapStale(m.statePath)
	if err != nil {
		return nil, nil
	}

	existing := findServer(sf)
	if existing == nil {
		return nil, nil
	}

	// Verify health.
	if !m.probeExisting(ctx, existing) {
		_ = removeServer(m.statePath, existing.PID)
		return nil, nil
	}

	return existing, nil
}

// Server returns a domain.Server for the managed process, suitable for
// passing to SessionProvider methods. Returns nil if not running.
func (m *ServerManager) Server(ctx context.Context) *domain.Server {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.proc != nil && m.proc.Running() {
		return &domain.Server{
			PID:       m.proc.cmd.Process.Pid,
			Port:      m.proc.Port(),
			Hostname:  m.hostname(),
			Directory: m.opts.ProcessOpts.Directory,
			Healthy:   true,
		}
	}

	// Check state file (unlocked read is fine — we hold mu).
	sf, _ := readState(m.statePath)
	if existing := findServer(sf); existing != nil {
		if processAlive(existing.PID) {
			return &domain.Server{
				PID:       existing.PID,
				Port:      existing.Port,
				Hostname:  existing.Hostname,
				Directory: existing.Directory,
				Version:   existing.Version,
				Healthy:   true,
			}
		}
	}

	return nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func (m *ServerManager) hostname() string {
	h := m.opts.ProcessOpts.Hostname
	if h == "" {
		return "127.0.0.1"
	}
	return h
}

// managedServerInfo builds a ManagedServer from the in-process handle.
func (m *ServerManager) managedServerInfo(ctx context.Context) (*domain.ManagedServer, error) {
	if m.proc == nil {
		return nil, fmt.Errorf("%w: no managed process", ErrNotRunning)
	}

	version := m.probeVersion(ctx, m.proc)

	return &domain.ManagedServer{
		PID:       m.proc.cmd.Process.Pid,
		Port:      m.proc.Port(),
		Hostname:  m.hostname(),
		Password:  m.opts.ProcessOpts.Password,
		Directory: m.opts.ProcessOpts.Directory,
		Version:   version,
		StartedAt: nowUTC(),
	}, nil
}

// probeExisting health-checks a server from the state file.
func (m *ServerManager) probeExisting(ctx context.Context, ms *domain.ManagedServer) bool {
	if !processAlive(ms.PID) {
		return false
	}

	srv := discoveredServer{
		PID:      ms.PID,
		Hostname: ms.Hostname,
		Port:     ms.Port,
	}
	cand := serverCandidate{
		pid:      srv.PID,
		hostname: srv.Hostname,
		port:     srv.Port,
	}

	result := probeCandidate(ctx, cand, m.opts.ProcessOpts.StartTimeout, nil)
	return result.Healthy
}

// probeVersion extracts the version string from a running server's
// health endpoint. Returns empty string on failure.
func (m *ServerManager) probeVersion(ctx context.Context, proc *Process) string {
	var h healthResponse
	err := proc.Client().Get(ctx, "/global/health", nil, &h)
	if err != nil {
		return ""
	}
	return h.Version
}

// connectToProcess creates a Process handle for an existing server
// described by a ManagedServer entry. Used to stop orphaned servers.
// The returned Process has no exec.Cmd — only Stop via dispose + signal
// is supported.
func connectToProcess(_ context.Context, ms domain.ManagedServer, popts ProcessOpts) (*Process, error) {
	hostname := ms.Hostname
	if hostname == "" {
		hostname = "127.0.0.1"
	}

	baseURL := fmt.Sprintf("http://%s:%d", hostname, ms.Port)
	opts := ProcessOpts{
		Hostname: hostname,
		Password: ms.Password,
		Username: popts.Username,
	}
	clientOpts := buildClientOpts(baseURL, opts)

	osProcHandle, err := os.FindProcess(ms.PID)
	if err != nil {
		return nil, fmt.Errorf("find process %d: %w", ms.PID, err)
	}

	done := make(chan struct{})
	go func() {
		// Best-effort wait. On Unix, Wait only works for child processes,
		// so this may return immediately with an error. That's fine —
		// we rely on processAlive checks elsewhere.
		_, _ = osProcHandle.Wait()
		close(done)
	}()

	// Build a minimal Process. The cmd field is nil (we didn't start it),
	// so we set up a fake exec.Cmd that wraps the os.Process for Stop().
	fakeCmd := &exec.Cmd{}
	fakeCmd.Process = osProcHandle

	return &Process{
		cmd:     fakeCmd,
		client:  opencode.NewClient(clientOpts...),
		baseURL: baseURL,
		port:    ms.Port,
		done:    done,
	}, nil
}
