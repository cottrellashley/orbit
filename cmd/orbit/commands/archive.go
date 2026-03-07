package commands

import (
	"fmt"
	"path/filepath"

	"github.com/cottrellashley/orbit/internal/archive"
	"github.com/cottrellashley/orbit/internal/role"
	"github.com/spf13/cobra"
)

// ArchiveCmd creates the `orbit archive` command.
func ArchiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive <role>[/subproject]",
		Short: "Archive a role or workspace project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]

			cfg, cfgPath, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			roleName, subproject := parseTarget(target)

			r, err := cfg.FindRole(roleName)
			if err != nil {
				return err
			}

			switch {
			case r.Type == role.Environment:
				// Archive the entire environment and remove from config
				dest, err := archive.Move(r.Path, cfg.ArchivePath)
				if err != nil {
					return err
				}
				cfg.RemoveRole(roleName)
				if err := saveConfig(cfgPath, cfg); err != nil {
					return fmt.Errorf("cannot save config: %w", err)
				}
				fmt.Printf("Archived %s to %s\n", roleName, dest)

			case r.Type == role.Workspace && subproject != "":
				// Archive a single project within the workspace
				projectPath := filepath.Join(r.Path, subproject)
				dest, err := archive.Move(projectPath, cfg.ArchivePath)
				if err != nil {
					return err
				}
				fmt.Printf("Archived %s/%s to %s\n", roleName, subproject, dest)

			case r.Type == role.Workspace && subproject == "":
				// Archive the entire workspace and remove from config
				dest, err := archive.Move(r.Path, cfg.ArchivePath)
				if err != nil {
					return err
				}
				cfg.RemoveRole(roleName)
				if err := saveConfig(cfgPath, cfg); err != nil {
					return fmt.Errorf("cannot save config: %w", err)
				}
				fmt.Printf("Archived workspace %s to %s\n", roleName, dest)

			default:
				return fmt.Errorf("unexpected role type: %s", r.Type)
			}

			return nil
		},
	}

	return cmd
}
