package opencode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
)

// ---------------------------------------------------------------------------
// JSON DTO — keeps serialization details out of the domain
// ---------------------------------------------------------------------------

type managedServerDTO struct {
	PID       int       `json:"pid"`
	Port      int       `json:"port"`
	Hostname  string    `json:"hostname"`
	Password  string    `json:"password,omitempty"`
	Directory string    `json:"directory"`
	Version   string    `json:"version,omitempty"`
	StartedAt time.Time `json:"started_at"`
}

func toManagedDTO(ms domain.ManagedServer) managedServerDTO {
	return managedServerDTO{
		PID:       ms.PID,
		Port:      ms.Port,
		Hostname:  ms.Hostname,
		Password:  ms.Password,
		Directory: ms.Directory,
		Version:   ms.Version,
		StartedAt: ms.StartedAt,
	}
}

func fromManagedDTO(d managedServerDTO) domain.ManagedServer {
	return domain.ManagedServer{
		PID:       d.PID,
		Port:      d.Port,
		Hostname:  d.Hostname,
		Password:  d.Password,
		Directory: d.Directory,
		Version:   d.Version,
		StartedAt: d.StartedAt,
	}
}

// ---------------------------------------------------------------------------
// State file — tracks managed OpenCode servers across Orbit restarts
// ---------------------------------------------------------------------------

// stateFile is the on-disk JSON format for tracked managed servers.
type stateFile struct {
	Servers []managedServerDTO `json:"servers"`
}

// defaultStateDir returns ~/.local/state/orbit.
func defaultStateDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "state", "orbit")
}

// defaultStatePath returns the full path to the servers.json state file.
func defaultStatePath() string {
	return filepath.Join(defaultStateDir(), "servers.json")
}

// readState reads and parses the state file. Returns an empty stateFile
// if the file does not exist.
func readState(path string) (*stateFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &stateFile{}, nil
		}
		return nil, fmt.Errorf("read state file %s: %w", path, err)
	}

	var sf stateFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return nil, fmt.Errorf("parse state file %s: %w", path, err)
	}
	return &sf, nil
}

// writeState writes the state file atomically (write to temp + rename).
func writeState(path string, sf *stateFile) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create state dir %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	data = append(data, '\n')

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write temp state file: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("rename state file: %w", err)
	}

	return nil
}

// addServer adds or updates a managed server entry in the state file.
func addServer(path string, ms domain.ManagedServer) error {
	sf, err := readState(path)
	if err != nil {
		sf = &stateFile{}
	}

	dto := toManagedDTO(ms)

	// Replace existing entry for same PID, or append.
	found := false
	for i, s := range sf.Servers {
		if s.PID == ms.PID {
			sf.Servers[i] = dto
			found = true
			break
		}
	}
	if !found {
		sf.Servers = append(sf.Servers, dto)
	}

	return writeState(path, sf)
}

// removeServer removes the entry with the given PID from the state file.
func removeServer(path string, pid int) error {
	sf, err := readState(path)
	if err != nil {
		return nil // nothing to remove
	}

	filtered := sf.Servers[:0]
	for _, s := range sf.Servers {
		if s.PID != pid {
			filtered = append(filtered, s)
		}
	}
	sf.Servers = filtered

	return writeState(path, sf)
}

// reapStale removes entries for processes that are no longer alive.
// Returns the cleaned state.
func reapStale(path string) (*stateFile, error) {
	sf, err := readState(path)
	if err != nil {
		return &stateFile{}, nil
	}

	alive := sf.Servers[:0]
	for _, s := range sf.Servers {
		if processAlive(s.PID) {
			alive = append(alive, s)
		}
	}
	sf.Servers = alive

	if err := writeState(path, sf); err != nil {
		return sf, err
	}
	return sf, nil
}

// findServer returns the first managed server entry, or nil if none.
func findServer(sf *stateFile) *domain.ManagedServer {
	if len(sf.Servers) == 0 {
		return nil
	}
	ms := fromManagedDTO(sf.Servers[0])
	return &ms
}

// nowUTC returns the current time in UTC, truncated to seconds for clean
// JSON output.
func nowUTC() time.Time {
	return time.Now().UTC().Truncate(time.Second)
}
