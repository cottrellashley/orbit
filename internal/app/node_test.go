package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
)

// deterministicUUID returns a closure that yields sequential UUIDs for testing.
func deterministicUUID(prefix string) func() string {
	n := 0
	return func() string {
		n++
		return prefix + "-" + string(rune('0'+n))
	}
}

func TestNodeService_RegisterNode(t *testing.T) {
	orig := uuidFunc
	defer func() { uuidFunc = orig }()
	uuidFunc = deterministicUUID("node")

	store := newMockNodeStore()
	provider := newMockSessionProvider()
	svc := NewNodeService(store, provider)
	ctx := context.Background()

	node, err := svc.RegisterNode(ctx, "agent.example.com", 8080, domain.ProviderOpenCode, "my-agent")
	if err != nil {
		t.Fatalf("RegisterNode: %v", err)
	}
	if node.ID != "node-1" {
		t.Errorf("ID = %q, want %q", node.ID, "node-1")
	}
	if node.Hostname != "agent.example.com" {
		t.Errorf("Hostname = %q, want %q", node.Hostname, "agent.example.com")
	}
	if node.Port != 8080 {
		t.Errorf("Port = %d, want 8080", node.Port)
	}
	if node.Provider != domain.ProviderOpenCode {
		t.Errorf("Provider = %v, want ProviderOpenCode", node.Provider)
	}
	if node.Origin != domain.OriginRegistered {
		t.Errorf("Origin = %v, want OriginRegistered", node.Origin)
	}
	if node.Name != "my-agent" {
		t.Errorf("Name = %q, want %q", node.Name, "my-agent")
	}

	// Verify it was persisted.
	nodes, _ := svc.ListNodes(ctx)
	if len(nodes) != 1 {
		t.Fatalf("ListNodes count = %d, want 1", len(nodes))
	}
}

func TestNodeService_RegisterNode_UpdateExisting(t *testing.T) {
	orig := uuidFunc
	defer func() { uuidFunc = orig }()
	uuidFunc = deterministicUUID("node")

	store := newMockNodeStore()
	provider := newMockSessionProvider()
	svc := NewNodeService(store, provider)
	ctx := context.Background()

	// First registration.
	first, err := svc.RegisterNode(ctx, "localhost", 3000, domain.ProviderOpenCode, "v1")
	if err != nil {
		t.Fatalf("first RegisterNode: %v", err)
	}

	// Second registration at same host:port should update in-place.
	second, err := svc.RegisterNode(ctx, "localhost", 3000, domain.ProviderOpenCode, "v2")
	if err != nil {
		t.Fatalf("second RegisterNode: %v", err)
	}

	if second.ID != first.ID {
		t.Errorf("ID changed: %q -> %q, want same", first.ID, second.ID)
	}
	if second.Name != "v2" {
		t.Errorf("Name = %q, want %q", second.Name, "v2")
	}

	// Still only one node in store.
	nodes, _ := svc.ListNodes(ctx)
	if len(nodes) != 1 {
		t.Fatalf("ListNodes count = %d, want 1", len(nodes))
	}
}

func TestNodeService_RegisterNode_Validation(t *testing.T) {
	store := newMockNodeStore()
	provider := newMockSessionProvider()
	svc := NewNodeService(store, provider)
	ctx := context.Background()

	tests := []struct {
		name     string
		hostname string
		port     int
	}{
		{"empty hostname", "", 8080},
		{"port zero", "localhost", 0},
		{"port negative", "localhost", -1},
		{"port too high", "localhost", 70000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.RegisterNode(ctx, tt.hostname, tt.port, domain.ProviderOpenCode, "")
			if err == nil {
				t.Errorf("expected error for %s", tt.name)
			}
		})
	}
}

func TestNodeService_RemoveNode(t *testing.T) {
	existing := &domain.Node{
		ID:       "abc-123",
		Hostname: "localhost",
		Port:     3000,
	}
	store := newMockNodeStore(existing)
	svc := NewNodeService(store, newMockSessionProvider())
	ctx := context.Background()

	if err := svc.RemoveNode(ctx, "abc-123"); err != nil {
		t.Fatalf("RemoveNode: %v", err)
	}

	nodes, _ := svc.ListNodes(ctx)
	if len(nodes) != 0 {
		t.Errorf("ListNodes count = %d, want 0", len(nodes))
	}
}

