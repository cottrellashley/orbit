// Package server implements the HTTP driver for Orbit. It exposes the
// app services as a REST API and serves the web UI as a single embedded
// HTML page. It also provides a reverse proxy to OpenCode servers so the
// frontend can fetch conversations without CORS issues.
package server

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/cottrellashley/orbit/web"
)

// Server is the HTTP driver for Orbit.
type Server struct {
	addr             string
	envSvc           EnvironmentService
	projSvc          ProjectService
	sessSvc          SessionService
	docSvc           DoctorService
	openSvc          OpenService
	ghSvc            GitHubService
	navSvc           NavigationService
	mdSvc            MarkdownService
	nodeSvc          NodeService     // nil until node migration is wired
	termSvc          TerminalService // nil until terminal is wired
	installSvc       InstallService  // nil until install is wired
	copilotSvc       CopilotService  // nil until copilot is wired
	managedServerURL string          // URL of the managed OpenCode server
	mux              *http.ServeMux
	httpSrv          *http.Server
}

// New creates a Server.
func New(
	addr string,
	envSvc EnvironmentService,
	projSvc ProjectService,
	sessSvc SessionService,
	docSvc DoctorService,
	openSvc OpenService,
	ghSvc GitHubService,
	navSvc NavigationService,
	mdSvc MarkdownService,
) *Server {
	s := &Server{
		addr:    addr,
		envSvc:  envSvc,
		projSvc: projSvc,
		sessSvc: sessSvc,
		docSvc:  docSvc,
		openSvc: openSvc,
		ghSvc:   ghSvc,
		navSvc:  navSvc,
		mdSvc:   mdSvc,
		mux:     http.NewServeMux(),
	}
	s.routes()
	return s
}

// SetNodeService attaches a NodeService. When set, the node-related
// API endpoints and node-based proxy become available.
func (s *Server) SetNodeService(svc NodeService) {
	s.nodeSvc = svc
}

// SetTerminalService attaches a TerminalService. When set, the terminal
// API endpoints become available.
func (s *Server) SetTerminalService(svc TerminalService) {
	s.termSvc = svc
}

// SetInstallService attaches an InstallService. When set, the install
// API endpoints become available.
func (s *Server) SetInstallService(svc InstallService) {
	s.installSvc = svc
}

// SetCopilotService attaches a CopilotService. When set, the copilot
// agent-task API endpoints become available.
func (s *Server) SetCopilotService(svc CopilotService) {
	s.copilotSvc = svc
}

// SetManagedServerURL records the URL of the managed OpenCode server
// so the frontend can discover it via GET /api/managed-server.
func (s *Server) SetManagedServerURL(url string) {
	s.managedServerURL = url
}

// ListenAndServe starts the HTTP server. It blocks until the context
// is cancelled or the server errors out.
func (s *Server) ListenAndServe(ctx context.Context) error {
	s.httpSrv = &http.Server{
		Addr:    s.addr,
		Handler: s.mux,
	}

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.addr, err)
	}

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.httpSrv.Shutdown(shutCtx)
	}()

	log.Printf("orbit UI listening on http://%s", ln.Addr())
	if err := s.httpSrv.Serve(ln); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// ---------------------------------------------------------------------------
// Routes
// ---------------------------------------------------------------------------

