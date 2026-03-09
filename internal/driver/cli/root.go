// Package cli implements the cobra-based CLI driver for Orbit.
package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/cottrellashley/orbit/internal/adapter/jsonstore"
	oc "github.com/cottrellashley/orbit/internal/adapter/opencode"
	"github.com/cottrellashley/orbit/internal/app"
	"github.com/cottrellashley/orbit/internal/driver/server"
	"github.com/cottrellashley/orbit/internal/port"
)

// defaultStopTimeout is the timeout for stopping a managed server on shutdown.
const defaultStopTimeout = 15 * time.Second

// configDir returns the orbit config directory.
func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "orbit")
}

// wire constructs all dependencies and returns the app services.
// This is the single composition root for the CLI driver.
func wire() (envSvc *app.EnvironmentService, sessSvc *app.SessionService, docSvc *app.DoctorService, openSvc *app.OpenService) {
	cfgDir := configDir()

	// Adapters
	envRepo := jsonstore.New(filepath.Join(cfgDir, "store"))
	provider := oc.NewAdapter()

	// App services (ProfileRepository is nil for now — profiles not yet implemented)
	envSvc = app.NewEnvironmentService(envRepo, nil)
	sessSvc = app.NewSessionService(envRepo, provider)
	docSvc = app.NewDoctorService(cfgDir, nil, envRepo, provider)
	openSvc = app.NewOpenService(envRepo, sessSvc)
	return
}

// newServerManager creates a ServerLifecycle for the given working directory.
// This is the only place adapter types are referenced for lifecycle construction.
func newServerManager(dir string) port.ServerLifecycle {
	return oc.NewServerManager(oc.ServerManagerOpts{
		ProcessOpts: oc.ProcessOpts{
			Directory: dir,
		},
	})
}

// NewRootCmd creates the top-level orbit command.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "orbit",
		Short:         "Orbit — manage AI coding environments",
		Long:          "Orbit is a CLI/UI tool for managing AI coding environments and sessions.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Create the lifecycle adapter for attach — reads state file only (no directory needed).
	attachMgr := newServerManager("")

	root.AddCommand(newServeCmd())
	root.AddCommand(newAttachCmd(attachMgr))

	return root
}

// newServeCmd creates the `orbit serve` command.
func newServeCmd() *cobra.Command {
	var addr string
	var ocDir string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the Orbit web UI and managed OpenCode server",
		Long: `Starts a managed OpenCode server (opencode serve) and the Orbit HTTP
server that serves the web interface and API. The OpenCode server is
stopped automatically when orbit serve exits.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			envSvc, sessSvc, docSvc, openSvc := wire()

			// Resolve the directory for the managed OpenCode server.
			dir := ocDir
			if dir == "" {
				var err error
				dir, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("get working directory: %w", err)
				}
			}

			// Create the ServerManager (implements port.ServerLifecycle).
			mgr := newServerManager(dir)

			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			// Start the managed OpenCode server.
			ms, err := mgr.Start(ctx)
			if err != nil {
				return fmt.Errorf("start opencode server: %w", err)
			}
			fmt.Fprintf(os.Stderr, "Managed OpenCode server running on http://%s:%d (PID %d)\n",
				ms.Hostname, ms.Port, ms.PID)

			// Tell the SessionService about the managed server.
			sessSvc.SetLifecycle(mgr)

			srv := server.New(addr, envSvc, sessSvc, docSvc, openSvc)

			fmt.Fprintf(os.Stderr, "Starting Orbit UI at http://%s\n", addr)

			// Run the HTTP server (blocks until context is cancelled).
			srvErr := srv.ListenAndServe(ctx)

			// Shutdown: stop the managed OpenCode server.
			fmt.Fprintln(os.Stderr, "Stopping managed OpenCode server...")
			stopCtx, stopCancel := context.WithTimeout(context.Background(), defaultStopTimeout)
			defer stopCancel()
			if err := mgr.Stop(stopCtx); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to stop opencode server: %v\n", err)
			}

			return srvErr
		},
	}

	cmd.Flags().StringVar(&addr, "addr", "127.0.0.1:3000", "address to listen on")
	cmd.Flags().StringVar(&ocDir, "dir", "", "working directory for the OpenCode server (default: current directory)")

	return cmd
}
