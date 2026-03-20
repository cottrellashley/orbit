package jsonstore_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cottrellashley/orbit/internal/adapter/jsonstore"
	"github.com/cottrellashley/orbit/internal/domain"
	"github.com/cottrellashley/orbit/internal/port"
)

// Verify NodeStore satisfies port.NodeStore at compile time.
var _ port.NodeStore = (*jsonstore.NodeStore)(nil)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func testNode(id, name, hostname string, port int) *domain.Node {
	now := time.Now().Add(-2 * time.Second).Truncate(time.Second)
	return &domain.Node{
		ID:        id,
		Name:      name,
		Provider:  domain.ProviderOpenCode,
		Origin:    domain.OriginRegistered,
		Hostname:  hostname,
		Port:      port,
		Directory: "/home/user/code",
		Version:   "0.1.0",
		PID:       0,
		Healthy:   true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// ---------------------------------------------------------------------------
// Basic CRUD operations
// ---------------------------------------------------------------------------

func TestNodeStore_ListEmpty(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.NewNodeStore(dir)

	nodes, err := s.List(context.Background())
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(nodes) != 0 {
		t.Errorf("List() returned %d nodes, want 0", len(nodes))
	}
}

func TestNodeStore_SaveAndList(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.NewNodeStore(dir)
	ctx := context.Background()

	n := testNode("node-001", "local-agent", "127.0.0.1", 3000)

	if err := s.Save(ctx, n); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("List() returned %d nodes, want 1", len(loaded))
	}

	got := loaded[0]
	if got.ID != "node-001" {
		t.Errorf("ID = %q, want %q", got.ID, "node-001")
	}
	if got.Name != "local-agent" {
		t.Errorf("Name = %q, want %q", got.Name, "local-agent")
	}
	if got.Provider != domain.ProviderOpenCode {
		t.Errorf("Provider = %v, want ProviderOpenCode", got.Provider)
	}
	if got.Origin != domain.OriginRegistered {
		t.Errorf("Origin = %v, want OriginRegistered", got.Origin)
	}
	if got.Hostname != "127.0.0.1" {
		t.Errorf("Hostname = %q, want %q", got.Hostname, "127.0.0.1")
	}
	if got.Port != 3000 {
		t.Errorf("Port = %d, want %d", got.Port, 3000)
	}
	if !got.Healthy {
		t.Error("Healthy = false, want true")
	}
}

func TestNodeStore_SaveUpdatesExisting(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.NewNodeStore(dir)
	ctx := context.Background()

	n := testNode("node-001", "original", "127.0.0.1", 3000)
	if err := s.Save(ctx, n); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Update the node.
	n.Name = "updated"
	n.Port = 4000
	n.Healthy = false
	if err := s.Save(ctx, n); err != nil {
		t.Fatalf("Save() update error: %v", err)
	}

	loaded, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("List() returned %d nodes, want 1 (should update, not duplicate)", len(loaded))
	}
	if loaded[0].Name != "updated" {
		t.Errorf("Name = %q, want %q", loaded[0].Name, "updated")
	}
	if loaded[0].Port != 4000 {
		t.Errorf("Port = %d, want %d", loaded[0].Port, 4000)
	}
	if loaded[0].Healthy {
		t.Error("Healthy = true, want false")
	}
}

func TestNodeStore_Get(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.NewNodeStore(dir)
	ctx := context.Background()

	n1 := testNode("node-001", "alpha", "127.0.0.1", 3000)
	n2 := testNode("node-002", "beta", "10.0.0.5", 4000)

	if err := s.Save(ctx, n1); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if err := s.Save(ctx, n2); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	got, err := s.Get(ctx, "node-002")
	if err != nil {
		t.Fatalf("Get(node-002) error: %v", err)
	}
	if got.Name != "beta" {
		t.Errorf("Name = %q, want %q", got.Name, "beta")
	}

	_, err = s.Get(ctx, "nonexistent")
	if !errors.Is(err, domain.ErrNodeNotFound) {
		t.Errorf("Get(nonexistent) error = %v, want ErrNodeNotFound", err)
	}
}

func TestNodeStore_GetByHostPort(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.NewNodeStore(dir)
	ctx := context.Background()

	n := testNode("node-001", "alpha", "10.0.0.5", 8080)
	if err := s.Save(ctx, n); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	got, err := s.GetByHostPort(ctx, "10.0.0.5", 8080)
	if err != nil {
		t.Fatalf("GetByHostPort() error: %v", err)
	}
	if got.ID != "node-001" {
		t.Errorf("ID = %q, want %q", got.ID, "node-001")
	}

	_, err = s.GetByHostPort(ctx, "10.0.0.5", 9999)
	if !errors.Is(err, domain.ErrNodeNotFound) {
		t.Errorf("GetByHostPort(wrong port) error = %v, want ErrNodeNotFound", err)
	}

	_, err = s.GetByHostPort(ctx, "192.168.1.1", 8080)
	if !errors.Is(err, domain.ErrNodeNotFound) {
		t.Errorf("GetByHostPort(wrong host) error = %v, want ErrNodeNotFound", err)
	}
}

