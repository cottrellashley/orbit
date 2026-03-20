// node_store.go implements port.NodeStore using atomic JSON file
// persistence. The data file lives at <dir>/nodes.json and uses a
// versioned envelope:
//
//	{"version": 1, "nodes": [...]}
//
// Nodes are keyed by their stable UUID. The store is single-user
// (no file locking). Writes use a temp file + rename pattern so a
// crash mid-write cannot corrupt existing data.
package jsonstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
)

const nodeCollection = "nodes"

// currentNodeVersion is the storage format version. Bump this when
// the DTO shape changes in a backward-incompatible way.
const currentNodeVersion = 1

// Compile-time check that NodeStore satisfies port.NodeStore.
// Import cycle prevention: we cannot import the port package here,
// so the wiring check lives in the composition root (cli/root.go).

// ---------------------------------------------------------------------------
// JSON DTOs — keeps serialization details out of the domain
// ---------------------------------------------------------------------------

type nodeDTO struct {
	ID        string    `json:"id"`
	Name      string    `json:"name,omitempty"`
	Provider  string    `json:"provider"`
	Origin    string    `json:"origin"`
	Hostname  string    `json:"hostname"`
	Port      int       `json:"port"`
	Directory string    `json:"directory,omitempty"`
	Version   string    `json:"version,omitempty"`
	PID       int       `json:"pid,omitempty"`
	Healthy   bool      `json:"healthy"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// nodeFileV1 is the versioned on-disk envelope for node storage.
type nodeFileV1 struct {
	Version int       `json:"version"`
	Nodes   []nodeDTO `json:"nodes"`
}

func toNodeDTO(n *domain.Node) nodeDTO {
	return nodeDTO{
		ID:        n.ID,
		Name:      n.Name,
		Provider:  n.Provider.String(),
		Origin:    n.Origin.String(),
		Hostname:  n.Hostname,
		Port:      n.Port,
		Directory: n.Directory,
		Version:   n.Version,
		PID:       n.PID,
		Healthy:   n.Healthy,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
}

func fromNodeDTO(d nodeDTO) *domain.Node {
	return &domain.Node{
		ID:        d.ID,
		Name:      d.Name,
		Provider:  domain.ParseNodeProvider(d.Provider),
		Origin:    domain.ParseNodeOrigin(d.Origin),
		Hostname:  d.Hostname,
		Port:      d.Port,
		Directory: d.Directory,
		Version:   d.Version,
		PID:       d.PID,
		Healthy:   d.Healthy,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

// ---------------------------------------------------------------------------
// NodeStore — port.NodeStore implementation
// ---------------------------------------------------------------------------

// NodeStore persists node registry entries as a versioned JSON file.
type NodeStore struct {
	dir string
}

// NewNodeStore creates a NodeStore rooted at dir.
// The directory is created on the first call to Save.
func NewNodeStore(dir string) *NodeStore {
	return &NodeStore{dir: dir}
}

// Dir returns the store's base directory.
func (s *NodeStore) Dir() string {
	return s.dir
}

// List returns all registered nodes.
// Returns an empty slice (not nil) if none exist.
func (s *NodeStore) List(_ context.Context) ([]*domain.Node, error) {
	return s.loadNodes()
}

// Get returns a single node by its stable ID.
// Returns domain.ErrNodeNotFound if not found.
func (s *NodeStore) Get(_ context.Context, id string) (*domain.Node, error) {
	nodes, err := s.loadNodes()
	if err != nil {
		return nil, err
	}
	for _, n := range nodes {
		if n.ID == id {
			return n, nil
		}
	}
	return nil, fmt.Errorf("node %q: %w", id, domain.ErrNodeNotFound)
}

// GetByHostPort returns the node matching the given hostname and port.
// Returns domain.ErrNodeNotFound if no match.
func (s *NodeStore) GetByHostPort(_ context.Context, hostname string, port int) (*domain.Node, error) {
	nodes, err := s.loadNodes()
	if err != nil {
		return nil, err
	}
	for _, n := range nodes {
		if n.Hostname == hostname && n.Port == port {
			return n, nil
		}
	}
	return nil, fmt.Errorf("node at %s:%d: %w", hostname, port, domain.ErrNodeNotFound)
}

// Save persists a node (create or update). If a node with the same ID
// already exists it is replaced; otherwise the node is appended.
func (s *NodeStore) Save(_ context.Context, node *domain.Node) error {
	nodes, err := s.loadNodes()
	if err != nil {
		return err
	}

	found := false
	for i, n := range nodes {
		if n.ID == node.ID {
			nodes[i] = node
			found = true
			break
		}
	}
	if !found {
		nodes = append(nodes, node)
	}

	return s.saveNodes(nodes)
}

// Delete removes a node by its stable ID.
// Returns domain.ErrNodeNotFound if not found.
func (s *NodeStore) Delete(_ context.Context, id string) error {
	nodes, err := s.loadNodes()
	if err != nil {
		return err
	}

	idx := -1
	for i, n := range nodes {
		if n.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("node %q: %w", id, domain.ErrNodeNotFound)
	}

	nodes = append(nodes[:idx], nodes[idx+1:]...)
	return s.saveNodes(nodes)
}

// ---------------------------------------------------------------------------
// Internal persistence
// ---------------------------------------------------------------------------

func (s *NodeStore) nodeFilePath() string {
	return filepath.Join(s.dir, nodeCollection+".json")
}

// loadNodes reads nodes.json. If the file does not exist, an empty
// slice is returned.
func (s *NodeStore) loadNodes() ([]*domain.Node, error) {
	p := s.nodeFilePath()

	b, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []*domain.Node{}, nil
		}
		return nil, fmt.Errorf("jsonstore: read %s: %w", p, err)
	}

	return s.parseNodeFile(b)
}

// parseNodeFile parses a versioned nodes.json payload.
func (s *NodeStore) parseNodeFile(data []byte) ([]*domain.Node, error) {
	var nf nodeFileV1
	if err := json.Unmarshal(data, &nf); err != nil {
		return nil, fmt.Errorf("jsonstore: parse nodes.json: %w", err)
	}

	// Version guard — reject files from the future.
	if nf.Version > currentNodeVersion {
		return nil, fmt.Errorf("jsonstore: nodes.json version %d is newer than supported (%d); please upgrade orbit",
			nf.Version, currentNodeVersion)
	}

	nodes := make([]*domain.Node, len(nf.Nodes))
	for i, d := range nf.Nodes {
		nodes[i] = fromNodeDTO(d)
	}
	return nodes, nil
}

// saveNodes writes the node slice atomically with the versioned envelope.
func (s *NodeStore) saveNodes(nodes []*domain.Node) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return fmt.Errorf("jsonstore: create directory %s: %w", s.dir, err)
	}

	dtos := make([]nodeDTO, len(nodes))
	for i, n := range nodes {
		dtos[i] = toNodeDTO(n)
	}

	nf := nodeFileV1{
		Version: currentNodeVersion,
		Nodes:   dtos,
	}

	b, err := json.MarshalIndent(nf, "", "  ")
	if err != nil {
		return fmt.Errorf("jsonstore: marshal nodes: %w", err)
	}
	b = append(b, '\n')

	// Atomic write: temp file + rename.
	tmp, err := os.CreateTemp(s.dir, nodeCollection+"-*.tmp")
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

	dest := s.nodeFilePath()
	if err := os.Rename(tmpPath, dest); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("jsonstore: rename %s -> %s: %w", tmpPath, dest, err)
	}

	return nil
}
