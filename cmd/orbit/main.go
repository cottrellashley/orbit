package main

import (
	"fmt"
	"os"

	"github.com/cottrellashley/orbit/cmd/orbit/commands"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "orbit",
		Short: "Manage AI environments and workspaces",
		Long:  "orbit is a role-based launcher for AI coding environments.\nIt manages directories, configuration scaffolding, and adapter-based launching.",
	}

	root.AddCommand(
		commands.InitCmd(),
		commands.OpenCmd(),
		commands.NewCmd(),
		commands.ListCmd(),
		commands.ArchiveCmd(),
		commands.StatusCmd(),
	)

	// Global flag for config path
	root.PersistentFlags().String("config", "", "path to config file (default ~/.config/orbit/config.yaml)")

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