func (s *Server) routes() {
	// Orbit API — shared
	s.mux.HandleFunc("GET /api/servers", s.handleListServers)
	s.mux.HandleFunc("GET /api/sessions", s.handleListSessions)
	s.mux.HandleFunc("GET /api/sessions/{id}", s.handleGetSession)
	s.mux.HandleFunc("POST /api/sessions/{id}/abort", s.handleAbortSession)
	s.mux.HandleFunc("DELETE /api/sessions/{id}", s.handleDeleteSession)
	s.mux.HandleFunc("GET /api/doctor", s.handleDoctor)

	// Project-first endpoints (new)
	s.mux.HandleFunc("GET /api/projects", s.handleListProjects)
	s.mux.HandleFunc("GET /api/projects/{name}", s.handleGetProject)
	s.mux.HandleFunc("POST /api/projects", s.handleRegisterProject)
	s.mux.HandleFunc("DELETE /api/projects/{name}", s.handleDeleteProject)

	// Legacy environment endpoints (compatibility aliases — still fully functional)
	s.mux.HandleFunc("GET /api/environments", s.handleListEnvironments)
	s.mux.HandleFunc("POST /api/environments", s.handleRegisterEnvironment)
	s.mux.HandleFunc("DELETE /api/environments/{name}", s.handleDeleteEnvironment)

	// GitHub integration endpoints
	s.mux.HandleFunc("GET /api/github/status", s.handleGitHubStatus)
	s.mux.HandleFunc("GET /api/github/repos", s.handleGitHubRepos)
	s.mux.HandleFunc("GET /api/github/repos/{owner}/{repo}/issues", s.handleGitHubIssues)

	// Navigation endpoints
	s.mux.HandleFunc("GET /api/projects/{name}/links", s.handleProjectLinks)

	// Markdown rendering endpoint
	s.mux.HandleFunc("POST /api/markdown/render", s.handleMarkdownRender)

	// Node registry endpoints (nil-guarded — return 501 when nodeSvc is not wired)
	s.mux.HandleFunc("GET /api/nodes", s.handleListNodes)
	s.mux.HandleFunc("GET /api/nodes/{id}", s.handleGetNode)
	s.mux.HandleFunc("POST /api/nodes", s.handleRegisterNode)
	s.mux.HandleFunc("DELETE /api/nodes/{id}", s.handleRemoveNode)
	s.mux.HandleFunc("POST /api/nodes/sync", s.handleSyncNodes)

	// Node-based proxy — routes by nodeID (new route)
	s.mux.HandleFunc("GET /api/nodes/{id}/proxy/{path...}", s.handleNodeProxy)
	s.mux.HandleFunc("POST /api/nodes/{id}/proxy/{path...}", s.handleNodeProxy)
	s.mux.HandleFunc("PUT /api/nodes/{id}/proxy/{path...}", s.handleNodeProxy)
	s.mux.HandleFunc("PATCH /api/nodes/{id}/proxy/{path...}", s.handleNodeProxy)
	s.mux.HandleFunc("DELETE /api/nodes/{id}/proxy/{path...}", s.handleNodeProxy)

	// Legacy proxy to OpenCode servers — supports all methods.
	// Pattern: /api/proxy/{port}/{path...}
	s.mux.HandleFunc("GET /api/proxy/{port}/{path...}", s.handleProxy)
	s.mux.HandleFunc("POST /api/proxy/{port}/{path...}", s.handleProxy)
	s.mux.HandleFunc("PUT /api/proxy/{port}/{path...}", s.handleProxy)
	s.mux.HandleFunc("PATCH /api/proxy/{port}/{path...}", s.handleProxy)
	s.mux.HandleFunc("DELETE /api/proxy/{port}/{path...}", s.handleProxy)

	// Managed server info (so frontend can discover the OpenCode URL)
	s.mux.HandleFunc("GET /api/managed-server", s.handleManagedServer)

	// Terminal endpoints (nil-guarded — return 501 when termSvc is not wired)
	s.mux.HandleFunc("POST /api/terminal", s.handleSpawnTerminal)
	s.mux.HandleFunc("GET /api/terminal", s.handleListTerminals)
	s.mux.HandleFunc("DELETE /api/terminal/{id}", s.handleKillTerminal)
	s.mux.HandleFunc("GET /api/terminal/{id}/ws", s.handleTerminalWS)

	// Install endpoints (nil-guarded — return 503 when installSvc is not wired)
	s.mux.HandleFunc("GET /api/installs", s.handleListInstalls)
	s.mux.HandleFunc("POST /api/installs/{name}", s.handleInstallTool)

	// Copilot agent-task endpoints (nil-guarded — return 503 when copilotSvc is not wired)
	s.mux.HandleFunc("GET /api/copilot/available", s.handleCopilotAvailable)
	s.mux.HandleFunc("GET /api/copilot/tasks", s.handleListCopilotTasks)
	s.mux.HandleFunc("GET /api/copilot/tasks/{id}", s.handleGetCopilotTask)
	s.mux.HandleFunc("POST /api/copilot/tasks", s.handleCreateCopilotTask)
	s.mux.HandleFunc("DELETE /api/copilot/tasks/{id}", s.handleStopCopilotTask)
	s.mux.HandleFunc("GET /api/copilot/tasks/{id}/logs", s.handleCopilotTaskLogs)

	// Static UI — SPA fallback (must be last).
	distFS, _ := fs.Sub(web.DistFS, "dist")
	fileServer := http.FileServer(http.FS(distFS))
	s.mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		// API paths should never reach here; 404 as a safety net.
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		// Try to serve the exact file (JS, CSS, images, etc.).
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if f, err := distFS.Open(path); err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}
		// File not found — SPA fallback: serve index.html for client-side routing.
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
