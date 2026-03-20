package domain

// ---------------------------------------------------------------------------
// Terminal domain types
//
// These model a PTY-backed terminal session managed by Orbit. Each
// terminal runs a subprocess (e.g. "opencode attach http://host:port")
// inside a pseudo-terminal, allowing the web UI to render the native TUI.
// ---------------------------------------------------------------------------

// Terminal represents a running PTY-backed terminal session.
type Terminal struct {
	// ID is the unique identifier for this terminal.
	ID string
	// Command is the executable that was spawned (e.g. "opencode").
	Command string
	// Args are the arguments passed to the command.
	Args []string
	// Cols is the current terminal width in columns.
	Cols uint16
	// Rows is the current terminal height in rows.
	Rows uint16
	// Running indicates whether the subprocess is still alive.
	Running bool
}

// TerminalSpawnOpts describes how to spawn a new PTY-backed terminal.
type TerminalSpawnOpts struct {
	// Command is the executable to run (e.g. "opencode").
	Command string
	// Args are the command-line arguments (e.g. ["attach", "http://host:port"]).
	Args []string
	// Cols is the initial terminal width in columns.
	Cols uint16
	// Rows is the initial terminal height in rows.
	Rows uint16
}