func TestNodeService_RemoveNode_NotFound(t *testing.T) {
	store := newMockNodeStore()
	svc := NewNodeService(store, newMockSessionProvider())
	ctx := context.Background()

	err := svc.RemoveNode(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for removing nonexistent node")
	}
	if !errors.Is(err, domain.ErrNodeNotFound) {
		t.Errorf("error = %v, want wrapping ErrNodeNotFound", err)
	}
}

func TestNodeService_GetNode(t *testing.T) {
	existing := &domain.Node{
		ID:       "abc-123",
		Hostname: "localhost",
		Port:     3000,
		Name:     "test-node",
	}
	store := newMockNodeStore(existing)
	svc := NewNodeService(store, newMockSessionProvider())
	ctx := context.Background()

	node, err := svc.GetNode(ctx, "abc-123")
	if err != nil {
		t.Fatalf("GetNode: %v", err)
	}
	if node.Name != "test-node" {
		t.Errorf("Name = %q, want %q", node.Name, "test-node")
	}
}

func TestNodeService_GetNode_NotFound(t *testing.T) {
	store := newMockNodeStore()
	svc := NewNodeService(store, newMockSessionProvider())
	ctx := context.Background()

	_, err := svc.GetNode(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, domain.ErrNodeNotFound) {
		t.Errorf("error = %v, want wrapping ErrNodeNotFound", err)
	}
}

