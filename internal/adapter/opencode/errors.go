package opencode

import "errors"

// Sentinel errors for process lifecycle. HTTP/API errors are handled
// by the official opencode-sdk-go (see github.com/sst/opencode-sdk-go.Error).
var (
	ErrNotInstalled = errors.New("opencode binary not found on PATH")
	ErrStartTimeout = errors.New("opencode server did not become healthy in time")
	ErrNotRunning   = errors.New("opencode server is not reachable")
)
