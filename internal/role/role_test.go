package role

import "testing"

func TestRoleIsEnvironment(t *testing.T) {
	r := &Role{Name: "exec", Type: Environment, Path: "/tmp/exec"}
	if !r.IsEnvironment() {
		t.Error("expected IsEnvironment() to return true")
	}
	if r.IsWorkspace() {
		t.Error("expected IsWorkspace() to return false")
	}
}

func TestRoleIsWorkspace(t *testing.T) {
	r := &Role{Name: "eng", Type: Workspace, Path: "/tmp/eng"}
	if !r.IsWorkspace() {
		t.Error("expected IsWorkspace() to return true")
	}
	if r.IsEnvironment() {
		t.Error("expected IsEnvironment() to return false")
	}
}

func TestRoleTypeConstants(t *testing.T) {
	if Environment != "environment" {
		t.Errorf("expected Environment = %q, got %q", "environment", Environment)
	}
	if Workspace != "workspace" {
		t.Errorf("expected Workspace = %q, got %q", "workspace", Workspace)
	}
}
