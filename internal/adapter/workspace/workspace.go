// Package workspace implements port.ConfigWorkspace, providing the
// canonical directory layout for the Orbit managed configuration
// workspace.
//
// The workspace root defaults to ~/.config/orbit but respects the
// XDG_CONFIG_HOME environment variable when set. All subdirectory
// paths are resolved relative to this root.
//
// Directory creation is performed lazily via EnsureDirs — the adapter
// never writes to disk during construction.
package workspace

import (
	"fmt"
	"os"
	"path/filepath"
)

// subdirectories defined by the Orbit config workspace layout.
var subdirs = []string{
	"profiles",
	"skills",
	"agents",
	"plans",
	"mcp",
	"commands",
	"state",
}

// Layout implements port.ConfigWorkspace.
type Layout struct {
	root string
}

// New creates a Layout rooted at the given directory.
// It does not create any directories — call EnsureDirs for that.
func New(root string) *Layout {
	return &Layout{root: root}
}

// DefaultRoot returns the default Orbit config directory.
// It respects XDG_CONFIG_HOME when set, otherwise falls back to
// ~/.config/orbit.
func DefaultRoot() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "orbit")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "orbit")
}

// Root returns the absolute path to the workspace root directory.
func (l *Layout) Root() string { return l.root }

// ProfilesDir returns the path to the profiles subdirectory.
func (l *Layout) ProfilesDir() string { return filepath.Join(l.root, "profiles") }

// SkillsDir returns the path to the skills subdirectory.
func (l *Layout) SkillsDir() string { return filepath.Join(l.root, "skills") }

// AgentsDir returns the path to the agents subdirectory.
func (l *Layout) AgentsDir() string { return filepath.Join(l.root, "agents") }

// PlansDir returns the path to the plans subdirectory.
func (l *Layout) PlansDir() string { return filepath.Join(l.root, "plans") }

// MCPDir returns the path to the MCP server configurations subdirectory.
func (l *Layout) MCPDir() string { return filepath.Join(l.root, "mcp") }

// CommandsDir returns the path to the custom slash commands subdirectory.
func (l *Layout) CommandsDir() string { return filepath.Join(l.root, "commands") }

// StateDir returns the path to the runtime state subdirectory.
func (l *Layout) StateDir() string { return filepath.Join(l.root, "state") }

// EnsureDirs creates the workspace root and all subdirectories if they
// do not already exist. Permissions are set to 0755.
func (l *Layout) EnsureDirs() error {
	for _, sub := range subdirs {
		dir := filepath.Join(l.root, sub)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("workspace: create %s: %w", dir, err)
		}
	}
	return nil
}
