package archive

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMove(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source directory with a file
	source := filepath.Join(tmpDir, "myproject")
	os.MkdirAll(source, 0755)
	os.WriteFile(filepath.Join(source, "file.txt"), []byte("hello"), 0644)

	archiveDir := filepath.Join(tmpDir, "archive")

	dest, err := Move(source, archiveDir)
	if err != nil {
		t.Fatalf("Move failed: %v", err)
	}

	// Source should no longer exist
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Error("expected source directory to be removed after move")
	}

	// Destination should exist
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		t.Error("expected destination to exist")
	}

	// Destination name should contain source name and timestamp
	base := filepath.Base(dest)
	if !strings.HasPrefix(base, "myproject-") {
		t.Errorf("expected dest name to start with 'myproject-', got %q", base)
	}

	// File should be in the new location
	data, err := os.ReadFile(filepath.Join(dest, "file.txt"))
	if err != nil {
		t.Fatalf("cannot read moved file: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("expected file content %q, got %q", "hello", string(data))
	}
}

func TestMoveCreatesArchiveDir(t *testing.T) {
	tmpDir := t.TempDir()

	source := filepath.Join(tmpDir, "project")
	os.MkdirAll(source, 0755)

	archiveDir := filepath.Join(tmpDir, "nested", "archive")

	_, err := Move(source, archiveDir)
	if err != nil {
		t.Fatalf("Move failed: %v", err)
	}

	// Archive directory should have been created
	if _, err := os.Stat(archiveDir); os.IsNotExist(err) {
		t.Error("expected archive directory to be created")
	}
}

func TestMoveMissingSource(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := Move(filepath.Join(tmpDir, "nonexistent"), filepath.Join(tmpDir, "archive"))
	if err == nil {
		t.Error("expected error for missing source directory")
	}
}
