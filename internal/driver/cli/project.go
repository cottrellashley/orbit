package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/cottrellashley/orbit/internal/app"
	"github.com/spf13/cobra"
)

// newProjectCmd creates the `orbit project` command group.
// Environment commands remain available alongside these; project commands
// are the preferred path forward per ADR-001.
func newProjectCmd(svc *app.ProjectService) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "project",
		Aliases: []string{"proj"},
		Short:   "Manage Orbit projects",
		Long: `Register, list, and remove projects. Projects are the successor to
environments — they support richer metadata including repository topology,
integration tags, and contained-repo info.`,
	}

	cmd.AddCommand(newProjectListCmd(svc))
	cmd.AddCommand(newProjectAddCmd(svc))
	cmd.AddCommand(newProjectRemoveCmd(svc))
	cmd.AddCommand(newProjectShowCmd(svc))

	return cmd
}

func newProjectListCmd(svc *app.ProjectService) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List registered projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			projects, err := svc.List()
			if err != nil {
				return err
			}
			if len(projects) == 0 {
				fmt.Fprintln(os.Stderr, "No projects registered. Use 'orbit project add' to register one.")
				return nil
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "NAME\tPATH\tTOPOLOGY\tDESCRIPTION")
			for _, p := range projects {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", p.Name, p.Path, p.Topology, p.Description)
			}
			return tw.Flush()
		},
	}
}

func newProjectAddCmd(svc *app.ProjectService) *cobra.Command {
	var description string

	cmd := &cobra.Command{
		Use:   "add <name> <path>",
		Short: "Register a new project",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := svc.Register(args[0], args[1], description)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Project %q registered at %s\n", p.Name, p.Path)
			return nil
		},
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "project description")

	return cmd
}

func newProjectRemoveCmd(svc *app.ProjectService) *cobra.Command {
	return &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm"},
		Short:   "Remove a registered project",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := svc.Delete(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Project %q removed.\n", args[0])
			return nil
		},
	}
}

func newProjectShowCmd(svc *app.ProjectService) *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Show details for a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := svc.Get(args[0])
			if err != nil {
				return err
			}

			fmt.Fprintf(os.Stdout, "Name:        %s\n", p.Name)
			fmt.Fprintf(os.Stdout, "Path:        %s\n", p.Path)
			fmt.Fprintf(os.Stdout, "Description: %s\n", p.Description)
			fmt.Fprintf(os.Stdout, "Topology:    %s\n", p.Topology)
			fmt.Fprintf(os.Stdout, "Created:     %s\n", p.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Fprintf(os.Stdout, "Updated:     %s\n", p.UpdatedAt.Format("2006-01-02 15:04:05"))

			if p.ProfileName != "" {
				fmt.Fprintf(os.Stdout, "Profile:     %s\n", p.ProfileName)
			}

			if len(p.Integrations) > 0 {
				fmt.Fprintf(os.Stdout, "Integrations:")
				for _, tag := range p.Integrations {
					fmt.Fprintf(os.Stdout, " %s", tag)
				}
				fmt.Fprintln(os.Stdout)
			}

			if len(p.Repos) > 0 {
				fmt.Fprintln(os.Stdout, "Repos:")
				for _, r := range p.Repos {
					fmt.Fprintf(os.Stdout, "  %s", r.Path)
					if r.CurrentBranch != "" {
						fmt.Fprintf(os.Stdout, " (%s)", r.CurrentBranch)
					}
					if r.RemoteURL != "" {
						fmt.Fprintf(os.Stdout, " -> %s", r.RemoteURL)
					}
					fmt.Fprintln(os.Stdout)
				}
			}

			return nil
		},
	}
}
