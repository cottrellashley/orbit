package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/cottrellashley/orbit/internal/domain"
)

// ---------------------------------------------------------------------------
// stubs for DoctorService
// ---------------------------------------------------------------------------

type stubDoctorProvider struct {
	installed  bool
	version    string
	versionErr error
}

func (p *stubDoctorProvider) DiscoverServers(_ context.Context) ([]domain.Server, error) {
	return nil, nil
}
func (p *stubDoctorProvider) ListSessions(_ context.Context, _ domain.Server) ([]domain.Session, error) {
	return nil, nil
}
func (p *stubDoctorProvider) GetSession(_ context.Context, _ domain.Server, _ string) (*domain.Session, error) {
	return nil, errors.New("not implemented")
}
func (p *stubDoctorProvider) CreateSession(_ context.Context, _ domain.Server, _ string) (*domain.Session, error) {
	return nil, errors.New("not implemented")
}
func (p *stubDoctorProvider) AbortSession(_ context.Context, _ domain.Server, _ string) error {
	return nil
}
func (p *stubDoctorProvider) DeleteSession(_ context.Context, _ domain.Server, _ string) error {
	return nil
}
func (p *stubDoctorProvider) IsInstalled() bool { return p.installed }
func (p *stubDoctorProvider) Version(_ context.Context) (string, error) {
	return p.version, p.versionErr
}

// ---------------------------------------------------------------------------
// DoctorService.checkConfigDir (via Run)
// ---------------------------------------------------------------------------

func TestDoctor_ConfigDirPass(t *testing.T) {
	dir := t.TempDir()
	svc := NewDoctorService(dir, nil, &stubEnvRepo{}, &stubDoctorProvider{installed: true, version: "1.0"})
	svc.SetTmuxLookup(func(s string) (string, error) { return "/usr/bin/tmux", nil })

	report := svc.Run(context.Background())

	var configResult *domain.CheckResult
	for i := range report.Results {
		if report.Results[i].Name == "config-dir" {
			configResult = &report.Results[i]
			break
		}
	}
	if configResult == nil {
		t.Fatal("config-dir check not found in report")
	}
	if configResult.Status != domain.CheckPass {
		t.Errorf("config-dir status = %v, want pass; message: %s", configResult.Status, configResult.Message)
	}
}

func TestDoctor_ConfigDirFail(t *testing.T) {
	nonexistent := filepath.Join(t.TempDir(), "does_not_exist")
	svc := NewDoctorService(nonexistent, nil, &stubEnvRepo{}, &stubDoctorProvider{installed: true, version: "1.0"})
	svc.SetTmuxLookup(func(s string) (string, error) { return "/usr/bin/tmux", nil })

	report := svc.Run(context.Background())

	var configResult *domain.CheckResult
	for i := range report.Results {
		if report.Results[i].Name == "config-dir" {
			configResult = &report.Results[i]
			break
		}
	}
	if configResult == nil {
		t.Fatal("config-dir check not found in report")
	}
	if configResult.Status != domain.CheckFail {
		t.Errorf("config-dir status = %v, want fail", configResult.Status)
	}
}

// ---------------------------------------------------------------------------
// DoctorService.checkTmux (via SetTmuxLookup)
// ---------------------------------------------------------------------------

func TestDoctor_TmuxPass(t *testing.T) {
	dir := t.TempDir()
	svc := NewDoctorService(dir, nil, &stubEnvRepo{}, &stubDoctorProvider{installed: true, version: "1.0"})
	svc.SetTmuxLookup(func(_ string) (string, error) { return "/usr/bin/tmux", nil })

	report := svc.Run(context.Background())

	var tmuxResult *domain.CheckResult
	for i := range report.Results {
		if report.Results[i].Name == "tmux" {
			tmuxResult = &report.Results[i]
			break
		}
	}
	if tmuxResult == nil {
		t.Fatal("tmux check not found in report")
	}
	if tmuxResult.Status != domain.CheckPass {
		t.Errorf("tmux status = %v, want pass", tmuxResult.Status)
	}
}

func TestDoctor_TmuxFail(t *testing.T) {
	dir := t.TempDir()
	svc := NewDoctorService(dir, nil, &stubEnvRepo{}, &stubDoctorProvider{installed: true, version: "1.0"})
	svc.SetTmuxLookup(func(_ string) (string, error) { return "", errors.New("not found") })

	report := svc.Run(context.Background())

	var tmuxResult *domain.CheckResult
	for i := range report.Results {
		if report.Results[i].Name == "tmux" {
			tmuxResult = &report.Results[i]
			break
		}
	}
	if tmuxResult == nil {
		t.Fatal("tmux check not found in report")
	}
	if tmuxResult.Status != domain.CheckFail {
		t.Errorf("tmux status = %v, want fail", tmuxResult.Status)
	}
}

