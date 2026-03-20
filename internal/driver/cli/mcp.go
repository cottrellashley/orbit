package cli

import (
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"

	mcpdriver "github.com/cottrellashley/orbit/internal/driver/mcpserver"
)

// newMCPCmd creates the `orbit mcp` command. It starts the MCP server on
// stdio transport so the chatbot's OpenCode instance (or any MCP client)
// can invoke Orbit tools as a subprocess.
func newMCPCmd(svc *services) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start the Orbit MCP server (stdio transport)",
		Long: `Starts the Orbit MCP server on stdio. This is designed to be invoked
as a subprocess by an MCP client (e.g. the chatbot's OpenCode server).

All Orbit tools are available — projects, sessions, nodes, diagnostics,
GitHub integration, and more.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			srv := mcpdriver.New(
				svc.project,
				svc.session,
				svc.node,
				svc.doctor,
				svc.github,
				svc.nav,
				svc.markdown,
				svc.env,
			)

			// Suppress any non-protocol output on stderr to keep stdio clean.
			fmt.Fprintln(os.Stderr, "Orbit MCP server starting on stdio...")

			return srv.Run(cmd.Context(), &mcp.StdioTransport{})
		},
	}

	return cmd
}
