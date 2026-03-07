package commands

import (
	"fmt"

	"github.com/cottrellashley/orbit/internal/role"
	"github.com/cottrellashley/orbit/internal/scaffold"
	"github.com/spf13/cobra"
)

// InitCmd creates the `orbit init` command.
func InitCmd() *cobra.Command {
	var (
		roleType    string
		rolePath    string
		tags        []string
		adapterName string
	)

	cmd := &cobra.Command{
		Use:   "init <name>",
		Short: "Register a new role and scaffold its directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Validate type
			rt := role.Type(roleType)
			if rt != role.Environment && rt != role.Workspace {
				return fmt.Errorf("invalid type %q: must be 'environment' or 'workspace'", roleType)
			}

			if rolePath == "" {
				return fmt.Errorf("--path is required")
			}

			cfg, cfgPath, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			r := &role.Role{
				Name:    name,
				Type:    rt,
				Path:    rolePath,
				Adapter: adapterName,
				Tags:    tags,
			}

			if err := cfg.AddRole(r); err != nil {
				return err
			}

			// Scaffold the directory
			if rt == role.Environment {
				tmpl := scaffold.OpenCodeTemplate()
				if err := scaffold.Apply(rolePath, tmpl); err != nil {
					return fmt.Errorf("scaffold failed: %w", err)
				}
				fmt.Printf("Scaffolded environment at %s\n", rolePath)
			} else {
				// For workspace, just create the parent directory
				if err := scaffold.Apply(rolePath, &scaffold.Template{}); err != nil {
					return fmt.Errorf("cannot create workspace directory: %w", err)
				}
				fmt.Printf("Created workspace at %s\n", rolePath)
			}

			if err := saveConfig(cfgPath, cfg); err != nil {
				return fmt.Errorf("cannot save config: %w", err)
			}

			fmt.Printf("Registered role %q (%s)\n", name, rt)
			return nil
		},
	}

	cmd.Flags().StringVar(&roleType, "type", "", "role type: environment or workspace (required)")
	cmd.Flags().StringVar(&rolePath, "path", "", "directory path for the role (required)")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "tags for the role")
	cmd.Flags().StringVar(&adapterName, "adapter", "", "adapter name (uses default if omitted)")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("path")

	return cmd
}
