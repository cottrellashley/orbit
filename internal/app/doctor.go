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
	projects  port.ProjectRepository // nil until project migration is wired
	provider  port.SessionProvider
	workspace port.ConfigWorkspace // nil if not wired yet

	// toolLookup resolves a binary name to its absolute path. It is
	// injected at construction (defaults to exec.LookPath-compatible
	// function) and may be overridden for tests.
	toolLookup func(string) (string, error)
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

// SetProjects attaches a ProjectRepository for project-level checks.
func (d *DoctorService) SetProjects(repo port.ProjectRepository) {
	d.projects = repo
}

// SetWorkspace attaches a ConfigWorkspace for workspace-level checks.
func (d *DoctorService) SetWorkspace(ws port.ConfigWorkspace) {
	d.workspace = ws
}

// SetTmuxLookup overrides the function used to find the tmux binary.
// Deprecated: use SetToolLookup instead which covers all tool checks.
func (d *DoctorService) SetTmuxLookup(fn func(string) (string, error)) {
	d.toolLookup = fn
}

// SetToolLookup overrides the function used to find binaries by name.
// The function signature matches exec.LookPath. Useful for testing.
func (d *DoctorService) SetToolLookup(fn func(string) (string, error)) {
	d.toolLookup = fn
}

// Run executes all diagnostic checks and returns a report.
func (d *DoctorService) Run(ctx context.Context) *domain.Report {
	r := &domain.Report{}
	r.Results = append(r.Results, d.checkConfigDir())
	r.Results = append(r.Results, d.checkProfilesDir())
	r.Results = append(r.Results, d.checkProvider(ctx))
	r.Results = append(r.Results, d.checkTmux())
	r.Results = append(r.Results, d.checkUV())
	r.Results = append(r.Results, d.checkGH())
	r.Results = append(r.Results, d.checkGit())
	r.Results = append(r.Results, d.checkEnvironmentPaths()...)
	r.Results = append(r.Results, d.checkProjectPaths()...)
	return r
}

func (d *DoctorService) checkConfigDir() domain.CheckResult {
	// Prefer ConfigWorkspace if available.
	dir := d.configDir
	if d.workspace != nil {
		dir = d.workspace.Root()
	}

	info, err := os.Stat(dir)
	if err != nil {
		return domain.CheckResult{
			Name:    "config-dir",
			Status:  domain.CheckFail,
			Message: fmt.Sprintf("config directory not found: %s", dir),
			Fix:     fmt.Sprintf("run: mkdir -p %s", dir),
		}
	}
	if !info.IsDir() {
		return domain.CheckResult{
			Name:    "config-dir",
			Status:  domain.CheckFail,
			Message: fmt.Sprintf("%s exists but is not a directory", dir),
			Fix:     fmt.Sprintf("remove the file and run: mkdir -p %s", dir),
		}
	}
	return domain.CheckResult{
		Name:    "config-dir",
		Status:  domain.CheckPass,
		Message: fmt.Sprintf("config directory exists: %s", dir),
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
	lookup := d.toolLookup
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

// checkUV checks whether the uv package manager is installed.
func (d *DoctorService) checkUV() domain.CheckResult {
	lookup := d.toolLookup
	if lookup == nil {
		return domain.CheckResult{
			Name:    "uv",
			Status:  domain.CheckWarn,
			Message: "tool lookup not configured",
		}
	}
	path, err := lookup("uv")
	if err != nil {
		return domain.CheckResult{
			Name:    "uv",
			Status:  domain.CheckWarn,
			Message: "uv not found on PATH",
			Fix:     "install uv: curl -LsSf https://astral.sh/uv/install.sh | sh",
		}
	}
	return domain.CheckResult{
		Name:    "uv",
		Status:  domain.CheckPass,
		Message: fmt.Sprintf("uv installed: %s", path),
	}
}

// checkGH checks whether the GitHub CLI is installed.
func (d *DoctorService) checkGH() domain.CheckResult {
	lookup := d.toolLookup
	if lookup == nil {
		return domain.CheckResult{
			Name:    "gh",
			Status:  domain.CheckWarn,
			Message: "tool lookup not configured",
		}
	}
	path, err := lookup("gh")
	if err != nil {
		return domain.CheckResult{
			Name:    "gh",
			Status:  domain.CheckWarn,
			Message: "GitHub CLI (gh) not found on PATH",
			Fix:     "install gh: brew install gh (macOS) or see https://cli.github.com",
		}
	}
	return domain.CheckResult{
		Name:    "gh",
		Status:  domain.CheckPass,
		Message: fmt.Sprintf("GitHub CLI installed: %s", path),
	}
}

// checkGit checks whether git is installed.
func (d *DoctorService) checkGit() domain.CheckResult {
	lookup := d.toolLookup
	if lookup == nil {
		return domain.CheckResult{
			Name:    "git",
			Status:  domain.CheckWarn,
			Message: "tool lookup not configured",
		}
	}
	path, err := lookup("git")
	if err != nil {
		return domain.CheckResult{
			Name:    "git",
			Status:  domain.CheckFail,
			Message: "git not found on PATH",
			Fix:     "install git: brew install git (macOS) or apt install git (Linux)",
		}
	}
	return domain.CheckResult{
		Name:    "git",
		Status:  domain.CheckPass,
		Message: fmt.Sprintf("git installed: %s", path),
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
				Fix:     "create the directory or remove the registration",
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

// checkProjectPaths validates registered project paths. Skipped when
// ProjectRepository is not wired.
func (d *DoctorService) checkProjectPaths() []domain.CheckResult {
	if d.projects == nil {
		return nil
	}

	projects, err := d.projects.List()
	if err != nil {
		return []domain.CheckResult{{
			Name:    "projects",
			Status:  domain.CheckWarn,
			Message: fmt.Sprintf("could not load projects: %v", err),
		}}
	}

	if len(projects) == 0 {
		return []domain.CheckResult{{
			Name:    "projects",
			Status:  domain.CheckPass,
			Message: "no projects registered",
		}}
	}

	var results []domain.CheckResult
	for _, proj := range projects {
		info, err := os.Stat(proj.Path)
		if err != nil {
			results = append(results, domain.CheckResult{
				Name:    fmt.Sprintf("project/%s", proj.Name),
				Status:  domain.CheckWarn,
				Message: fmt.Sprintf("path does not exist: %s", proj.Path),
				Fix:     "create the directory or remove the project",
			})
			continue
		}
		if !info.IsDir() {
			results = append(results, domain.CheckResult{
				Name:    fmt.Sprintf("project/%s", proj.Name),
				Status:  domain.CheckWarn,
				Message: fmt.Sprintf("path is not a directory: %s", proj.Path),
			})
			continue
		}

		// Check git repo presence at project root.
		checkName := fmt.Sprintf("project/%s", proj.Name)
		msg := fmt.Sprintf("path exists: %s", proj.Path)

		gitDir := proj.Path + "/.git"
		if _, gitErr := os.Stat(gitDir); gitErr == nil {
			msg += " (git repo detected)"
		}

		results = append(results, domain.CheckResult{
			Name:    checkName,
			Status:  domain.CheckPass,
			Message: msg,
		})
	}
	return results
}
