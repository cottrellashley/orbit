package port

import (
	"context"

	"github.com/cottrellashley/orbit/internal/domain"
)

// ToolInstaller is implemented by any adapter that can check the
// installation status of a tool and optionally install it.
//
// Each adapter owns one or more tools. For example, the GitHub adapter
// may return separate ToolInstaller values for "gh" and "gh-copilot".
type ToolInstaller interface {
	// Name returns the tool's canonical name (e.g. "opencode", "gh").
	Name() string

	// Description returns a one-line summary of what the tool is.
	Description() string

	// Check returns the current install status and version.
	Check(ctx context.Context) (domain.ToolInfo, error)

	// Install attempts to install the tool and returns the result.
	Install(ctx context.Context) (domain.InstallResult, error)
}
