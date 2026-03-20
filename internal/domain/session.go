package domain

import "time"

// Session is Orbit's enriched view of a coding session, combining
// data from the session provider with Orbit's environment registry.
// The domain does not know which provider (OpenCode, etc.) backs the session.
type Session struct {
	ID              string    // provider-assigned session ID
	Title           string    // session title
	NodeID          string    // stable node ID that hosts this session (empty during migration)
	EnvironmentName string    // matched Orbit environment name (empty if unmatched)
	EnvironmentPath string    // the environment path that matched (empty if unmatched)
	ServerDir       string    // the server's working directory (deprecated: use NodeID)
	ServerPort      int       // the port of the server hosting this session (deprecated: use NodeID)
	Status          string    // "idle", "busy", "retry", or "unknown"
	CreatedAt       time.Time // session creation time
	UpdatedAt       time.Time // session last update time
}

// Server is a summary of a discovered coding-agent server.
type Server struct {
	PID       int
	Port      int
	Hostname  string
	Directory string
	Version   string
	Healthy   bool
}

// ManagedServer is the persistent state of a server that Orbit launched
// and is responsible for managing. This is what gets written to the state
// file so Orbit can reconnect across restarts.
type ManagedServer struct {
	PID       int
	Port      int
	Hostname  string
	Password  string
	Directory string
	Version   string
	StartedAt time.Time
}
