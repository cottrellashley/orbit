package commands

import (
	"fmt"
	"os"

	"github.com/cottrellashley/orbit/internal/role"
	"github.com/spf13/cobra"
)

// StatusCmd creates the `orbit status` command.
func StatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [role]",
		Short: "Show status of orbit or a specific role",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, cfgPath, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			if len(args) == 0 {
				// Overall status
				fmt.Printf("Config:   %s\n", cfgPath)
				fmt.Printf("Archive:  %s\n", cfg.ArchivePath)
				fmt.Printf("Roles:    %d\n", len(cfg.Roles))
				fmt.Printf("Adapters: %d\n", len(cfg.Adapters))

				for _, a := range cfg.Adapters {
					def := ""
					if a.Default {
						def = " (default)"
					}
					fmt.Printf("  - %s → %s%s\n", a.Name, a.Command, def)
				}
				return nil
			}

			// Specific role status
			r, err := cfg.FindRole(args[0])
			if err != nil {
				return err
			}

			exists := "yes"
			if _, err := os.Stat(r.Path); os.IsNotExist(err) {
				exists = "no"
			}

			fmt.Printf("Name:     %s\n", r.Name)
			fmt.Printf("Type:     %s\n", r.Type)
			fmt.Printf("Path:     %s\n", r.Path)
			fmt.Printf("Exists:   %s\n", exists)

			if r.Adapter != "" {
				fmt.Printf("Adapter:  %s\n", r.Adapter)
			} else {
				fmt.Printf("Adapter:  (default)\n")
			}

			if len(r.Tags) > 0 {
				fmt.Printf("Tags:     %v\n", r.Tags)
			}

			if r.Type == role.Workspace {
				count := countChildren(r.Path)
				fmt.Printf("Projects: %d\n", count)
			}

			// Check for OpenCode config
			hasConfig := false
			if _, err := os.Stat(r.Path + "/opencode.json"); err == nil {
				hasConfig = true
			}
			fmt.Printf("Config:   %v\n", hasConfig)

			return nil
		},
	}

	return cmd
}
