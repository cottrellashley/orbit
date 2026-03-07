package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cottrellashley/orbit/internal/adapter"
	"github.com/cottrellashley/orbit/internal/config"
	"github.com/cottrellashley/orbit/internal/role"
)

// setupTestConfig creates a temporary config file with sensible defaults
// and returns the config path and a cleanup function.
func setupTestConfig(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	archivePath := filepath.Join(dir, "archive")

	cfg := &config.Config{
		ArchivePath: archivePath,
		Adapters: []*adapter.Adapter{
			{Name: "echo", Command: "echo", Default: true},
		},
		Roles: []*role.Role{},
	}

	if err := config.Save(cfgPath, cfg); err != nil {
		t.Fatalf("cannot create test config: %v", err)
	}

	return cfgPath, dir
}

func TestInitCmdEnvironment(t *testing.T) {
	cfgPath, dir := setupTestConfig(t)
	envPath := filepath.Join(dir, "myenv")

	cmd := InitCmd()
	cmd.Root().PersistentFlags().String("config", cfgPath, "")
	cmd.SetArgs([]string{"testenv", "--type", "environment", "--path", envPath, "--tag", "test"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	// Verify directory was scaffolded
	if _, err := os.Stat(filepath.Join(envPath, "opencode.json")); os.IsNotExist(err) {
		t.Error("expected opencode.json to be scaffolded")
	}
	if _, err := os.Stat(filepath.Join(envPath, "AGENTS.md")); os.IsNotExist(err) {
		t.Error("expected AGENTS.md to be scaffolded")
	}
	if _, err := os.Stat(filepath.Join(envPath, ".opencode", "commands")); os.IsNotExist(err) {
		t.Error("expected .opencode/commands to be scaffolded")
	}

	// Verify role was saved to config
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("cannot load config: %v", err)
	}

	r, err := cfg.FindRole("testenv")
	if err != nil {
		t.Fatalf("role not found in config: %v", err)
	}
	if r.Type != role.Environment {
		t.Errorf("expected type %q, got %q", role.Environment, r.Type)
	}
}

func TestInitCmdWorkspace(t *testing.T) {
	cfgPath, dir := setupTestConfig(t)
	wsPath := filepath.Join(dir, "myworkspace")

	cmd := InitCmd()
	cmd.Root().PersistentFlags().String("config", cfgPath, "")
	cmd.SetArgs([]string{"testws", "--type", "workspace", "--path", wsPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	// Workspace should exist but NOT have opencode.json
	if _, err := os.Stat(wsPath); os.IsNotExist(err) {
		t.Error("expected workspace directory to exist")
	}
	if _, err := os.Stat(filepath.Join(wsPath, "opencode.json")); !os.IsNotExist(err) {
		t.Error("workspace should NOT have opencode.json scaffolded")
	}
}

func TestInitCmdDuplicateName(t *testing.T) {
	cfgPath, dir := setupTestConfig(t)

	cmd := InitCmd()
	cmd.Root().PersistentFlags().String("config", cfgPath, "")
	cmd.SetArgs([]string{"dup", "--type", "environment", "--path", filepath.Join(dir, "first")})
	cmd.Execute()

	// Second init with same name should fail
	cmd2 := InitCmd()
	cmd2.Root().PersistentFlags().String("config", cfgPath, "")
	cmd2.SetArgs([]string{"dup", "--type", "environment", "--path", filepath.Join(dir, "second")})

	if err := cmd2.Execute(); err == nil {
		t.Error("expected error when adding duplicate role name")
	}
}

func TestInitCmdInvalidType(t *testing.T) {
	cfgPath, dir := setupTestConfig(t)

	cmd := InitCmd()
	cmd.Root().PersistentFlags().String("config", cfgPath, "")
	cmd.SetArgs([]string{"bad", "--type", "invalid", "--path", filepath.Join(dir, "bad")})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error for invalid role type")
	}
}

func TestNewCmdCreatesProject(t *testing.T) {
	cfgPath, dir := setupTestConfig(t)
	wsPath := filepath.Join(dir, "workspace")

	// First init a workspace
	initCmd := InitCmd()
	initCmd.Root().PersistentFlags().String("config", cfgPath, "")
	initCmd.SetArgs([]string{"ws", "--type", "workspace", "--path", wsPath})
	initCmd.Execute()

	// Create a new project in the workspace
	cmd := NewCmd()
	cmd.Root().PersistentFlags().String("config", cfgPath, "")
	cmd.SetArgs([]string{"ws", "myproject"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("new command failed: %v", err)
	}

	projectPath := filepath.Join(wsPath, "myproject")
	if _, err := os.Stat(filepath.Join(projectPath, "opencode.json")); os.IsNotExist(err) {
		t.Error("expected opencode.json in new project")
	}
	if _, err := os.Stat(filepath.Join(projectPath, "AGENTS.md")); os.IsNotExist(err) {
		t.Error("expected AGENTS.md in new project")
	}
}

func TestNewCmdRejectsEnvironment(t *testing.T) {
	cfgPath, dir := setupTestConfig(t)

	initCmd := InitCmd()
	initCmd.Root().PersistentFlags().String("config", cfgPath, "")
	initCmd.SetArgs([]string{"env", "--type", "environment", "--path", filepath.Join(dir, "env")})
	initCmd.Execute()

	cmd := NewCmd()
	cmd.Root().PersistentFlags().String("config", cfgPath, "")
	cmd.SetArgs([]string{"env", "subproject"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error creating project in environment role")
	}
}

func TestListCmd(t *testing.T) {
	cfgPath, dir := setupTestConfig(t)

	// Create a couple of roles
	for _, name := range []string{"a", "b"} {
		cmd := InitCmd()
		cmd.Root().PersistentFlags().String("config", cfgPath, "")
		cmd.SetArgs([]string{name, "--type", "environment", "--path", filepath.Join(dir, name)})
		cmd.Execute()
	}

	cmd := ListCmd()
	cmd.Root().PersistentFlags().String("config", cfgPath, "")
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("list command failed: %v", err)
	}
}

func TestListCmdWithTagFilter(t *testing.T) {
	cfgPath, dir := setupTestConfig(t)

	// Create roles with different tags
	cmd1 := InitCmd()
	cmd1.Root().PersistentFlags().String("config", cfgPath, "")
	cmd1.SetArgs([]string{"tagged", "--type", "environment", "--path", filepath.Join(dir, "tagged"), "--tag", "special"})
	cmd1.Execute()

	cmd2 := InitCmd()
	cmd2.Root().PersistentFlags().String("config", cfgPath, "")
	cmd2.SetArgs([]string{"untagged", "--type", "environment", "--path", filepath.Join(dir, "untagged")})
	cmd2.Execute()

	cmd := ListCmd()
	cmd.Root().PersistentFlags().String("config", cfgPath, "")
	cmd.SetArgs([]string{"--tag", "special"})

	// Should not error
	if err := cmd.Execute(); err != nil {
		t.Fatalf("list with tag filter failed: %v", err)
	}
}

func TestStatusCmd(t *testing.T) {
	cfgPath, dir := setupTestConfig(t)

	initCmd := InitCmd()
	initCmd.Root().PersistentFlags().String("config", cfgPath, "")
	initCmd.SetArgs([]string{"myenv", "--type", "environment", "--path", filepath.Join(dir, "myenv")})
	initCmd.Execute()

	// Overall status
	cmd := StatusCmd()
	cmd.Root().PersistentFlags().String("config", cfgPath, "")
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("status command failed: %v", err)
	}

	// Specific role status
	cmd2 := StatusCmd()
	cmd2.Root().PersistentFlags().String("config", cfgPath, "")
	cmd2.SetArgs([]string{"myenv"})

	if err := cmd2.Execute(); err != nil {
		t.Fatalf("status for specific role failed: %v", err)
	}
}

func TestArchiveCmdEnvironment(t *testing.T) {
	cfgPath, dir := setupTestConfig(t)
	envPath := filepath.Join(dir, "todelete")

	initCmd := InitCmd()
	initCmd.Root().PersistentFlags().String("config", cfgPath, "")
	initCmd.SetArgs([]string{"todelete", "--type", "environment", "--path", envPath})
	initCmd.Execute()

	cmd := ArchiveCmd()
	cmd.Root().PersistentFlags().String("config", cfgPath, "")
	cmd.SetArgs([]string{"todelete"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("archive command failed: %v", err)
	}

	// Original directory should be gone
	if _, err := os.Stat(envPath); !os.IsNotExist(err) {
		t.Error("expected environment directory to be removed after archive")
	}

	// Role should be removed from config
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("cannot load config: %v", err)
	}
	if _, err := cfg.FindRole("todelete"); err == nil {
		t.Error("expected role to be removed from config after archive")
	}
}

func TestArchiveCmdWorkspaceProject(t *testing.T) {
	cfgPath, dir := setupTestConfig(t)
	wsPath := filepath.Join(dir, "workspace")

	// Init workspace
	initCmd := InitCmd()
	initCmd.Root().PersistentFlags().String("config", cfgPath, "")
	initCmd.SetArgs([]string{"ws", "--type", "workspace", "--path", wsPath})
	initCmd.Execute()

	// Create a project
	newCmd := NewCmd()
	newCmd.Root().PersistentFlags().String("config", cfgPath, "")
	newCmd.SetArgs([]string{"ws", "proj"})
	newCmd.Execute()

	// Archive just the project
	cmd := ArchiveCmd()
	cmd.Root().PersistentFlags().String("config", cfgPath, "")
	cmd.SetArgs([]string{"ws/proj"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("archive workspace project failed: %v", err)
	}

	// Project should be gone
	projPath := filepath.Join(wsPath, "proj")
	if _, err := os.Stat(projPath); !os.IsNotExist(err) {
		t.Error("expected project directory to be removed")
	}

	// Workspace role should still exist in config
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("cannot load config: %v", err)
	}
	if _, err := cfg.FindRole("ws"); err != nil {
		t.Error("workspace role should still exist after archiving a project")
	}
}
