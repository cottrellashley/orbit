package role

// Type distinguishes how a role manages its directory.
type Type string

const (
	// Environment is a single directory that is itself the managed instance.
	Environment Type = "environment"
	// Workspace is a parent directory whose children are each independent instances.
	Workspace Type = "workspace"
)

// Role is the core domain type. A named reference to a directory
// with a type, an adapter, and optional tags.
type Role struct {
	Name    string   `yaml:"name"`
	Type    Type     `yaml:"type"`
	Path    string   `yaml:"path"`
	Adapter string   `yaml:"adapter,omitempty"`
	Tags    []string `yaml:"tags,omitempty"`
}

// IsEnvironment returns true if the role is a single environment.
func (r *Role) IsEnvironment() bool {
	return r.Type == Environment
}

// IsWorkspace returns true if the role is a workspace with child environments.
func (r *Role) IsWorkspace() bool {
	return r.Type == Workspace
}
