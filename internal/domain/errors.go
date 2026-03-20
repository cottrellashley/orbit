package domain

import "errors"

// Sentinel errors used across the domain.
var (
	// ErrNotFound indicates a requested resource does not exist.
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists indicates a resource with the same identifier already exists.
	ErrAlreadyExists = errors.New("already exists")

	// ErrInvalidPath indicates a filesystem path is invalid or does not exist.
	ErrInvalidPath = errors.New("invalid path")

	// ErrNotAuthenticated indicates a required authentication credential
	// is missing or invalid (e.g. no GitHub token available).
	ErrNotAuthenticated = errors.New("not authenticated")

	// ErrRateLimited indicates the remote API rate limit has been exceeded.
	ErrRateLimited = errors.New("rate limited")

	// ErrRendererUnavailable indicates no full-featured markdown renderer
	// is available and the fallback was used instead.
	ErrRendererUnavailable = errors.New("renderer unavailable")

	// ErrNodeNotFound indicates a requested node does not exist in the registry.
	ErrNodeNotFound = errors.New("node not found")

	// ErrNodeUnhealthy indicates a node exists but is not responding to health checks.
	ErrNodeUnhealthy = errors.New("node unhealthy")

	// ErrCopilotUnavailable indicates the Copilot agent-task tooling is not
	// available (e.g. gh CLI too old or not installed).
	ErrCopilotUnavailable = errors.New("copilot agent-task not available")
)