func TestNodeService_SyncDiscoveredNodes_NewDiscovery(t *testing.T) {
	orig := uuidFunc
	defer func() { uuidFunc = orig }()
	uuidFunc = deterministicUUID("disc")

	store := newMockNodeStore()
	provider := newMockSessionProvider()
	provider.servers = []domain.Server{
		{PID: 100, Port: 3000, Hostname: "127.0.0.1", Directory: "/home/user/project", Version: "0.5.0", Healthy: true},
	}
	svc := NewNodeService(store, provider)
	ctx := context.Background()

	nodes, err := svc.SyncDiscoveredNodes(ctx)
	if err != nil {
		t.Fatalf("SyncDiscoveredNodes: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("got %d nodes, want 1", len(nodes))
	}

	n := nodes[0]
	if n.Origin != domain.OriginDiscovered {
		t.Errorf("Origin = %v, want OriginDiscovered", n.Origin)
	}
	if n.PID != 100 {
		t.Errorf("PID = %d, want 100", n.PID)
	}
	if !n.Healthy {
		t.Error("Healthy = false, want true")
	}
}

func TestNodeService_SyncDiscoveredNodes_UpdatesExisting(t *testing.T) {
	past := time.Now().Add(-2 * time.Second)
	existing := &domain.Node{
		ID:        "existing-1",
		Hostname:  "127.0.0.1",
		Port:      3000,
		Origin:    domain.OriginDiscovered,
		Version:   "0.4.0",
		PID:       99,
		Healthy:   true,
		CreatedAt: past,
		UpdatedAt: past,
	}
	store := newMockNodeStore(existing)
	provider := newMockSessionProvider()
	provider.servers = []domain.Server{
		{PID: 200, Port: 3000, Hostname: "127.0.0.1", Directory: "/updated/dir", Version: "0.5.0", Healthy: true},
	}
	svc := NewNodeService(store, provider)
	ctx := context.Background()

	nodes, err := svc.SyncDiscoveredNodes(ctx)
	if err != nil {
		t.Fatalf("SyncDiscoveredNodes: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("got %d nodes, want 1", len(nodes))
	}

	n := nodes[0]
	if n.ID != "existing-1" {
		t.Errorf("ID = %q, want %q (should preserve existing)", n.ID, "existing-1")
	}
	if n.PID != 200 {
		t.Errorf("PID = %d, want 200 (should update)", n.PID)
	}
	if n.Version != "0.5.0" {
		t.Errorf("Version = %q, want %q", n.Version, "0.5.0")
	}
	if n.Directory != "/updated/dir" {
		t.Errorf("Directory = %q, want %q", n.Directory, "/updated/dir")
	}
	if !n.UpdatedAt.After(past) {
		t.Error("UpdatedAt should have been refreshed")
	}
}

func TestNodeService_SyncDiscoveredNodes_MarksStaleUnhealthy(t *testing.T) {
	past := time.Now().Add(-2 * time.Second)
	staleNode := &domain.Node{
		ID:        "stale-1",
		Hostname:  "127.0.0.1",
		Port:      9999,
		Origin:    domain.OriginDiscovered,
		Healthy:   true,
		CreatedAt: past,
		UpdatedAt: past,
	}
	store := newMockNodeStore(staleNode)
	// Provider returns NO servers — the stale node should be marked unhealthy.
	provider := newMockSessionProvider()
	provider.servers = []domain.Server{}

	svc := NewNodeService(store, provider)
	ctx := context.Background()

	nodes, err := svc.SyncDiscoveredNodes(ctx)
	if err != nil {
		t.Fatalf("SyncDiscoveredNodes: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("got %d nodes, want 1", len(nodes))
	}
	if nodes[0].Healthy {
		t.Error("stale discovered node should be marked unhealthy")
	}
}

func TestNodeService_SyncDiscoveredNodes_LeavesRegisteredAlone(t *testing.T) {
	past := time.Now().Add(-2 * time.Second)
	registeredNode := &domain.Node{
		ID:        "reg-1",
		Hostname:  "remote.example.com",
		Port:      8080,
		Origin:    domain.OriginRegistered,
		Healthy:   true,
		CreatedAt: past,
		UpdatedAt: past,
	}
	store := newMockNodeStore(registeredNode)
	// Provider returns NO servers — but registered nodes should NOT be touched.
	provider := newMockSessionProvider()
	provider.servers = []domain.Server{}

	svc := NewNodeService(store, provider)
	ctx := context.Background()

	nodes, err := svc.SyncDiscoveredNodes(ctx)
	if err != nil {
		t.Fatalf("SyncDiscoveredNodes: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("got %d nodes, want 1", len(nodes))
	}
	if !nodes[0].Healthy {
		t.Error("registered node should remain healthy (not touched by sync)")
	}
}

func TestNodeService_SyncDiscoveredNodes_DiscoverError(t *testing.T) {
	store := newMockNodeStore()
	provider := newMockSessionProvider()
	provider.discoverErr = errors.New("process scan failed")

	svc := NewNodeService(store, provider)
	ctx := context.Background()

	_, err := svc.SyncDiscoveredNodes(ctx)
	if err == nil {
		t.Fatal("expected error when discovery fails")
	}
}

func TestNodeService_SyncDiscoveredNodes_MultipleServers(t *testing.T) {
	orig := uuidFunc
	defer func() { uuidFunc = orig }()
	uuidFunc = deterministicUUID("multi")

	store := newMockNodeStore()
	provider := newMockSessionProvider()
	provider.servers = []domain.Server{
		{PID: 10, Port: 3000, Hostname: "127.0.0.1", Directory: "/a", Version: "1.0", Healthy: true},
		{PID: 20, Port: 3001, Hostname: "127.0.0.1", Directory: "/b", Version: "1.0", Healthy: true},
		{PID: 30, Port: 3002, Hostname: "127.0.0.1", Directory: "/c", Version: "1.0", Healthy: false},
	}
	svc := NewNodeService(store, provider)
	ctx := context.Background()

	nodes, err := svc.SyncDiscoveredNodes(ctx)
	if err != nil {
		t.Fatalf("SyncDiscoveredNodes: %v", err)
	}
	if len(nodes) != 3 {
		t.Fatalf("got %d nodes, want 3", len(nodes))
	}

	// Third node should be unhealthy (discovered as unhealthy).
	for _, n := range nodes {
		if n.Port == 3002 && n.Healthy {
			t.Error("node on port 3002 should be unhealthy")
		}
	}
}
