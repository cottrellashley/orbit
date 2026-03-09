package domain

import "time"

// Environment represents a registered orbit environment — an external
// filesystem path that was optionally initialized from a profile.
// After creation, the environment contents are independent of Orbit;
// Orbit only keeps this registry entry.
type Environment struct {
	Name        string
	Path        string // absolute filesystem path
	ProfileName string // profile used at creation (may be empty)
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
