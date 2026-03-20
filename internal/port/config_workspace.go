package port

// ConfigWorkspace provides the directory layout and policies for the
// Orbit managed configuration workspace (e.g. ~/.config/orbit/).
//
// This is a read-only contract — it resolves paths and checks existence
// but does not create directories. Directory creation is the adapter's
// or composition root's responsibility.
type ConfigWorkspace interface {
	// Root returns the absolute path to the workspace root directory.
	Root() string

	// ProfilesDir returns the path to the profiles subdirectory.
	ProfilesDir() string

	// SkillsDir returns the path to the skills subdirectory.
	SkillsDir() string

	// AgentsDir returns the path to the agents subdirectory.
	AgentsDir() string

	// PlansDir returns the path to the plans subdirectory.
	PlansDir() string

	// MCPDir returns the path to the MCP server configurations subdirectory.
	MCPDir() string

	// CommandsDir returns the path to the custom slash commands subdirectory.
	CommandsDir() string

	// StateDir returns the path to the runtime state subdirectory.
	StateDir() string

	// EnsureDirs creates the workspace directory tree if any directories
	// are missing. Returns nil if everything already exists.
	EnsureDirs() error
}
