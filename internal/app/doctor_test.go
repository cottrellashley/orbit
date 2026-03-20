package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cottrellashley/orbit/internal/domain"
)

func TestDoctorService_Run_AllPass(t *testing.T) {
	cfgDir := t.TempDir()
	profilesDir := filepath.Join(cfgDir, "profiles")
	os.MkdirAll(profilesDir, 0755)

	envRepo := newMockEnvRepo()
	profiles := newMockProfileRepo(profilesDir)
	provider := newMockSessionProvider()
	provider.installed = true
	provider.version = "1.0.0"

	svc := NewDoctorService(cfgDir, profiles, envRepo, provider)
	svc.SetToolLookup(func(name string) (string, error) {
		return "/usr/bin/" + name, nil
	})

	report := svc.Run(context.Background())
	for _, r := range report.Results {
		if r.Status == domain.CheckFail {
			t.Errorf("check %q failed: %s", r.Name, r.Message)
		}
	}
}

func TestDoctorService_ToolChecks(t *testing.T) {
	cfgDir := t.TempDir()
	envRepo := newMockEnvRepo()
	provider := newMockSessionProvider()

	svc := NewDoctorService(cfgDir, nil, envRepo, provider)

	// All tools found.
	svc.SetToolLookup(func(name string) (string, error) {
		return "/usr/bin/" + name, nil
	})

	report := svc.Run(context.Background())

	checkNames := make(map[string]domain.CheckStatus)
	for _, r := range report.Results {
		checkNames[r.Name] = r.Status
	}

	for _, name := range []string{"tmux", "uv", "gh", "git"} {
		status, ok := checkNames[name]
		if !ok {
			t.Errorf("expected check %q to be present", name)
			continue
		}
		if status != domain.CheckPass {
			t.Errorf("expected check %q to pass, got %v", name, status)
		}
	}
}

func TestDoctorService_ToolChecks_NotConfigured(t *testing.T) {
	cfgDir := t.TempDir()
	envRepo := newMockEnvRepo()
	provider := newMockSessionProvider()

	svc := NewDoctorService(cfgDir, nil, envRepo, provider)
	// Don't set tool lookup — should get warnings.

	report := svc.Run(context.Background())

	checkNames := make(map[string]domain.CheckStatus)
	for _, r := range report.Results {
		checkNames[r.Name] = r.Status
	}

	for _, name := range []string{"tmux", "uv", "gh", "git"} {
		status, ok := checkNames[name]
		if !ok {
			t.Errorf("expected check %q to be present", name)
			continue
		}
		if status != domain.CheckWarn {
			t.Errorf("expected check %q to warn (no lookup configured), got %v", name, status)
		}
	}
}

func TestDoctorService_ProjectPathChecks(t *testing.T) {
	cfgDir := t.TempDir()
	projDir := t.TempDir()

	// Create a .git directory in projDir to trigger git detection.
	os.MkdirAll(filepath.Join(projDir, ".git"), 0755)

	envRepo := newMockEnvRepo()
	provider := newMockSessionProvider()
	projRepo := newMockProjectRepo(
		&domain.Project{Name: "myproj", Path: projDir},
		&domain.Project{Name: "ghost", Path: "/nonexistent/ghost"},
	)

	svc := NewDoctorService(cfgDir, nil, envRepo, provider)
	svc.SetProjects(projRepo)

	report := svc.Run(context.Background())

	checkMap := make(map[string]domain.CheckResult)
	for _, r := range report.Results {
		checkMap[r.Name] = r
	}

	// myproj should pass and mention git.
	myproj, ok := checkMap["project/myproj"]
	if !ok {
		t.Fatal("expected check 'project/myproj' to be present")
	}
	if myproj.Status != domain.CheckPass {
		t.Fatalf("expected pass for myproj, got %v: %s", myproj.Status, myproj.Message)
	}
	if !strings.Contains(myproj.Message, "git repo detected") {
		t.Fatalf("expected git detection in message, got %q", myproj.Message)
	}

	// ghost should warn.
	ghost, ok := checkMap["project/ghost"]
	if !ok {
		t.Fatal("expected check 'project/ghost' to be present")
	}
	if ghost.Status != domain.CheckWarn {
		t.Fatalf("expected warn for ghost, got %v", ghost.Status)
	}
}

func TestDoctorService_ProjectPathChecks_NoRepo(t *testing.T) {
	cfgDir := t.TempDir()
	envRepo := newMockEnvRepo()
	provider := newMockSessionProvider()

	// No project repo wired — should produce no project checks.
	svc := NewDoctorService(cfgDir, nil, envRepo, provider)

	report := svc.Run(context.Background())
	for _, r := range report.Results {
		if strings.HasPrefix(r.Name, "project/") {
			t.Fatalf("unexpected project check when repo not wired: %s", r.Name)
		}
	}
}

func TestDoctorService_WorkspaceOverridesConfigDir(t *testing.T) {
	origDir := "/nonexistent/config/dir"
	wsDir := t.TempDir()

	envRepo := newMockEnvRepo()
	provider := newMockSessionProvider()

	svc := NewDoctorService(origDir, nil, envRepo, provider)
	svc.SetWorkspace(&mockConfigWorkspace{root: wsDir})

	report := svc.Run(context.Background())

	for _, r := range report.Results {
		if r.Name == "config-dir" {
			if r.Status != domain.CheckPass {
				t.Fatalf("expected pass with workspace override, got %v: %s", r.Status, r.Message)
			}
			if !strings.Contains(r.Message, wsDir) {
				t.Fatalf("expected workspace root in message, got %q", r.Message)
			}
			return
		}
	}
	t.Fatal("config-dir check not found in report")
}

func TestDoctorService_EnvironmentPathChecks(t *testing.T) {
	cfgDir := t.TempDir()
	existingDir := t.TempDir()

	envRepo := newMockEnvRepo(
		&domain.Environment{Name: "valid", Path: existingDir},
		&domain.Environment{Name: "broken", Path: "/nonexistent/broken"},
	)
	provider := newMockSessionProvider()

	svc := NewDoctorService(cfgDir, nil, envRepo, provider)

	report := svc.Run(context.Background())

	checkMap := make(map[string]domain.CheckResult)
	for _, r := range report.Results {
		checkMap[r.Name] = r
	}

	if check, ok := checkMap["env/valid"]; !ok {
		t.Fatal("expected env/valid check")
	} else if check.Status != domain.CheckPass {
		t.Fatalf("expected pass for valid env, got %v", check.Status)
	}

	if check, ok := checkMap["env/broken"]; !ok {
		t.Fatal("expected env/broken check")
	} else if check.Status != domain.CheckWarn {
		t.Fatalf("expected warn for broken env, got %v", check.Status)
	}
}
