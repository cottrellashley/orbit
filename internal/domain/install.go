package domain

// InstallStatus represents whether a tool is installed.
type InstallStatus string

const (
	InstallStatusInstalled    InstallStatus = "installed"
	InstallStatusNotInstalled InstallStatus = "not_installed"
	InstallStatusUnknown      InstallStatus = "unknown"
)

// String returns the string representation of the status.
func (s InstallStatus) String() string { return string(s) }

// ToolInfo describes a single installable tool and its current state.
type ToolInfo struct {
	Name        string        // canonical name, e.g. "opencode", "gh", "uv"
	Description string        // one-line summary
	Status      InstallStatus // current install state
	Version     string        // version string (empty if not installed)
}

// InstallResult is returned after an install attempt.
type InstallResult struct {
	Name    string // tool name
	Success bool   // true if install succeeded
	Version string // version after install (empty on failure)
	Error   string // human-readable error (empty on success)
}
