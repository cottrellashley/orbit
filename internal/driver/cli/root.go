// Package cli implements the cobra-based CLI driver for Orbit.
package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	ccadapter "github.com/cottrellashley/orbit/internal/adapter/claudecode"
	copilotadapter "github.com/cottrellashley/orbit/internal/adapter/copilot"
	gh "github.com/cottrellashley/orbit/internal/adapter/github"
	"github.com/cottrellashley/orbit/internal/adapter/jsonstore"
	mdadapter "github.com/cottrellashley/orbit/internal/adapter/markdown"
	oc "github.com/cottrellashley/orbit/internal/adapter/opencode"
	termadapter "github.com/cottrellashley/orbit/internal/adapter/terminal"
	uvadapter "github.com/cottrellashley/orbit/internal/adapter/uv"
	"github.com/cottrellashley/orbit/internal/adapter/workspace"
	"github.com/cottrellashley/orbit/internal/app"
	"github.com/cottrellashley/orbit/internal/driver/server"
	"github.com/cottrellashley/orbit/internal/port"
)

// defaultStopTimeout is the timeout for stopping a managed server on shutdown.
const defaultStopTimeout = 15 * time.Second

// services bundles all app services returned from the composition root.
// Adding new services here keeps the wire() signature stable.
type services struct {
	env      *app.EnvironmentService
	project  *app.ProjectService
	session  *app.SessionService
	doctor   *app.DoctorService
	open     *app.OpenService
	query    *app.StateQuerier
	github   *app.GitHubService
	nav      *app.NavigationService
	markdown *app.MarkdownService
	node     *app.NodeService
	terminal *app.TerminalService
	install  *app.InstallService
	copilot  *app.CopilotService
	ws       port.ConfigWorkspace
}

// wire constructs all dependencies and returns the app services.
// This is the single composition root for the CLI driver.
func wire() *services {
	// Config workspace — single source of truth for Orbit directory layout.
	ws := workspace.New(workspace.DefaultRoot())
	cfgDir := ws.Root()
	storeDir := filepath.Join(cfgDir, "store")

	// Adapters
	envRepo := jsonstore.New(storeDir)
	projectRepo := jsonstore.NewProjectStore(storeDir)
	provider := oc.NewAdapter()
	ghClient := gh.NewClient(cfgDir)
	mdRenderer := mdadapter.NewFallbackRenderer()

	// App services (ProfileRepository is nil for now — profiles not yet implemented)
	envSvc := app.NewEnvironmentService(envRepo, nil)
	projectSvc := app.NewProjectService(projectRepo, nil)
	sessSvc := app.NewSessionService(envRepo, provider)
	docSvc := app.NewDoctorService(cfgDir, nil, envRepo, provider)
	openSvc := app.NewOpenService(envRepo, sessSvc)
	githubSvc := app.NewGitHubService(ghClient)
	navSvc := app.NewNavigationService()
	markdownSvc := app.NewMarkdownService(mdRenderer)

	// Wire project-aware dependencies into services that support them.
	sessSvc.SetProjects(projectRepo)
	docSvc.SetProjects(projectRepo)
	docSvc.SetWorkspace(ws)
	docSvc.SetToolLookup(exec.LookPath)
	openSvc.SetProjects(projectRepo, sessSvc)

	// Node registry — persistent store + app service.
	nodeStore := jsonstore.NewNodeStore(storeDir)
	nodeSvc := app.NewNodeService(nodeStore, provider)
	sessSvc.SetNodeStore(nodeStore)

	// Terminal manager — PTY-backed subprocess terminals for the web UI.
	termMgr := termadapter.New()
	termSvc := app.NewTerminalService(termMgr)

	// Install service — aggregates all ToolInstaller adapters.
	installSvc := app.NewInstallService(
		oc.NewInstaller(),
		gh.NewGHInstaller(),
		gh.NewCopilotInstaller(),
		uvadapter.NewInstaller(),
		ccadapter.NewInstaller(),
	)

	// Copilot coding agent service — shells out to gh agent-task.
	copilotProvider := copilotadapter.NewAdapter()
	copilotSvc := app.NewCopilotService(copilotProvider)

	// Assistant-ready read-only facade.
	querySvc := app.NewStateQuerier(projectSvc, sessSvc, docSvc, openSvc)

	return &services{
		env:      envSvc,
		project:  projectSvc,
		session:  sessSvc,
		doctor:   docSvc,
		open:     openSvc,
		query:    querySvc,
		github:   githubSvc,
		nav:      navSvc,
		markdown: markdownSvc,
		node:     nodeSvc,
		terminal: termSvc,
		install:  installSvc,
		copilot:  copilotSvc,
		ws:       ws,
	}
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
		Short:         "Orbit — manage AI coding projects and sessions",
		Long:          "Orbit is a CLI/UI tool for managing AI coding projects, environments, and sessions.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Wire services once — shared across all subcommands.
	svc := wire()

	// Create the lifecycle adapter for attach — reads state file only (no directory needed).
	attachMgr := newServerManager("")

	root.AddCommand(newServeCmd(svc))
	root.AddCommand(newAttachCmd(attachMgr))
	root.AddCommand(newProjectCmd(svc.project))
	root.AddCommand(newMCPCmd(svc))
	root.AddCommand(newInstallCmd(svc.install))
	root.AddCommand(newCopilotCmd(svc.copilot))

	return root
}

// newServeCmd creates the `orbit serve` command.
func newServeCmd(svc *services) *cobra.Command {
	var addr string
	var ocDir string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the Orbit web UI and managed OpenCode server",
		Long: `Starts a managed OpenCode server (opencode serve) and the Orbit HTTP
server that serves the web interface and API. The OpenCode server is
stopped automatically when orbit serve exits.`,
		RunE: func(cmd *cobra.Command, args []string) error {
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
			svc.session.SetLifecycle(mgr)

			srv := server.New(addr, svc.env, svc.project, svc.session, svc.doctor, svc.open, svc.github, svc.nav, svc.markdown)
			srv.SetNodeService(svc.node)
			srv.SetTerminalService(svc.terminal)
			srv.SetInstallService(svc.install)
			srv.SetCopilotService(svc.copilot)
			srv.SetManagedServerURL(fmt.Sprintf("http://%s:%d", ms.Hostname, ms.Port))

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
