package cli

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/cottrellashley/orbit/internal/app"
	"github.com/cottrellashley/orbit/internal/domain"
)

// newCopilotCmd creates the `orbit copilot` command group.
func newCopilotCmd(svc *app.CopilotService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copilot",
		Short: "Manage GitHub Copilot coding agent tasks",
		Long:  "List, view, create, and stop Copilot coding agent tasks (cloud-based PR agents).",
	}

	cmd.AddCommand(newCopilotListCmd(svc))
	cmd.AddCommand(newCopilotViewCmd(svc))
	cmd.AddCommand(newCopilotCreateCmd(svc))
	cmd.AddCommand(newCopilotStopCmd(svc))
	cmd.AddCommand(newCopilotLogsCmd(svc))

	return cmd
}

// newCopilotListCmd creates `orbit copilot list`.
func newCopilotListCmd(svc *app.CopilotService) *cobra.Command {
	var owner, repo string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Copilot agent tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			tasks, err := svc.ListTasks(cmd.Context(), owner, repo)
			if err != nil {
				return fmt.Errorf("list tasks: %w", err)
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "ID\tSTATUS\tTITLE\tURL")
			for _, t := range tasks {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", t.ID, t.Status, t.Title, t.HTMLURL)
			}
			return tw.Flush()
		},
	}

	cmd.Flags().StringVar(&owner, "owner", "", "filter by owner (org or user)")
	cmd.Flags().StringVar(&repo, "repo", "", "filter by repository name")

	return cmd
}

// newCopilotViewCmd creates `orbit copilot view <session-id>`.
func newCopilotViewCmd(svc *app.CopilotService) *cobra.Command {
	return &cobra.Command{
		Use:   "view <session-id>",
		Short: "View details of a Copilot agent task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

			task, err := svc.GetTask(cmd.Context(), sessionID)
			if err != nil {
				return fmt.Errorf("get task: %w", err)
			}

			fmt.Fprintf(os.Stdout, "ID:       %s\n", task.ID)
			fmt.Fprintf(os.Stdout, "Status:   %s\n", task.Status)
			fmt.Fprintf(os.Stdout, "Title:    %s\n", task.Title)
			fmt.Fprintf(os.Stdout, "Repo:     %s/%s\n", task.Owner, task.Repo)
			fmt.Fprintf(os.Stdout, "PR:       #%d\n", task.PRNumber)
			fmt.Fprintf(os.Stdout, "URL:      %s\n", task.HTMLURL)
			fmt.Fprintf(os.Stdout, "Created:  %s\n", task.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Fprintf(os.Stdout, "Updated:  %s\n", task.UpdatedAt.Format("2006-01-02 15:04:05"))
			if task.Prompt != "" {
				fmt.Fprintf(os.Stdout, "\nPrompt:\n%s\n", task.Prompt)
			}
			return nil
		},
	}
}

// newCopilotCreateCmd creates `orbit copilot create`.
func newCopilotCreateCmd(svc *app.CopilotService) *cobra.Command {
	var baseBranch, instructions string

	cmd := &cobra.Command{
		Use:   "create <owner> <repo> <prompt>",
		Short: "Start a new Copilot coding agent task",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			task, err := svc.CreateTask(cmd.Context(), domain.CopilotTaskCreateOpts{
				Owner:              args[0],
				Repo:               args[1],
				Prompt:             args[2],
				BaseBranch:         baseBranch,
				CustomInstructions: instructions,
			})
			if err != nil {
				return fmt.Errorf("create task: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Task created: %s\n", task.ID)
			if task.HTMLURL != "" {
				fmt.Fprintf(os.Stdout, "PR: %s\n", task.HTMLURL)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&baseBranch, "base", "", "base branch (default: repo default)")
	cmd.Flags().StringVar(&instructions, "instructions", "", "custom instructions for the agent")

	return cmd
}

// newCopilotStopCmd creates `orbit copilot stop <session-id>`.
func newCopilotStopCmd(svc *app.CopilotService) *cobra.Command {
	return &cobra.Command{
		Use:   "stop <session-id>",
		Short: "Stop a running Copilot agent task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

			if err := svc.StopTask(cmd.Context(), sessionID); err != nil {
				return fmt.Errorf("stop task: %w", err)
			}
			fmt.Fprintf(os.Stdout, "Task %s stopped\n", sessionID)
			return nil
		},
	}
}

// newCopilotLogsCmd creates `orbit copilot logs <session-id>`.
func newCopilotLogsCmd(svc *app.CopilotService) *cobra.Command {
	return &cobra.Command{
		Use:   "logs <session-id>",
		Short: "Stream logs for a Copilot agent task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

			rc, err := svc.TaskLogs(cmd.Context(), sessionID)
			if err != nil {
				return fmt.Errorf("task logs: %w", err)
			}
			defer rc.Close()

			_, err = io.Copy(os.Stdout, rc)
			return err
		},
	}
}