func TestDoctor_TmuxNoLookup(t *testing.T) {
	dir := t.TempDir()
	svc := NewDoctorService(dir, nil, &stubEnvRepo{}, &stubDoctorProvider{installed: true, version: "1.0"})
	// No SetTmuxLookup — should produce a warning.

	report := svc.Run(context.Background())

	var tmuxResult *domain.CheckResult
	for i := range report.Results {
		if report.Results[i].Name == "tmux" {
			tmuxResult = &report.Results[i]
			break
		}
	}
	if tmuxResult == nil {
		t.Fatal("tmux check not found in report")
	}
	if tmuxResult.Status != domain.CheckWarn {
		t.Errorf("tmux status = %v, want warn", tmuxResult.Status)
	}
}

// ---------------------------------------------------------------------------
// DoctorService.checkProvider
// ---------------------------------------------------------------------------

func TestDoctor_ProviderNotInstalled(t *testing.T) {
	dir := t.TempDir()
	svc := NewDoctorService(dir, nil, &stubEnvRepo{}, &stubDoctorProvider{installed: false})
	svc.SetTmuxLookup(func(s string) (string, error) { return "/usr/bin/tmux", nil })

	report := svc.Run(context.Background())

	var providerResult *domain.CheckResult
	for i := range report.Results {
		if report.Results[i].Name == "coding-agent" {
			providerResult = &report.Results[i]
			break
		}
	}
	if providerResult == nil {
		t.Fatal("coding-agent check not found in report")
	}
	if providerResult.Status != domain.CheckFail {
		t.Errorf("coding-agent status = %v, want fail", providerResult.Status)
	}
}

func TestDoctor_ProviderVersionError(t *testing.T) {
	dir := t.TempDir()
	p := &stubDoctorProvider{installed: true, versionErr: errors.New("exec error")}
	svc := NewDoctorService(dir, nil, &stubEnvRepo{}, p)
	svc.SetTmuxLookup(func(s string) (string, error) { return "/usr/bin/tmux", nil })

	report := svc.Run(context.Background())

	var providerResult *domain.CheckResult
	for i := range report.Results {
		if report.Results[i].Name == "coding-agent" {
			providerResult = &report.Results[i]
			break
		}
	}
	if providerResult == nil {
		t.Fatal("coding-agent check not found in report")
	}
	if providerResult.Status != domain.CheckWarn {
		t.Errorf("coding-agent status = %v, want warn", providerResult.Status)
	}
}

// ---------------------------------------------------------------------------
// DoctorService.checkEnvironmentPaths
// ---------------------------------------------------------------------------

func TestDoctor_EnvPathPass(t *testing.T) {
	configDir := t.TempDir()
	envDir := t.TempDir()

	envs := []*domain.Environment{{Name: "myenv", Path: envDir}}
	svc := NewDoctorService(configDir, nil, &stubEnvRepo{envs: envs}, &stubDoctorProvider{installed: true, version: "1"})
	svc.SetTmuxLookup(func(s string) (string, error) { return "/usr/bin/tmux", nil })

	report := svc.Run(context.Background())

	var envResult *domain.CheckResult
	for i := range report.Results {
		if report.Results[i].Name == "env/myenv" {
			envResult = &report.Results[i]
			break
		}
	}
	if envResult == nil {
		t.Fatal("env/myenv check not found in report")
	}
	if envResult.Status != domain.CheckPass {
		t.Errorf("env/myenv status = %v, want pass", envResult.Status)
	}
}

func TestDoctor_EnvPathMissing(t *testing.T) {
	configDir := t.TempDir()
	missing := filepath.Join(t.TempDir(), "gone")

	envs := []*domain.Environment{{Name: "gone", Path: missing}}
	svc := NewDoctorService(configDir, nil, &stubEnvRepo{envs: envs}, &stubDoctorProvider{installed: true, version: "1"})
	svc.SetTmuxLookup(func(s string) (string, error) { return "/usr/bin/tmux", nil })

	report := svc.Run(context.Background())

	var envResult *domain.CheckResult
	for i := range report.Results {
		if report.Results[i].Name == "env/gone" {
			envResult = &report.Results[i]
			break
		}
	}
	if envResult == nil {
		t.Fatal("env/gone check not found in report")
	}
	if envResult.Status != domain.CheckWarn {
		t.Errorf("env/gone status = %v, want warn", envResult.Status)
	}
}

func TestDoctor_EnvPathIsFile(t *testing.T) {
	configDir := t.TempDir()

	// Create a file instead of a directory.
	f, err := os.CreateTemp(t.TempDir(), "notadir")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	envs := []*domain.Environment{{Name: "notadir", Path: f.Name()}}
	svc := NewDoctorService(configDir, nil, &stubEnvRepo{envs: envs}, &stubDoctorProvider{installed: true, version: "1"})
	svc.SetTmuxLookup(func(s string) (string, error) { return "/usr/bin/tmux", nil })

	report := svc.Run(context.Background())

	var envResult *domain.CheckResult
	for i := range report.Results {
		if report.Results[i].Name == "env/notadir" {
			envResult = &report.Results[i]
			break
		}
	}
	if envResult == nil {
		t.Fatal("env/notadir check not found in report")
	}
	if envResult.Status != domain.CheckWarn {
		t.Errorf("env/notadir status = %v, want warn", envResult.Status)
	}
}
