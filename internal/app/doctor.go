package app

import (
	"context"
	"fmt"
	"os"

	"github.com/cottrellashley/orbit/internal/domain"
	"github.com/cottrellashley/orbit/internal/port"
)

// DoctorService runs diagnostic checks to verify the orbit environment
// is correctly configured and dependencies are available.
type DoctorService struct {
	configDir string
	profiles  port.ProfileRepository
	envs      port.EnvironmentRepository
	provider  port.SessionProvider

	// tmuxLookup allows test injection.
	tmuxLookup func(string) (string, error)
}

// NewDoctorService creates a DoctorService.
func NewDoctorService(configDir string, profiles port.ProfileRepository, envs port.EnvironmentRepository, provider port.SessionProvider) *DoctorService {
	return &DoctorService{
		configDir: configDir,
		profiles:  profiles,
		envs:      envs,
		provider:  provider,
	}
}

// SetTmuxLookup overrides the function used to find the tmux binary.
// Useful for testing.
func (d *DoctorService) SetTmuxLookup(fn func(string) (string, error)) {
	d.tmuxLookup = fn
}

// Run executes all diagnostic checks and returns a report.
func (d *DoctorService) Run(ctx context.Context) *domain.Report {
	r := &domain.Report{}
	r.Results = append(r.Results, d.checkConfigDir())
	r.Results = append(r.Results, d.checkProfilesDir())
	r.Results = append(r.Results, d.checkProvider(ctx))
	r.Results = append(r.Results, d.checkTmux())
	r.Results = append(r.Results, d.checkEnvironmentPaths()...)
	return r
}

func (d *DoctorService) checkConfigDir() domain.CheckResult {
	info, err := os.Stat(d.configDir)
	if err != nil {
		return domain.CheckResult{
			Name:    "config-dir",
			Status:  domain.CheckFail,
			Message: fmt.Sprintf("config directory not found: %s", d.configDir),
			Fix:     fmt.Sprintf("run: mkdir -p %s", d.configDir),
		}
	}
	if !info.IsDir() {
		return domain.CheckResult{
			Name:    "config-dir",
			Status:  domain.CheckFail,
			Message: fmt.Sprintf("%s exists but is not a directory", d.configDir),
			Fix:     fmt.Sprintf("remove the file and run: mkdir -p %s", d.configDir),
		}
	}
	return domain.CheckResult{
		Name:    "config-dir",
		Status:  domain.CheckPass,
		Message: fmt.Sprintf("config directory exists: %s", d.configDir),
	}
}

func (d *DoctorService) checkProfilesDir() domain.CheckResult {
	if d.profiles == nil {
		return domain.CheckResult{
			Name:    "profiles-dir",
			Status:  domain.CheckWarn,
			Message: "profile repository not configured",
			Fix:     "profiles support is not yet implemented",
		}
	}
	dir := d.profiles.Dir()
	info, err := os.Stat(dir)
	if err != nil {
		return domain.CheckResult{
			Name:    "profiles-dir",
			Status:  domain.CheckWarn,
			Message: fmt.Sprintf("profiles directory not found: %s", dir),
			Fix:     fmt.Sprintf("run: mkdir -p %s", dir),
		}
	}
	if !info.IsDir() {
		return domain.CheckResult{
			Name:    "profiles-dir",
			Status:  domain.CheckFail,
			Message: fmt.Sprintf("%s exists but is not a directory", dir),
			Fix:     fmt.Sprintf("remove the file and run: mkdir -p %s", dir),
		}
	}
	return domain.CheckResult{
		Name:    "profiles-dir",
		Status:  domain.CheckPass,
		Message: fmt.Sprintf("profiles directory exists: %s", dir),
	}
}

func (d *DoctorService) checkProvider(ctx context.Context) domain.CheckResult {
	if !d.provider.IsInstalled() {
		return domain.CheckResult{
			Name:    "coding-agent",
			Status:  domain.CheckFail,
			Message: "coding agent binary not found on PATH",
			Fix:     "install the coding agent (e.g. opencode: https://opencode.ai)",
		}
	}

	ver, err := d.provider.Version(ctx)
	if err != nil {
		return domain.CheckResult{
			Name:    "coding-agent",
			Status:  domain.CheckWarn,
			Message: "coding agent found but could not determine version",
			Fix:     "check that the coding agent runs correctly",
		}
	}

	return domain.CheckResult{
		Name:    "coding-agent",
		Status:  domain.CheckPass,
		Message: fmt.Sprintf("coding agent installed: %s", ver),
	}
}

func (d *DoctorService) checkTmux() domain.CheckResult {
	lookup := d.tmuxLookup
	if lookup == nil {
		return domain.CheckResult{
			Name:    "tmux",
			Status:  domain.CheckWarn,
			Message: "tmux check not configured",
		}
	}
	path, err := lookup("tmux")
	if err != nil {
		return domain.CheckResult{
			Name:    "tmux",
			Status:  domain.CheckFail,
			Message: "tmux not found on PATH",
			Fix:     "install tmux: brew install tmux (macOS) or apt install tmux (Linux)",
		}
	}
	return domain.CheckResult{
		Name:    "tmux",
		Status:  domain.CheckPass,
		Message: fmt.Sprintf("tmux installed: %s", path),
	}
}

func (d *DoctorService) checkEnvironmentPaths() []domain.CheckResult {
	envs, err := d.envs.List()
	if err != nil {
		return []domain.CheckResult{{
			Name:    "environments",
			Status:  domain.CheckWarn,
			Message: fmt.Sprintf("could not load environments: %v", err),
		}}
	}

	if len(envs) == 0 {
		return []domain.CheckResult{{
			Name:    "environments",
			Status:  domain.CheckPass,
			Message: "no environments registered",
		}}
	}

	var results []domain.CheckResult
	for _, env := range envs {
		info, err := os.Stat(env.Path)
		if err != nil {
			results = append(results, domain.CheckResult{
				Name:    fmt.Sprintf("env/%s", env.Name),
				Status:  domain.CheckWarn,
				Message: fmt.Sprintf("path does not exist: %s", env.Path),
				Fix:     fmt.Sprintf("create the directory or remove the registration"),
			})
			continue
		}
		if !info.IsDir() {
			results = append(results, domain.CheckResult{
				Name:    fmt.Sprintf("env/%s", env.Name),
				Status:  domain.CheckWarn,
				Message: fmt.Sprintf("path is not a directory: %s", env.Path),
			})
			continue
		}
		results = append(results, domain.CheckResult{
			Name:    fmt.Sprintf("env/%s", env.Name),
			Status:  domain.CheckPass,
			Message: fmt.Sprintf("path exists: %s", env.Path),
		})
	}
	return results
}
