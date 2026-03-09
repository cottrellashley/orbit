package cli

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/cottrellashley/orbit/internal/port"
	"github.com/spf13/cobra"
)

// newAttachCmd creates the `orbit attach` command.
// It reads the state file to find the managed OpenCode server and
// execs `opencode attach <url>` so the user gets a TUI connected
// to the Orbit-managed server.
func newAttachCmd(lc port.ServerLifecycle) *cobra.Command {
	var (
		session  string
		cont     bool
		fork     bool
		password string
	)

	cmd := &cobra.Command{
		Use:   "attach",
		Short: "Attach an OpenCode TUI to the managed server",
		Long: `Connects an OpenCode TUI session to the server managed by orbit serve.
The managed server's URL is read from the state file. This is equivalent
to running 'opencode attach <url>' with the correct address.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Read the state file to find the managed server.
			ms, err := lc.Status(cmd.Context())
			if err != nil || ms == nil {
				return fmt.Errorf("no managed OpenCode server is running (start one with 'orbit serve')")
			}

			// Build the opencode attach command.
			url := fmt.Sprintf("http://%s:%d", ms.Hostname, ms.Port)

			binPath, err := exec.LookPath("opencode")
			if err != nil {
				return fmt.Errorf("opencode binary not found: %w", err)
			}

			attachArgs := []string{"opencode", "attach", url}

			if session != "" {
				attachArgs = append(attachArgs, "--session", session)
			}
			if cont {
				attachArgs = append(attachArgs, "--continue")
			}
			if fork {
				attachArgs = append(attachArgs, "--fork")
			}

			pw := password
			if pw == "" {
				pw = ms.Password
			}
			if pw != "" {
				attachArgs = append(attachArgs, "--password", pw)
			}

			fmt.Fprintf(os.Stderr, "Attaching to %s ...\n", url)

			// Replace this process with opencode attach (exec).
			return syscall.Exec(binPath, attachArgs, os.Environ())
		},
	}

	cmd.Flags().StringVarP(&session, "session", "s", "", "session ID to attach to")
	cmd.Flags().BoolVarP(&cont, "continue", "c", false, "continue the last session")
	cmd.Flags().BoolVar(&fork, "fork", false, "fork the session")
	cmd.Flags().StringVarP(&password, "password", "p", "", "server password (default: from state file)")

	return cmd
}
