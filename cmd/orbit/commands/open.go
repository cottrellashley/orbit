package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cottrellashley/orbit/internal/role"
	"github.com/spf13/cobra"
)

// OpenCmd creates the `orbit open` command.
func OpenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open <role>[/subproject]",
		Short: "Launch an adapter in a role's directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]

			cfg, _, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			// Parse role/subproject
			roleName, subproject := parseTarget(target)

			r, err := cfg.FindRole(roleName)
			if err != nil {
				return err
			}

			a, err := cfg.ResolveAdapter(r)
			if err != nil {
				return err
			}

			var dir string

			switch r.Type {
			case role.Environment:
				if subproject != "" {
					return fmt.Errorf("role %q is an environment, not a workspace — cannot specify a subproject", roleName)
				}
				dir = r.Path

			case role.Workspace:
				if subproject != "" {
					dir = filepath.Join(r.Path, subproject)
				} else {
					// List projects and prompt user to pick
					sub, err := pickSubproject(r.Path)
					if err != nil {
						return err
					}
					dir = filepath.Join(r.Path, sub)
				}
			}

			// Verify directory exists
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				return fmt.Errorf("directory does not exist: %s", dir)
			}

			fmt.Printf("Launching %s in %s\n", a.Name, dir)
			return a.Exec(dir)
		},
	}

	return cmd
}

// parseTarget splits "rolename/subproject" into parts.
func parseTarget(target string) (string, string) {
	parts := strings.SplitN(target, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}

// pickSubproject lists children of a workspace directory and prompts the user.
func pickSubproject(workspacePath string) (string, error) {
	entries, err := os.ReadDir(workspacePath)
	if err != nil {
		return "", fmt.Errorf("cannot read workspace %s: %w", workspacePath, err)
	}

	var projects []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			projects = append(projects, e.Name())
		}
	}

	if len(projects) == 0 {
		return "", fmt.Errorf("no projects found in %s", workspacePath)
	}

	fmt.Println("Projects:")
	for i, p := range projects {
		fmt.Printf("  [%d] %s\n", i+1, p)
	}

	fmt.Print("Select project: ")
	var choice int
	if _, err := fmt.Scan(&choice); err != nil {
		return "", fmt.Errorf("invalid selection: %w", err)
	}

	if choice < 1 || choice > len(projects) {
		return "", fmt.Errorf("selection %d out of range", choice)
	}

	return projects[choice-1], nil
}
