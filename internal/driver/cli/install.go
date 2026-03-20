package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/cottrellashley/orbit/internal/app"
)

// newInstallCmd creates the `orbit install` command group.
func newInstallCmd(svc *app.InstallService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Manage tool installations",
		Long:  "List, check, and install development tools managed by Orbit.",
	}

	cmd.AddCommand(newInstallListCmd(svc))
	cmd.AddCommand(newInstallRunCmd(svc))

	return cmd
}

// newInstallListCmd creates the `orbit install list` command.
func newInstallListCmd(svc *app.InstallService) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all tools and their install status",
		RunE: func(cmd *cobra.Command, args []string) error {
			tools, err := svc.ListAll(cmd.Context())
			if err != nil {
				return fmt.Errorf("list tools: %w", err)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "NAME\tSTATUS\tVERSION\tDESCRIPTION")
			for _, t := range tools {
				ver := t.Version
				if ver == "" {
					ver = "-"
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", t.Name, t.Status, ver, t.Description)
			}
			return tw.Flush()
		},
	}
}

// newInstallRunCmd creates the `orbit install <name>` command.
func newInstallRunCmd(svc *app.InstallService) *cobra.Command {
	return &cobra.Command{
		Use:   "run <name>",
		Short: "Install a specific tool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			fmt.Fprintf(os.Stderr, "Installing %s...\n", name)

			result, err := svc.Install(cmd.Context(), name)
			if err != nil {
				return fmt.Errorf("install %s: %w", name, err)
			}

			if result.Success {
				fmt.Fprintf(os.Stdout, "Successfully installed %s", name)
				if result.Version != "" {
					fmt.Fprintf(os.Stdout, " (%s)", result.Version)
				}
				fmt.Fprintln(os.Stdout)
			} else {
				fmt.Fprintf(os.Stderr, "Failed to install %s: %s\n", name, result.Error)
				os.Exit(1)
			}
			return nil
		},
	}
}
