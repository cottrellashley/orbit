package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cottrellashley/orbit/internal/adapter"
	"github.com/cottrellashley/orbit/internal/role"
	"gopkg.in/yaml.v3"
)

// Config is the top-level configuration for orbit.
type Config struct {
	ArchivePath string             `yaml:"archive_path"`
	Adapters    []*adapter.Adapter `yaml:"adapters"`
	Roles       []*role.Role       `yaml:"roles"`
}

// DefaultPath returns the default config file location.
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "orbit", "config.yaml")
}

// Load reads and parses the config file.
func Load(path string) (*Config, error) {
	path = expandHome(path)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read config at %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("cannot parse config at %s: %w", path, err)
	}

	// Expand ~ in paths
	cfg.ArchivePath = expandHome(cfg.ArchivePath)
	for _, r := range cfg.Roles {
		r.Path = expandHome(r.Path)
	}

	return &cfg, nil
}

// Save writes the config to disk.
func Save(path string, cfg *Config) error {
	path = expandHome(path)

	// Create parent directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create config directory %s: %w", dir, err)
	}

	// Make a copy for serialization with unexpanded paths
	out := *cfg
	out.ArchivePath = unexpandHome(cfg.ArchivePath)
	outRoles := make([]*role.Role, len(cfg.Roles))
	for i, r := range cfg.Roles {
		copy := *r
		copy.Path = unexpandHome(r.Path)
		outRoles[i] = &copy
	}
	out.Roles = outRoles

	data, err := yaml.Marshal(&out)
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// FindRole looks up a role by name.
func (c *Config) FindRole(name string) (*role.Role, error) {
	for _, r := range c.Roles {
		if r.Name == name {
			return r, nil
		}
	}
	return nil, fmt.Errorf("role %q not found", name)
}

// AddRole adds a role to the config. Errors if name already exists.
func (c *Config) AddRole(r *role.Role) error {
	if _, err := c.FindRole(r.Name); err == nil {
		return fmt.Errorf("role %q already exists", r.Name)
	}
	c.Roles = append(c.Roles, r)
	return nil
}

// RemoveRole removes a role by name. Returns the removed role.
func (c *Config) RemoveRole(name string) (*role.Role, error) {
	for i, r := range c.Roles {
		if r.Name == name {
			c.Roles = append(c.Roles[:i], c.Roles[i+1:]...)
			return r, nil
		}
	}
	return nil, fmt.Errorf("role %q not found", name)
}

// AdapterRegistry builds an adapter registry from the config.
func (c *Config) AdapterRegistry() *adapter.Registry {
	return adapter.NewRegistry(c.Adapters)
}

// ResolveAdapter finds the adapter for a role (or the default).
func (c *Config) ResolveAdapter(r *role.Role) (*adapter.Adapter, error) {
	reg := c.AdapterRegistry()
	if r.Adapter != "" {
		return reg.Get(r.Adapter)
	}
	return reg.Default()
}

func expandHome(path string) string {
	if len(path) == 0 {
		return path
	}
	if path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}

func unexpandHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if len(path) >= len(home) && path[:len(home)] == home {
		return "~" + path[len(home):]
	}
	return path
}
