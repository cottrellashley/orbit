package port

// WorkspaceManager orchestrates terminal multiplexer sessions and windows.
// The domain does not know whether this is tmux, Zellij, or something else.
type WorkspaceManager interface {
	// EnsureSession creates the multiplexer session if it doesn't exist.
	EnsureSession(startDir string) error

	// HasSession returns true if the multiplexer session exists.
	HasSession() (bool, error)

	// NewWindow creates a new window in the session.
	NewWindow(windowName, dir string) error

	// SendKeys sends keystrokes to a target window.
	SendKeys(target, keys string) error

	// SelectWindow switches to a window by name or index.
	SelectWindow(target string) error

	// ListWindows returns window names in the session.
	ListWindows() ([]string, error)

	// KillWindow closes a window by name.
	KillWindow(target string) error

	// AttachOrSwitch attaches to or switches to the session.
	AttachOrSwitch() error

	// IsInsideSession reports whether the current process is inside
	// the multiplexer already.
	IsInsideSession() bool
}
