package domain

import "time"

// CopilotTaskStatus represents the state of a Copilot coding agent task.
type CopilotTaskStatus string

const (
	// CopilotTaskRunning means the agent is actively working.
	CopilotTaskRunning CopilotTaskStatus = "running"
	// CopilotTaskCompleted means the agent finished and the PR is ready for review.
	CopilotTaskCompleted CopilotTaskStatus = "completed"
	// CopilotTaskStopped means the task was manually stopped.
	CopilotTaskStopped CopilotTaskStatus = "stopped"
	// CopilotTaskFailed means the agent encountered an unrecoverable error.
	CopilotTaskFailed CopilotTaskStatus = "failed"
	// CopilotTaskUnknown is the zero value — status could not be determined.
	CopilotTaskUnknown CopilotTaskStatus = "unknown"
)

// String returns the string representation of the status.
func (s CopilotTaskStatus) String() string { return string(s) }

// ParseCopilotTaskStatus converts a stored string to a CopilotTaskStatus.
func ParseCopilotTaskStatus(s string) CopilotTaskStatus {
	switch s {
	case "running", "in_progress":
		return CopilotTaskRunning
	case "completed":
		return CopilotTaskCompleted
	case "stopped":
		return CopilotTaskStopped
	case "failed":
		return CopilotTaskFailed
	default:
		return CopilotTaskUnknown
	}
}

// CopilotTask represents a single Copilot coding agent task (session).
// Each task corresponds to a PR created by the Copilot coding agent.
// This is intentionally separate from Node/Session — Copilot tasks are
// cloud-based, have no local process, and their lifecycle is PR-centric.
type CopilotTask struct {
	// ID is the Copilot agent session UUID (primary key).
	ID string
	// Owner is the GitHub org or user that owns the repository.
	Owner string
	// Repo is the repository name (without owner prefix).
	Repo string
	// PRNumber is the pull request number created by this task (0 if pending).
	PRNumber int
	// Title is the PR title or a summary of the task prompt.
	Title string
	// Prompt is the original instruction given to the agent.
	Prompt string
	// Status is the current state of the agent task.
	Status CopilotTaskStatus
	// HTMLURL is a direct link to the PR on GitHub.
	HTMLURL string
	// Branch is the git branch the agent is working on.
	Branch string
	// Draft indicates whether the PR is still a draft.
	Draft bool
	// CreatedAt is when the task was initiated.
	CreatedAt time.Time
	// UpdatedAt is when the task status was last refreshed.
	UpdatedAt time.Time
}

// CopilotTaskCreateOpts holds options for creating a new Copilot agent task.
type CopilotTaskCreateOpts struct {
	// Owner is the GitHub org or user (required).
	Owner string
	// Repo is the repository name (required).
	Repo string
	// Prompt is the instruction for the agent (required).
	Prompt string
	// BaseBranch is the branch to work from (empty = repo default).
	BaseBranch string
	// CustomInstructions provides additional context (optional).
	CustomInstructions string
}
