package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
)

// Template defines what files to create when scaffolding a directory.
type Template struct {
	// Files maps relative path to file content.
	Files map[string]string
	// Dirs lists directories to create (relative paths).
	Dirs []string
}

// OpenCodeTemplate returns the default scaffold template for OpenCode environments.
func OpenCodeTemplate() *Template {
	return &Template{
		Files: map[string]string{
			"opencode.json": "{}\n",
			"AGENTS.md":     "# Instructions\n\n<!-- Define agent behavior here -->\n",
		},
		Dirs: []string{
			".opencode/commands",
			".opencode/skills",
			".opencode/agents",
		},
	}
}

// Apply creates the directory structure and files defined by the template.
// It will not overwrite existing files.
func Apply(dir string, t *Template) error {
	// Create the root directory
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create directory %s: %w", dir, err)
	}

	// Create subdirectories
	for _, d := range t.Dirs {
		path := filepath.Join(dir, d)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("cannot create directory %s: %w", path, err)
		}
	}

	// Create files (skip if they already exist)
	for name, content := range t.Files {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			// File exists, skip
			continue
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("cannot write file %s: %w", path, err)
		}
	}

	return nil
}
