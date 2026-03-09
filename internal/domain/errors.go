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
)
