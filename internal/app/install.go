package app

import (
	"context"
	"fmt"

	"github.com/cottrellashley/orbit/internal/domain"
	"github.com/cottrellashley/orbit/internal/port"
)

// InstallService aggregates all registered ToolInstaller adapters and
// provides a unified interface for checking and installing tools.
type InstallService struct {
	installers []port.ToolInstaller
}

// NewInstallService creates an InstallService from the given installers.
func NewInstallService(installers ...port.ToolInstaller) *InstallService {
	return &InstallService{installers: installers}
}

// ListAll checks every registered tool and returns their current status.
func (s *InstallService) ListAll(ctx context.Context) ([]domain.ToolInfo, error) {
	tools := make([]domain.ToolInfo, 0, len(s.installers))
	for _, inst := range s.installers {
		info, err := inst.Check(ctx)
		if err != nil {
			// On error, return what we know with unknown status.
			tools = append(tools, domain.ToolInfo{
				Name:        inst.Name(),
				Description: inst.Description(),
				Status:      domain.InstallStatusUnknown,
			})
			continue
		}
		tools = append(tools, info)
	}
	return tools, nil
}

// Install finds the installer by name and runs its Install method.
func (s *InstallService) Install(ctx context.Context, name string) (domain.InstallResult, error) {
	for _, inst := range s.installers {
		if inst.Name() == name {
			return inst.Install(ctx)
		}
	}
	return domain.InstallResult{
		Name:  name,
		Error: fmt.Sprintf("unknown tool: %s", name),
	}, fmt.Errorf("tool %q: %w", name, domain.ErrNotFound)
}
