package domain

// Profile represents a reusable starter kit for orbit environments.
// A profile is a directory containing configuration files (opencode.json,
// AGENTS.md, skills, etc.) plus metadata.
type Profile struct {
	Name        string // directory name, used as identifier
	Description string // human-readable description
	Path        string // absolute path to the profile directory
	Notes       string // optional notes / instructions
}
