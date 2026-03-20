package app

import (
	"context"
	"fmt"
	"io"

	"github.com/cottrellashley/orbit/internal/domain"
)

// copilotTaskProvider is the consumer-defined interface for the Copilot
// coding agent port. It mirrors port.CopilotTaskProvider but is defined
// here so the app layer does not import port types directly (hexagonal
// consumer-defined rule).
type copilotTaskProvider interface {
	ListTasks(ctx context.Context, owner, repo string) ([]domain.CopilotTask, error)
	GetTask(ctx context.Context, sessionID string) (*domain.CopilotTask, error)
	CreateTask(ctx context.Context, opts domain.CopilotTaskCreateOpts) (*domain.CopilotTask, error)
	StopTask(ctx context.Context, sessionID string) error
	TaskLogs(ctx context.Context, sessionID string) (io.ReadCloser, error)
	IsAvailable() bool
}

// CopilotService orchestrates Copilot coding agent operations for drivers.
// It delegates all work to the CopilotTaskProvider port.
type CopilotService struct {
	cp copilotTaskProvider
}

// NewCopilotService creates a CopilotService.
func NewCopilotService(cp copilotTaskProvider) *CopilotService {
	return &CopilotService{cp: cp}
}

// IsAvailable reports whether the Copilot agent-task tooling is usable.
func (s *CopilotService) IsAvailable() bool {
	return s.cp.IsAvailable()
}

// ListTasks returns Copilot agent tasks, optionally scoped by owner and repo.
func (s *CopilotService) ListTasks(ctx context.Context, owner, repo string) ([]domain.CopilotTask, error) {
	if !s.cp.IsAvailable() {
		return nil, domain.ErrCopilotUnavailable
	}
	return s.cp.ListTasks(ctx, owner, repo)
}

// GetTask returns a single Copilot agent task by session ID.
func (s *CopilotService) GetTask(ctx context.Context, sessionID string) (*domain.CopilotTask, error) {
	if !s.cp.IsAvailable() {
		return nil, domain.ErrCopilotUnavailable
	}
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}
	return s.cp.GetTask(ctx, sessionID)
}

// CreateTask starts a new Copilot coding agent task.
func (s *CopilotService) CreateTask(ctx context.Context, opts domain.CopilotTaskCreateOpts) (*domain.CopilotTask, error) {
	if !s.cp.IsAvailable() {
		return nil, domain.ErrCopilotUnavailable
	}
	if opts.Owner == "" || opts.Repo == "" || opts.Prompt == "" {
		return nil, fmt.Errorf("owner, repo, and prompt are required")
	}
	return s.cp.CreateTask(ctx, opts)
}

// StopTask stops a running Copilot agent task by session ID.
func (s *CopilotService) StopTask(ctx context.Context, sessionID string) error {
	if !s.cp.IsAvailable() {
		return domain.ErrCopilotUnavailable
	}
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}
	return s.cp.StopTask(ctx, sessionID)
}

// TaskLogs returns the session log stream for a task.
// The caller is responsible for closing the returned reader.
func (s *CopilotService) TaskLogs(ctx context.Context, sessionID string) (io.ReadCloser, error) {
	if !s.cp.IsAvailable() {
		return nil, domain.ErrCopilotUnavailable
	}
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}
	return s.cp.TaskLogs(ctx, sessionID)
}