func TestNodeStore_Delete(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.NewNodeStore(dir)
	ctx := context.Background()

	n1 := testNode("node-001", "alpha", "127.0.0.1", 3000)
	n2 := testNode("node-002", "beta", "127.0.0.1", 4000)

	if err := s.Save(ctx, n1); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if err := s.Save(ctx, n2); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	if err := s.Delete(ctx, "node-001"); err != nil {
		t.Fatalf("Delete(node-001) error: %v", err)
	}

	remaining, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(remaining) != 1 {
		t.Fatalf("List() returned %d nodes, want 1", len(remaining))
	}
	if remaining[0].ID != "node-002" {
		t.Errorf("remaining[0].ID = %q, want %q", remaining[0].ID, "node-002")
	}

	err = s.Delete(ctx, "nonexistent")
	if !errors.Is(err, domain.ErrNodeNotFound) {
		t.Errorf("Delete(nonexistent) error = %v, want ErrNodeNotFound", err)
	}
}

// ---------------------------------------------------------------------------
// Versioned format
// ---------------------------------------------------------------------------

func TestNodeStore_VersionedFormat(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.NewNodeStore(dir)
	ctx := context.Background()

	n := testNode("node-001", "test", "127.0.0.1", 3000)
	if err := s.Save(ctx, n); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Read raw JSON and verify version field exists.
	raw, err := os.ReadFile(filepath.Join(dir, "nodes.json"))
	if err != nil {
		t.Fatalf("read nodes.json: %v", err)
	}

	var envelope struct {
		Version int `json:"version"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("parse nodes.json: %v", err)
	}
	if envelope.Version != 1 {
		t.Errorf("version = %d, want 1", envelope.Version)
	}
}

func TestNodeStore_RejectsFutureVersion(t *testing.T) {
	dir := t.TempDir()

	data := `{"version": 999, "nodes": []}`
	if err := os.WriteFile(filepath.Join(dir, "nodes.json"), []byte(data), 0o644); err != nil {
		t.Fatalf("write nodes.json: %v", err)
	}

	s := jsonstore.NewNodeStore(dir)
	_, err := s.List(context.Background())
	if err == nil {
		t.Fatal("List() should error on future version")
	}
}

// ---------------------------------------------------------------------------
// Atomic write safety
// ---------------------------------------------------------------------------

func TestNodeStore_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.NewNodeStore(dir)
	ctx := context.Background()

	n := testNode("node-001", "first", "127.0.0.1", 3000)
	if err := s.Save(ctx, n); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify no leftover temp files.
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("leftover temp file: %s", e.Name())
		}
	}
}

// ---------------------------------------------------------------------------
// Provider and Origin round-trip
// ---------------------------------------------------------------------------

func TestNodeStore_ProviderRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.NewNodeStore(dir)
	ctx := context.Background()

	tests := []struct {
		provider domain.NodeProvider
	}{
		{domain.ProviderUnknown},
		{domain.ProviderOpenCode},
	}

	for _, tt := range tests {
		t.Run(tt.provider.String(), func(t *testing.T) {
			n := testNode("prov-rt", "rt", "127.0.0.1", 3000)
			n.Provider = tt.provider
			if err := s.Save(ctx, n); err != nil {
				t.Fatalf("Save() error: %v", err)
			}

			got, err := s.Get(ctx, "prov-rt")
			if err != nil {
				t.Fatalf("Get() error: %v", err)
			}
			if got.Provider != tt.provider {
				t.Errorf("Provider = %v, want %v", got.Provider, tt.provider)
			}
		})
	}
}

func TestNodeStore_OriginRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.NewNodeStore(dir)
	ctx := context.Background()

	tests := []struct {
		origin domain.NodeOrigin
	}{
		{domain.OriginDiscovered},
		{domain.OriginRegistered},
	}

	for _, tt := range tests {
		t.Run(tt.origin.String(), func(t *testing.T) {
			n := testNode("orig-rt", "rt", "127.0.0.1", 3000)
			n.Origin = tt.origin
			if err := s.Save(ctx, n); err != nil {
				t.Fatalf("Save() error: %v", err)
			}

			got, err := s.Get(ctx, "orig-rt")
			if err != nil {
				t.Fatalf("Get() error: %v", err)
			}
			if got.Origin != tt.origin {
				t.Errorf("Origin = %v, want %v", got.Origin, tt.origin)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Password is NOT persisted (security)
// ---------------------------------------------------------------------------

func TestNodeStore_PasswordNotPersisted(t *testing.T) {
	dir := t.TempDir()
	s := jsonstore.NewNodeStore(dir)
	ctx := context.Background()

	n := testNode("node-secret", "secret-agent", "127.0.0.1", 3000)
	n.Password = "super-secret-token"
	if err := s.Save(ctx, n); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Password should NOT appear in the JSON file.
	raw, err := os.ReadFile(filepath.Join(dir, "nodes.json"))
	if err != nil {
		t.Fatalf("read nodes.json: %v", err)
	}

	var envelope struct {
		Nodes []map[string]any `json:"nodes"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("parse nodes.json: %v", err)
	}
	if len(envelope.Nodes) != 1 {
		t.Fatalf("expected 1 node in file, got %d", len(envelope.Nodes))
	}

	// The password field must not appear in serialized JSON because
	// the nodeDTO deliberately excludes it.
	if _, ok := envelope.Nodes[0]["password"]; ok {
		t.Error("password field found in nodes.json — secrets must not be persisted in the node store")
	}

	// Round-tripped node should have empty password (not persisted).
	got, err := s.Get(ctx, "node-secret")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got.Password != "" {
		t.Errorf("Password = %q after round-trip, want empty (not persisted)", got.Password)
	}
}
