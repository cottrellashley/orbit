package commands

import (
	"fmt"

	"github.com/cottrellashley/orbit/internal/role"
	"github.com/cottrellashley/orbit/internal/scaffold"
	"github.com/spf13/cobra"
)

// NewCmd creates the `orbit new` command.
func NewCmd() *cobra.Command {
	var open bool

	cmd := &cobra.Command{
		Use:   "new <workspace-role> <name>",
		Short: "Create a new project in a workspace",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			roleName := args[0]
			projectName := args[1]

			cfg, _, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			r, err := cfg.FindRole(roleName)
			if err != nil {
				return err
			}

			if r.Type != role.Workspace {
				return fmt.Errorf("role %q is not a workspace — use 'orbit init' for environments", roleName)
			}

			projectPath := r.Path + "/" + projectName

			// Scaffold the project with OpenCode template
			tmpl := scaffold.OpenCodeTemplate()
			if err := scaffold.Apply(projectPath, tmpl); err != nil {
				return fmt.Errorf("scaffold failed: %w", err)
			}

			fmt.Printf("Created project %s at %s\n", projectName, projectPath)

			if open {
				a, err := cfg.ResolveAdapter(r)
				if err != nil {
					return err
				}
				fmt.Printf("Launching %s in %s\n", a.Name, projectPath)
				return a.Exec(projectPath)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&open, "open", false, "open the project immediately after creation")

	return cmd
}
