package adapter

import "testing"

func TestNewRegistry(t *testing.T) {
	adapters := []*Adapter{
		{Name: "opencode", Command: "opencode", Default: true},
		{Name: "cursor", Command: "cursor", Args: []string{"--folder"}},
	}
	reg := NewRegistry(adapters)

	a, err := reg.Get("opencode")
	if err != nil {
		t.Fatalf("expected to find opencode adapter: %v", err)
	}
	if a.Command != "opencode" {
		t.Errorf("expected command %q, got %q", "opencode", a.Command)
	}
}

func TestRegistryGetNotFound(t *testing.T) {
	reg := NewRegistry(nil)
	_, err := reg.Get("missing")
	if err == nil {
		t.Error("expected error for missing adapter")
	}
}

func TestRegistryDefault(t *testing.T) {
	adapters := []*Adapter{
		{Name: "cursor", Command: "cursor"},
		{Name: "opencode", Command: "opencode", Default: true},
	}
	reg := NewRegistry(adapters)

	a, err := reg.Default()
	if err != nil {
		t.Fatalf("expected to find default adapter: %v", err)
	}
	if a.Name != "opencode" {
		t.Errorf("expected default adapter %q, got %q", "opencode", a.Name)
	}
}

func TestRegistryDefaultNone(t *testing.T) {
	adapters := []*Adapter{
		{Name: "cursor", Command: "cursor"},
	}
	reg := NewRegistry(adapters)

	_, err := reg.Default()
	if err == nil {
		t.Error("expected error when no default adapter is set")
	}
}

func TestRegistryAll(t *testing.T) {
	adapters := []*Adapter{
		{Name: "a", Command: "a"},
		{Name: "b", Command: "b"},
	}
	reg := NewRegistry(adapters)

	all := reg.All()
	if len(all) != 2 {
		t.Errorf("expected 2 adapters, got %d", len(all))
	}
}

func TestSplitPath(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"", 0},
		{"/usr/bin:/usr/local/bin", 2},
		{"/single", 1},
		{"/a:/b:/c", 3},
	}

	for _, tt := range tests {
		result := splitPath(tt.input)
		if len(result) != tt.expected {
			t.Errorf("splitPath(%q): expected %d parts, got %d (%v)",
				tt.input, tt.expected, len(result), result)
		}
	}
}
