package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyCreatesDirectoryAndFiles(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "project")

	tmpl := &Template{
		Files: map[string]string{
			"config.json": "{}\n",
			"README.md":   "# Test\n",
		},
		Dirs: []string{
			"src",
			"src/lib",
		},
	}

	if err := Apply(dir, tmpl); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Verify directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("expected project directory to exist")
	}

	// Verify subdirectories
	for _, d := range tmpl.Dirs {
		path := filepath.Join(dir, d)
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			t.Errorf("expected directory %s to exist", d)
		}
		if err == nil && !info.IsDir() {
			t.Errorf("expected %s to be a directory", d)
		}
	}

	// Verify files
	for name, expectedContent := range tmpl.Files {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("cannot read file %s: %v", name, err)
			continue
		}
		if string(data) != expectedContent {
			t.Errorf("file %s: expected %q, got %q", name, expectedContent, string(data))
		}
	}
}

func TestApplyDoesNotOverwriteExistingFiles(t *testing.T) {
	dir := t.TempDir()

	// Create an existing file
	existingPath := filepath.Join(dir, "config.json")
	os.WriteFile(existingPath, []byte("existing content"), 0644)

	tmpl := &Template{
		Files: map[string]string{
			"config.json": "new content",
		},
	}

	if err := Apply(dir, tmpl); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Verify existing file was not overwritten
	data, _ := os.ReadFile(existingPath)
	if string(data) != "existing content" {
		t.Errorf("expected existing file to be preserved, got %q", string(data))
	}
}

func TestApplyEmptyTemplate(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "empty")

	tmpl := &Template{}
	if err := Apply(dir, tmpl); err != nil {
		t.Fatalf("Apply with empty template failed: %v", err)
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("expected directory to be created even with empty template")
	}
}

func TestOpenCodeTemplate(t *testing.T) {
	tmpl := OpenCodeTemplate()

	if tmpl == nil {
		t.Fatal("OpenCodeTemplate returned nil")
	}

	if _, ok := tmpl.Files["opencode.json"]; !ok {
		t.Error("expected opencode.json in template files")
	}

	if _, ok := tmpl.Files["AGENTS.md"]; !ok {
		t.Error("expected AGENTS.md in template files")
	}

	expectedDirs := []string{
		".opencode/commands",
		".opencode/skills",
		".opencode/agents",
	}
	for _, d := range expectedDirs {
		found := false
		for _, td := range tmpl.Dirs {
			if td == d {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected directory %q in template", d)
		}
	}
}

func TestApplyOpenCodeTemplate(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "project")
	tmpl := OpenCodeTemplate()

	if err := Apply(dir, tmpl); err != nil {
		t.Fatalf("Apply OpenCode template failed: %v", err)
	}

	// Verify all expected paths exist
	paths := []string{
		"opencode.json",
		"AGENTS.md",
		".opencode/commands",
		".opencode/skills",
		".opencode/agents",
	}

	for _, p := range paths {
		full := filepath.Join(dir, p)
		if _, err := os.Stat(full); os.IsNotExist(err) {
			t.Errorf("expected %s to exist after scaffold", p)
		}
	}
}
