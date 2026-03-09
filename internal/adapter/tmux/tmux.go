// Package tmux implements port.WorkspaceManager using the tmux terminal
// multiplexer. All operations shell out to the tmux binary. The Runner
// interface allows injecting a mock for testing.
package tmux

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Runner executes tmux commands. The default implementation shells out
// to the tmux binary. Tests can inject a mock.
type Runner interface {
	// Run executes a tmux command and captures its output.
	Run(args ...string) (string, error)
	// RunTTY executes a tmux command with the terminal attached.
	// Used for interactive commands like attach-session and switch-client.
	RunTTY(args ...string) error
}

// Manager implements port.WorkspaceManager for tmux.
type Manager struct {
	session string
	runner  Runner
}

// ManagerOption configures a Manager.
type ManagerOption func(*Manager)

// WithRunner replaces the default exec-based runner (for testing).
func WithRunner(r Runner) ManagerOption {
	return func(m *Manager) { m.runner = r }
}

// NewManager creates a WorkspaceManager backed by tmux for the given
// session name.
func NewManager(session string, opts ...ManagerOption) *Manager {
	m := &Manager{
		session: session,
		runner:  &execRunner{},
	}
	for _, o := range opts {
		o(m)
	}
	return m
}

// ---------------------------------------------------------------------------
// port.WorkspaceManager implementation
// ---------------------------------------------------------------------------

// EnsureSession creates the tmux session if it doesn't already exist.
// If startDir is non-empty, the initial window starts in that directory.
func (m *Manager) EnsureSession(startDir string) error {
	exists, err := m.HasSession()
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	args := []string{"new-session", "-d", "-s", m.session}
	if startDir != "" {
		args = append(args, "-c", startDir)
	}
	_, err = m.runner.Run(args...)
	return err
}

// HasSession returns true if the tmux session exists.
func (m *Manager) HasSession() (bool, error) {
	_, err := m.runner.Run("has-session", "-t", m.session)
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// NewWindow creates a new window in the session.
func (m *Manager) NewWindow(windowName, dir string) error {
	args := []string{"new-window", "-t", m.session, "-n", windowName}
	if dir != "" {
		args = append(args, "-c", dir)
	}
	_, err := m.runner.Run(args...)
	return err
}

// SendKeys sends keystrokes to a target window.
func (m *Manager) SendKeys(target, keys string) error {
	fullTarget := fmt.Sprintf("%s:%s", m.session, target)
	_, err := m.runner.Run("send-keys", "-t", fullTarget, keys, "Enter")
	return err
}

// SelectWindow switches to a window by name or index.
func (m *Manager) SelectWindow(target string) error {
	fullTarget := fmt.Sprintf("%s:%s", m.session, target)
	_, err := m.runner.Run("select-window", "-t", fullTarget)
	return err
}

// ListWindows returns window names in the session.
func (m *Manager) ListWindows() ([]string, error) {
	output, err := m.runner.Run("list-windows", "-t", m.session, "-F", "#{window_name}")
	if err != nil {
		return nil, err
	}
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, nil
	}
	return strings.Split(output, "\n"), nil
}

// KillWindow closes a window by name.
func (m *Manager) KillWindow(target string) error {
	fullTarget := fmt.Sprintf("%s:%s", m.session, target)
	_, err := m.runner.Run("kill-window", "-t", fullTarget)
	return err
}

// AttachOrSwitch attaches to the session if not inside tmux, or
// switches the client if already inside tmux.
func (m *Manager) AttachOrSwitch() error {
	if m.IsInsideSession() {
		return m.runner.RunTTY("switch-client", "-t", m.session)
	}
	return m.runner.RunTTY("attach-session", "-t", m.session)
}

// IsInsideSession reports whether the current process is running inside tmux.
func (m *Manager) IsInsideSession() bool {
	return os.Getenv("TMUX") != ""
}

// ---------------------------------------------------------------------------
// Default runner — shells out to tmux binary
// ---------------------------------------------------------------------------

type execRunner struct{}

func (r *execRunner) Run(args ...string) (string, error) {
	cmd := exec.Command("tmux", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("tmux %s: %w (output: %s)",
			strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

func (r *execRunner) RunTTY(args ...string) error {
	cmd := exec.Command("tmux", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tmux %s: %w", strings.Join(args, " "), err)
	}
	return nil
}
