package workspace_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cottrellashley/orbit/internal/adapter/workspace"
	"github.com/cottrellashley/orbit/internal/port"
)

// Verify Layout satisfies the port.ConfigWorkspace interface at compile time.
var _ port.ConfigWorkspace = (*workspace.Layout)(nil)

func TestNew(t *testing.T) {
	root := "/tmp/test-orbit"
	l := workspace.New(root)
	if l.Root() != root {
		t.Errorf("Root() = %q, want %q", l.Root(), root)
	}
}

func TestSubdirectoryPaths(t *testing.T) {
	root := "/home/user/.config/orbit"
	l := workspace.New(root)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ProfilesDir", l.ProfilesDir(), filepath.Join(root, "profiles")},
		{"SkillsDir", l.SkillsDir(), filepath.Join(root, "skills")},
		{"AgentsDir", l.AgentsDir(), filepath.Join(root, "agents")},
		{"PlansDir", l.PlansDir(), filepath.Join(root, "plans")},
		{"MCPDir", l.MCPDir(), filepath.Join(root, "mcp")},
		{"CommandsDir", l.CommandsDir(), filepath.Join(root, "commands")},
		{"StateDir", l.StateDir(), filepath.Join(root, "state")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestEnsureDirs(t *testing.T) {
	root := t.TempDir()
	l := workspace.New(root)

	if err := l.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs() error: %v", err)
	}

	// Verify all expected subdirectories were created.
	expected := []string{
		"profiles", "skills", "agents", "plans", "mcp", "commands", "state",
	}
	for _, sub := range expected {
		dir := filepath.Join(root, sub)
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("expected directory %q to exist: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected %q to be a directory", dir)
		}
	}
}

func TestEnsureDirsIdempotent(t *testing.T) {
	root := t.TempDir()
	l := workspace.New(root)

	// Call twice — should not error the second time.
	if err := l.EnsureDirs(); err != nil {
		t.Fatalf("first EnsureDirs() error: %v", err)
	}
	if err := l.EnsureDirs(); err != nil {
		t.Fatalf("second EnsureDirs() error: %v", err)
	}
}

func TestDefaultRoot(t *testing.T) {
	// When XDG_CONFIG_HOME is set, DefaultRoot should use it.
	original := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", original)

	os.Setenv("XDG_CONFIG_HOME", "/custom/config")
	got := workspace.DefaultRoot()
	want := filepath.Join("/custom/config", "orbit")
	if got != want {
		t.Errorf("DefaultRoot() with XDG = %q, want %q", got, want)
	}

	// When XDG_CONFIG_HOME is unset, falls back to ~/.config/orbit.
	os.Unsetenv("XDG_CONFIG_HOME")
	got = workspace.DefaultRoot()
	home, _ := os.UserHomeDir()
	want = filepath.Join(home, ".config", "orbit")
	if got != want {
		t.Errorf("DefaultRoot() without XDG = %q, want %q", got, want)
	}
}
