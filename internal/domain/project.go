package domain

import "time"

// ProjectTopology describes how a project relates to git repositories.
type ProjectTopology int

const (
	// TopologyUnknown means the topology has not been detected yet.
	TopologyUnknown ProjectTopology = iota
	// TopologySingleRepo means the project root is itself a git repository.
	TopologySingleRepo
	// TopologyMultiRepo means the project root contains multiple git repositories.
	TopologyMultiRepo
)

// String returns a human-readable label.
func (t ProjectTopology) String() string {
	switch t {
	case TopologySingleRepo:
		return "single-repo"
	case TopologyMultiRepo:
		return "multi-repo"
	default:
		return "unknown"
	}
}

// IntegrationTag identifies a detected tool or platform integration within
// a project. Tags are auto-detected from project directory contents and
// used by the UI to display visual indicators.
type IntegrationTag string

const (
	TagGit      IntegrationTag = "git"
	TagPython   IntegrationTag = "python"
	TagUV       IntegrationTag = "uv"
	TagNode     IntegrationTag = "node"
	TagOpenCode IntegrationTag = "opencode"
	TagGitHub   IntegrationTag = "github"
)

// RepoInfo holds metadata about a single git repository discovered
// within a project. For single-repo projects there is one RepoInfo
// at the project root; for multi-repo projects there may be several.
type RepoInfo struct {
	// Path is the absolute filesystem path to the repository root.
	Path string
	// RemoteURL is the primary remote URL (empty if none configured).
	RemoteURL string
	// CurrentBranch is the checked-out branch (empty if detached HEAD).
	CurrentBranch string
}

// Project is the successor concept to Environment. It represents a
// registered workspace — a filesystem path with richer metadata about
// repository topology, detected integrations, and contained repos.
//
// During the migration period both Environment and Project may coexist.
// See [ProjectFromEnvironment] for the compatibility bridge.
type Project struct {
	// Name is the unique identifier for this project.
	Name string
	// Path is the absolute filesystem path to the project root.
	Path string
	// Description is a human-readable summary.
	Description string
	// ProfileName is the profile used at creation (may be empty).
	ProfileName string
	// Topology describes the project's relationship to git repositories.
	Topology ProjectTopology
	// Integrations lists the detected tool/platform tags.
	Integrations []IntegrationTag
	// Repos holds metadata for git repositories within the project.
	// May be empty if detection has not run yet.
	Repos []RepoInfo
	// CreatedAt is when the project was registered.
	CreatedAt time.Time
	// UpdatedAt is when the project metadata was last refreshed.
	UpdatedAt time.Time
}

// ProjectFromEnvironment converts a legacy Environment into a Project.
// The resulting Project has TopologyUnknown and no integrations — those
// must be populated by a detection pass. This is the forward-migration
// bridge used during the Environment → Project transition.
func ProjectFromEnvironment(env *Environment) *Project {
	if env == nil {
		return nil
	}
	return &Project{
		Name:        env.Name,
		Path:        env.Path,
		Description: env.Description,
		ProfileName: env.ProfileName,
		Topology:    TopologyUnknown,
		CreatedAt:   env.CreatedAt,
		UpdatedAt:   env.UpdatedAt,
	}
}

// EnvironmentFromProject converts a Project back to a legacy Environment.
// This is the backward-compatibility bridge so existing code that operates
// on Environment can continue to work during the migration period.
func EnvironmentFromProject(p *Project) *Environment {
	if p == nil {
		return nil
	}
	return &Environment{
		Name:        p.Name,
		Path:        p.Path,
		Description: p.Description,
		ProfileName: p.ProfileName,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
