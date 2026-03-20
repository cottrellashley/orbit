package opencode

import (
	"context"
	"fmt"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
	"github.com/sst/opencode-sdk-go"
	"github.com/sst/opencode-sdk-go/option"
)

// Adapter implements port.SessionProvider by wrapping the official OpenCode
// Go SDK. It translates between SDK types and domain types.
type Adapter struct {
	binary       string
	probeTimeout time.Duration
	clientOpts   []option.RequestOption
}

// AdapterOption configures an Adapter.
type AdapterOption func(*Adapter)

// WithBinary sets the opencode binary name (default "opencode").
func WithBinary(binary string) AdapterOption {
	return func(a *Adapter) { a.binary = binary }
}

// WithProbeTimeout sets the timeout for health-probing discovered servers.
func WithProbeTimeout(d time.Duration) AdapterOption {
	return func(a *Adapter) { a.probeTimeout = d }
}

// WithClientOpts passes additional SDK request options to clients
// created during discovery and per-server operations.
func WithClientOpts(opts ...option.RequestOption) AdapterOption {
	return func(a *Adapter) { a.clientOpts = append(a.clientOpts, opts...) }
}

// NewAdapter creates a SessionProvider backed by the OpenCode SDK.
func NewAdapter(opts ...AdapterOption) *Adapter {
	a := &Adapter{
		binary:       "opencode",
		probeTimeout: 2 * time.Second,
	}
	for _, o := range opts {
		o(a)
	}
	return a
}

// ---------------------------------------------------------------------------
// port.SessionProvider implementation
// ---------------------------------------------------------------------------

// DiscoverServers scans the process table for running OpenCode servers and
// probes each candidate for health.
func (a *Adapter) DiscoverServers(ctx context.Context) ([]domain.Server, error) {
	discovered, err := DiscoverServers(ctx, DiscoverOpts{
		Binary:       a.binary,
		ProbeTimeout: a.probeTimeout,
		ClientOpts:   a.clientOpts,
	})
	if err != nil {
		return nil, fmt.Errorf("opencode discover: %w", err)
	}

	servers := make([]domain.Server, len(discovered))
	for i, d := range discovered {
		servers[i] = toDomainServer(d)
	}
	return servers, nil
}

// ListSessions returns all sessions from the given node.
func (a *Adapter) ListSessions(ctx context.Context, node domain.Node) ([]domain.Session, error) {
	client := a.clientFor(node)

	sessions, err := client.Session.List(ctx, opencode.SessionListParams{})
	if err != nil {
		return nil, fmt.Errorf("opencode list sessions: %w", err)
	}

	// Fetch session statuses to enrich domain sessions.
	statuses := a.fetchStatuses(ctx, client)

	result := make([]domain.Session, 0, len(*sessions))
	for _, s := range *sessions {
		// Skip sub-sessions (children of another session).
		if s.ParentID != "" {
			continue
		}
		result = append(result, toDomainSession(s, node, statuses))
	}
	return result, nil
}

// GetSession fetches a single session by ID from the given node.
func (a *Adapter) GetSession(ctx context.Context, node domain.Node, sessionID string) (*domain.Session, error) {
	client := a.clientFor(node)

	s, err := client.Session.Get(ctx, sessionID, opencode.SessionGetParams{})
	if err != nil {
		return nil, fmt.Errorf("opencode get session %s: %w", sessionID, err)
	}

	statuses := a.fetchStatuses(ctx, client)
	ds := toDomainSession(*s, node, statuses)
	return &ds, nil
}

// CreateSession creates a new session on the given node.
func (a *Adapter) CreateSession(ctx context.Context, node domain.Node, title string) (*domain.Session, error) {
	client := a.clientFor(node)

	params := opencode.SessionNewParams{}
	if title != "" {
		params.Title = opencode.F(title)
	}

	s, err := client.Session.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("opencode create session: %w", err)
	}

	ds := toDomainSession(*s, node, nil)
	return &ds, nil
}

// AbortSession stops a running session on the given node.
func (a *Adapter) AbortSession(ctx context.Context, node domain.Node, sessionID string) error {
	client := a.clientFor(node)
	_, err := client.Session.Abort(ctx, sessionID, opencode.SessionAbortParams{})
	if err != nil {
		return fmt.Errorf("opencode abort session %s: %w", sessionID, err)
	}
	return nil
}

// DeleteSession removes a session from the given node.
func (a *Adapter) DeleteSession(ctx context.Context, node domain.Node, sessionID string) error {
	client := a.clientFor(node)
	_, err := client.Session.Delete(ctx, sessionID, opencode.SessionDeleteParams{})
	if err != nil {
		return fmt.Errorf("opencode delete session %s: %w", sessionID, err)
	}
	return nil
}

// IsInstalled reports whether the opencode binary is on PATH.
func (a *Adapter) IsInstalled() bool {
	return IsInstalled(a.binary)
}

// Version returns the opencode binary's version string.
func (a *Adapter) Version(ctx context.Context) (string, error) {
	v, err := Version(ctx, a.binary)
	if err != nil {
		return "", fmt.Errorf("opencode version: %w", err)
	}
	return v, nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// clientFor creates an SDK client targeting a specific node.
func (a *Adapter) clientFor(node domain.Node) *opencode.Client {
	baseURL := fmt.Sprintf("http://%s:%d", node.Hostname, node.Port)
	opts := make([]option.RequestOption, 0, len(a.clientOpts)+1)
	opts = append(opts, option.WithBaseURL(baseURL))
	opts = append(opts, a.clientOpts...)
	return opencode.NewClient(opts...)
}

// sessionStatus is the response shape for GET /session/status, which is
// not yet available in the SDK.
type sessionStatus struct {
	Type    string  `json:"type"`
	Attempt *int    `json:"attempt,omitempty"`
	Message *string `json:"message,omitempty"`
}

// fetchStatuses retrieves session statuses using the SDK's raw Get method.
// This endpoint is not yet in the SDK. Best-effort: returns nil on error.
func (a *Adapter) fetchStatuses(ctx context.Context, client *opencode.Client) map[string]sessionStatus {
	var statuses map[string]sessionStatus
	err := client.Get(ctx, "/session/status", nil, &statuses)
	if err != nil {
		return nil
	}
	return statuses
}

// toDomainServer converts a discoveredServer to a domain.Server.
func toDomainServer(d discoveredServer) domain.Server {
	return domain.Server{
		PID:       d.PID,
		Port:      d.Port,
		Hostname:  d.Hostname,
		Directory: d.Directory,
		Version:   d.Version,
		Healthy:   d.Healthy,
	}
}

// toDomainSession converts an SDK Session to a domain.Session,
// enriching it with node info and optional status data.
func toDomainSession(s opencode.Session, node domain.Node, statuses map[string]sessionStatus) domain.Session {
	ds := domain.Session{
		ID:         s.ID,
		Title:      s.Title,
		NodeID:     node.ID,
		ServerDir:  node.Directory, // deprecated: kept for backward compat
		ServerPort: node.Port,      // deprecated: kept for backward compat
		Status:     "unknown",
		CreatedAt:  time.UnixMilli(int64(s.Time.Created)),
		UpdatedAt:  time.UnixMilli(int64(s.Time.Updated)),
	}

	if st, ok := statuses[s.ID]; ok {
		ds.Status = st.Type
	}

	return ds
}
