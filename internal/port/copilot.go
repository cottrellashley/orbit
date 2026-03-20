package port

import (
	"context"
	"io"

	"github.com/cottrellashley/orbit/internal/domain"
)

// CopilotTaskProvider discovers and manages Copilot coding agent tasks.
// This is the adapter boundary — the domain does not know whether the
// backing implementation uses the gh CLI, the GitHub API, or something else.
type CopilotTaskProvider interface {
	// ListTasks returns Copilot agent tasks visible to the authenticated user.
	// If owner is non-empty, results are scoped to that org or user.
	// If repo is also non-empty, results are scoped to that specific repository.
	ListTasks(ctx context.Context, owner, repo string) ([]domain.CopilotTask, error)

	// GetTask returns a single Copilot agent task by session ID (UUID).
	GetTask(ctx context.Context, sessionID string) (*domain.CopilotTask, error)

	// CreateTask starts a new Copilot coding agent task.
	CreateTask(ctx context.Context, opts domain.CopilotTaskCreateOpts) (*domain.CopilotTask, error)

	// StopTask stops a running Copilot agent task by session ID.
	StopTask(ctx context.Context, sessionID string) error

	// TaskLogs returns the session log for a task as a stream.
	// The caller is responsible for closing the returned reader.
	TaskLogs(ctx context.Context, sessionID string) (io.ReadCloser, error)

	// IsAvailable reports whether the Copilot agent-task tooling is usable
	// (e.g. gh CLI v2.80.0+ with agent-task support).
	IsAvailable() bool
}
