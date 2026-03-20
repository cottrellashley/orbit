package domain

import "time"

// NodeProvider identifies the type of AI agent server backing a node.
// The domain defines the enum; adapters implement provider-specific logic.
type NodeProvider int

const (
	// ProviderUnknown is the zero value — provider has not been identified.
	ProviderUnknown NodeProvider = iota
	// ProviderOpenCode indicates the node runs an OpenCode server.
	ProviderOpenCode
)

// String returns a human-readable label for the provider.
func (p NodeProvider) String() string {
	switch p {
	case ProviderOpenCode:
		return "opencode"
	default:
		return "unknown"
	}
}

// ParseNodeProvider converts a stored string back to a NodeProvider.
func ParseNodeProvider(s string) NodeProvider {
	switch s {
	case "opencode":
		return ProviderOpenCode
	default:
		return ProviderUnknown
	}
}

// NodeOrigin describes how a node was added to the registry.
type NodeOrigin int

const (
	// OriginDiscovered means the node was found by automatic local process scanning.
	OriginDiscovered NodeOrigin = iota
	// OriginRegistered means the node was manually registered (e.g. a remote server).
	OriginRegistered
)

// String returns a human-readable label for the origin.
func (o NodeOrigin) String() string {
	switch o {
	case OriginDiscovered:
		return "discovered"
	case OriginRegistered:
		return "registered"
	default:
		return "unknown"
	}
}

// ParseNodeOrigin converts a stored string back to a NodeOrigin.
func ParseNodeOrigin(s string) NodeOrigin {
	switch s {
	case "discovered":
		return OriginDiscovered
	case "registered":
		return OriginRegistered
	default:
		return OriginDiscovered
	}
}

// Node is any AI agent server — local or remote — that Orbit manages.
// It is the central abstraction that replaces the older Server type,
// adding a stable identity, provider awareness, and support for remote
// servers alongside locally-discovered processes.
//
// A Node's ID is a stable UUID assigned at registration time. The
// hostname:port pair may change across restarts; the ID does not.
type Node struct {
	// ID is a stable unique identifier (UUID) assigned at creation.
	ID string
	// Name is an optional user-friendly display name.
	Name string
	// Provider identifies the type of AI agent server (OpenCode, etc.).
	Provider NodeProvider
	// Origin describes how this node was added (discovered vs registered).
	Origin NodeOrigin

	// Hostname is the network address of the server (e.g. "127.0.0.1", "agent.example.com").
	Hostname string
	// Port is the TCP port the server listens on.
	Port int
	// Directory is the working directory of the server process (may be empty for remote nodes).
	Directory string
	// Version is the server software version (may be empty if unknown).
	Version string

	// PID is the OS process ID. Only meaningful for locally-discovered nodes;
	// zero for remote nodes.
	PID int
	// Password is the authentication credential for the server.
	// Must never be logged or included in error messages.
	Password string

	// Healthy is the last-known health status.
	Healthy bool
	// CreatedAt is when the node was first registered.
	CreatedAt time.Time
	// UpdatedAt is when the node metadata was last refreshed.
	UpdatedAt time.Time
}
