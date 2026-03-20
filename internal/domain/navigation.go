package domain

// JumpLink represents a navigable outbound URL that the UI can open.
// It is the standard return type for all navigation helpers.
type JumpLink struct {
	// Label is a short human-readable description, e.g. "GitHub Repo".
	Label string
	// URL is the fully-qualified URL to open.
	URL string
	// Kind classifies the link for UI grouping / icon selection.
	Kind LinkKind
}

// LinkKind categorises a JumpLink so drivers can render appropriate icons
// or group links by type.
type LinkKind string

const (
	LinkRepo    LinkKind = "repo"
	LinkIssue   LinkKind = "issue"
	LinkPR      LinkKind = "pull-request"
	LinkBranch  LinkKind = "branch"
	LinkSession LinkKind = "session"
	LinkWeb     LinkKind = "web"
)
