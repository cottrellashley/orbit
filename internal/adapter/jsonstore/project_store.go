// project_store.go implements port.ProjectRepository using atomic JSON
// file persistence. The data file lives at <dir>/projects.json and uses
// a versioned envelope:
//
//	{"version": 1, "projects": [...]}
//
// On first load, if projects.json does not exist but environments.json
// does, the store auto-migrates by reading the legacy environment data
// and converting it to projects via domain.ProjectFromEnvironment. The
// original environments.json is never modified or removed.
package jsonstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
)

const projectCollection = "projects"

// currentProjectVersion is the storage format version. Bump this when
// the DTO shape changes in a backward-incompatible way.
const currentProjectVersion = 1

// ---------------------------------------------------------------------------
// JSON DTOs — keeps serialization details out of the domain
// ---------------------------------------------------------------------------

type repoInfoDTO struct {
	Path          string `json:"path"`
	RemoteURL     string `json:"remote_url,omitempty"`
	CurrentBranch string `json:"current_branch,omitempty"`
}

type projectDTO struct {
	Name         string        `json:"name"`
	Path         string        `json:"path"`
	Description  string        `json:"description"`
	ProfileName  string        `json:"profile_name,omitempty"`
	Topology     string        `json:"topology"`
	Integrations []string      `json:"integrations,omitempty"`
	Repos        []repoInfoDTO `json:"repos,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// projectFileV1 is the versioned on-disk envelope for project storage.
type projectFileV1 struct {
	Version  int          `json:"version"`
	Projects []projectDTO `json:"projects"`
}

func toProjectDTO(p *domain.Project) projectDTO {
	dto := projectDTO{
		Name:        p.Name,
		Path:        p.Path,
		Description: p.Description,
		ProfileName: p.ProfileName,
		Topology:    p.Topology.String(),
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
	for _, tag := range p.Integrations {
		dto.Integrations = append(dto.Integrations, string(tag))
	}
	for _, r := range p.Repos {
		dto.Repos = append(dto.Repos, repoInfoDTO{
			Path:          r.Path,
			RemoteURL:     r.RemoteURL,
			CurrentBranch: r.CurrentBranch,
		})
	}
	return dto
}

// parseTopology converts a stored string back to a ProjectTopology.
func parseTopology(s string) domain.ProjectTopology {
	switch s {
	case "single-repo":
		return domain.TopologySingleRepo
	case "multi-repo":
		return domain.TopologyMultiRepo
	default:
		return domain.TopologyUnknown
	}
}

func fromProjectDTO(d projectDTO) *domain.Project {
	p := &domain.Project{
		Name:        d.Name,
		Path:        d.Path,
		Description: d.Description,
		ProfileName: d.ProfileName,
		Topology:    parseTopology(d.Topology),
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
	for _, tag := range d.Integrations {
		p.Integrations = append(p.Integrations, domain.IntegrationTag(tag))
	}
	for _, r := range d.Repos {
		p.Repos = append(p.Repos, domain.RepoInfo{
			Path:          r.Path,
			RemoteURL:     r.RemoteURL,
			CurrentBranch: r.CurrentBranch,
		})
	}
	return p
}

// ---------------------------------------------------------------------------
// ProjectStore — port.ProjectRepository implementation
// ---------------------------------------------------------------------------

// ProjectStore persists project registry entries as a versioned JSON
// file. It can auto-migrate from a legacy environments.json file.
type ProjectStore struct {
	dir string
}

// NewProjectStore creates a ProjectStore rooted at dir.
// The directory is created on the first call to Save.
func NewProjectStore(dir string) *ProjectStore {
	return &ProjectStore{dir: dir}
}

// Dir returns the store's base directory.
func (s *ProjectStore) Dir() string {
	return s.dir
}

// List returns all registered projects.
// Returns an empty slice (not nil) if none exist.
func (s *ProjectStore) List() ([]*domain.Project, error) {
	return s.loadProjects()
}

// Get returns a single project by name.
// Returns domain.ErrNotFound if not found.
func (s *ProjectStore) Get(name string) (*domain.Project, error) {
	projects, err := s.loadProjects()
	if err != nil {
		return nil, err
	}
	for _, p := range projects {
		if p.Name == name {
			return p, nil
		}
	}
	return nil, fmt.Errorf("project %q: %w", name, domain.ErrNotFound)
}

// GetByPath returns the project whose registered path matches.
// Returns domain.ErrNotFound if no match.
func (s *ProjectStore) GetByPath(path string) (*domain.Project, error) {
	projects, err := s.loadProjects()
	if err != nil {
		return nil, err
	}
	for _, p := range projects {
		if p.Path == path {
			return p, nil
		}
	}
	return nil, fmt.Errorf("project at path %q: %w", path, domain.ErrNotFound)
}

// Save persists the full project list (create or update).
func (s *ProjectStore) Save(projects []*domain.Project) error {
	return s.saveProjects(projects)
}

// Delete removes a project by name.
// Returns domain.ErrNotFound if not found.
func (s *ProjectStore) Delete(name string) error {
	projects, err := s.loadProjects()
	if err != nil {
		return err
	}

	idx := -1
	for i, p := range projects {
		if p.Name == name {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("project %q: %w", name, domain.ErrNotFound)
	}

	projects = append(projects[:idx], projects[idx+1:]...)
	return s.saveProjects(projects)
}

// ---------------------------------------------------------------------------
// Internal persistence
// ---------------------------------------------------------------------------

func (s *ProjectStore) projectFilePath() string {
	return filepath.Join(s.dir, projectCollection+".json")
}

func (s *ProjectStore) envFilePath() string {
	return filepath.Join(s.dir, collectionName+".json")
}

// loadProjects reads projects.json. If the file does not exist, it
// attempts to migrate from environments.json. If neither file exists,
// an empty slice is returned.
func (s *ProjectStore) loadProjects() ([]*domain.Project, error) {
	p := s.projectFilePath()

	b, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return s.migrateFromEnvironments()
		}
		return nil, fmt.Errorf("jsonstore: read %s: %w", p, err)
	}

	return s.parseProjectFile(b)
}

// parseProjectFile parses a versioned projects.json payload.
func (s *ProjectStore) parseProjectFile(data []byte) ([]*domain.Project, error) {
	var pf projectFileV1
	if err := json.Unmarshal(data, &pf); err != nil {
		return nil, fmt.Errorf("jsonstore: parse projects.json: %w", err)
	}

	// Version guard — reject files from the future.
	if pf.Version > currentProjectVersion {
		return nil, fmt.Errorf("jsonstore: projects.json version %d is newer than supported (%d); please upgrade orbit",
			pf.Version, currentProjectVersion)
	}

	projects := make([]*domain.Project, len(pf.Projects))
	for i, d := range pf.Projects {
		projects[i] = fromProjectDTO(d)
	}
	return projects, nil
}

// migrateFromEnvironments reads environments.json, converts each entry
// to a Project, and returns the result. It does NOT write projects.json
// automatically — the caller (or a subsequent Save) does that. The
// original environments.json is never modified.
func (s *ProjectStore) migrateFromEnvironments() ([]*domain.Project, error) {
	ep := s.envFilePath()

	b, err := os.ReadFile(ep)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []*domain.Project{}, nil
		}
		return nil, fmt.Errorf("jsonstore: read %s for migration: %w", ep, err)
	}

	var dtos []envDTO
	if err := json.Unmarshal(b, &dtos); err != nil {
		return nil, fmt.Errorf("jsonstore: parse %s for migration: %w", ep, err)
	}

	projects := make([]*domain.Project, len(dtos))
	for i, d := range dtos {
		env := fromDTO(d)
		projects[i] = domain.ProjectFromEnvironment(env)
	}
	return projects, nil
}

// saveProjects writes the project slice atomically with the versioned
// envelope.
func (s *ProjectStore) saveProjects(projects []*domain.Project) error {
	if err := os.MkdirAll(s.dir, 0755); err != nil {
		return fmt.Errorf("jsonstore: create directory %s: %w", s.dir, err)
	}

	dtos := make([]projectDTO, len(projects))
	for i, p := range projects {
		dtos[i] = toProjectDTO(p)
	}

	pf := projectFileV1{
		Version:  currentProjectVersion,
		Projects: dtos,
	}

	b, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		return fmt.Errorf("jsonstore: marshal projects: %w", err)
	}
	b = append(b, '\n')

	// Atomic write: temp file + rename.
	tmp, err := os.CreateTemp(s.dir, projectCollection+"-*.tmp")
	if err != nil {
		return fmt.Errorf("jsonstore: create temp: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("jsonstore: write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("jsonstore: close temp: %w", err)
	}

	dest := s.projectFilePath()
	if err := os.Rename(tmpPath, dest); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("jsonstore: rename %s -> %s: %w", tmpPath, dest, err)
	}

	return nil
}
