// Package server implements the HTTP driver for Orbit. It exposes the
// app services as a REST API and serves the web UI as a single embedded
// HTML page. It also provides a reverse proxy to OpenCode servers so the
// frontend can fetch conversations without CORS issues.
package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
)

//go:embed static/index.html
var staticFS embed.FS

// ---------------------------------------------------------------------------
// Service interfaces — consumer-defined, decoupled from concrete app types
// ---------------------------------------------------------------------------

// EnvironmentService is the subset of app.EnvironmentService used by the HTTP driver.
type EnvironmentService interface {
	List() ([]*domain.Environment, error)
	Register(name, path, description string) (*domain.Environment, error)
	Delete(name string) error
}

// SessionService is the subset of app.SessionService used by the HTTP driver.
type SessionService interface {
	DiscoverServers(ctx context.Context) ([]domain.Server, error)
	ListAll(ctx context.Context) ([]domain.Session, error)
	GetSession(ctx context.Context, sessionID string) (*domain.Session, error)
	AbortSession(ctx context.Context, sessionID string) error
	DeleteSession(ctx context.Context, sessionID string) error
}

// DoctorService is the subset of app.DoctorService used by the HTTP driver.
type DoctorService interface {
	Run(ctx context.Context) *domain.Report
}

// OpenService is the subset of app.OpenService used by the HTTP driver.
type OpenService interface {
	Resolve(ctx context.Context, envName string) (*domain.OpenPlan, error)
}

// Server is the HTTP driver for Orbit.
type Server struct {
	addr    string
	envSvc  EnvironmentService
	sessSvc SessionService
	docSvc  DoctorService
	openSvc OpenService
	mux     *http.ServeMux
	httpSrv *http.Server
}

// New creates a Server.
func New(
	addr string,
	envSvc EnvironmentService,
	sessSvc SessionService,
	docSvc DoctorService,
	openSvc OpenService,
) *Server {
	s := &Server{
		addr:    addr,
		envSvc:  envSvc,
		sessSvc: sessSvc,
		docSvc:  docSvc,
		openSvc: openSvc,
		mux:     http.NewServeMux(),
	}
	s.routes()
	return s
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
	// Orbit API
	s.mux.HandleFunc("GET /api/servers", s.handleListServers)
	s.mux.HandleFunc("GET /api/sessions", s.handleListSessions)
	s.mux.HandleFunc("GET /api/sessions/{id}", s.handleGetSession)
	s.mux.HandleFunc("POST /api/sessions/{id}/abort", s.handleAbortSession)
	s.mux.HandleFunc("DELETE /api/sessions/{id}", s.handleDeleteSession)
	s.mux.HandleFunc("GET /api/environments", s.handleListEnvironments)
	s.mux.HandleFunc("POST /api/environments", s.handleRegisterEnvironment)
	s.mux.HandleFunc("DELETE /api/environments/{name}", s.handleDeleteEnvironment)
	s.mux.HandleFunc("GET /api/doctor", s.handleDoctor)

	// Proxy to OpenCode servers — supports all methods.
	// Pattern: /api/proxy/{port}/{path...}
	s.mux.HandleFunc("GET /api/proxy/{port}/{path...}", s.handleProxy)
	s.mux.HandleFunc("POST /api/proxy/{port}/{path...}", s.handleProxy)
	s.mux.HandleFunc("PUT /api/proxy/{port}/{path...}", s.handleProxy)
	s.mux.HandleFunc("PATCH /api/proxy/{port}/{path...}", s.handleProxy)
	s.mux.HandleFunc("DELETE /api/proxy/{port}/{path...}", s.handleProxy)

	// Static UI — must be last (catch-all).
	s.mux.HandleFunc("GET /{$}", s.handleIndex)
}

// ---------------------------------------------------------------------------
// Static
// ---------------------------------------------------------------------------

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	data, err := staticFS.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

// ---------------------------------------------------------------------------
// API response DTOs — keeps JSON serialization in the driver layer
// ---------------------------------------------------------------------------

type serverJSON struct {
	PID       int    `json:"pid"`
	Port      int    `json:"port"`
	Hostname  string `json:"hostname"`
	Directory string `json:"directory"`
	Version   string `json:"version,omitempty"`
	Healthy   bool   `json:"healthy"`
}

func toServerJSON(s domain.Server) serverJSON {
	return serverJSON{
		PID:       s.PID,
		Port:      s.Port,
		Hostname:  s.Hostname,
		Directory: s.Directory,
		Version:   s.Version,
		Healthy:   s.Healthy,
	}
}

type sessionJSON struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	EnvironmentName string `json:"environment_name"`
	EnvironmentPath string `json:"environment_path"`
	ServerDir       string `json:"server_dir"`
	ServerPort      int    `json:"server_port"`
	Status          string `json:"status"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

func toSessionJSON(s domain.Session) sessionJSON {
	return sessionJSON{
		ID:              s.ID,
		Title:           s.Title,
		EnvironmentName: s.EnvironmentName,
		EnvironmentPath: s.EnvironmentPath,
		ServerDir:       s.ServerDir,
		ServerPort:      s.ServerPort,
		Status:          s.Status,
		CreatedAt:       s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       s.UpdatedAt.Format(time.RFC3339),
	}
}

type environmentJSON struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	ProfileName string `json:"profile_name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func toEnvironmentJSON(e *domain.Environment) environmentJSON {
	return environmentJSON{
		Name:        e.Name,
		Path:        e.Path,
		ProfileName: e.ProfileName,
		Description: e.Description,
		CreatedAt:   e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   e.UpdatedAt.Format(time.RFC3339),
	}
}

// ---------------------------------------------------------------------------
// Orbit API handlers
// ---------------------------------------------------------------------------

func (s *Server) handleListServers(w http.ResponseWriter, r *http.Request) {
	servers, err := s.sessSvc.DiscoverServers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]serverJSON, len(servers))
	for i, srv := range servers {
		out[i] = toServerJSON(srv)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleListSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := s.sessSvc.ListAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]sessionJSON, len(sessions))
	for i, sess := range sessions {
		out[i] = toSessionJSON(sess)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleGetSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	session, err := s.sessSvc.GetSession(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toSessionJSON(*session))
}

func (s *Server) handleAbortSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.sessSvc.AbortSession(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.sessSvc.DeleteSession(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleListEnvironments(w http.ResponseWriter, r *http.Request) {
	envs, err := s.envSvc.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]environmentJSON, len(envs))
	for i, env := range envs {
		out[i] = toEnvironmentJSON(env)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleRegisterEnvironment(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		Path        string `json:"path"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if body.Name == "" || body.Path == "" {
		writeError(w, http.StatusBadRequest, "name and path are required")
		return
	}
	env, err := s.envSvc.Register(body.Name, body.Path, body.Description)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, toEnvironmentJSON(env))
}

func (s *Server) handleDeleteEnvironment(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := s.envSvc.Delete(name); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleDoctor(w http.ResponseWriter, r *http.Request) {
	report := s.docSvc.Run(r.Context())

	type checkJSON struct {
		Name    string `json:"name"`
		Status  string `json:"status"`
		Message string `json:"message"`
		Fix     string `json:"fix,omitempty"`
	}

	results := make([]checkJSON, len(report.Results))
	for i, c := range report.Results {
		results[i] = checkJSON{
			Name:    c.Name,
			Status:  c.Status.String(),
			Message: c.Message,
			Fix:     c.Fix,
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      report.OK(),
		"results": results,
	})
}

// ---------------------------------------------------------------------------
// Proxy to OpenCode servers
// ---------------------------------------------------------------------------

func (s *Server) handleProxy(w http.ResponseWriter, r *http.Request) {
	portStr := r.PathValue("port")
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		writeError(w, http.StatusBadRequest, "invalid port")
		return
	}

	targetPath := "/" + r.PathValue("path")
	if r.URL.RawQuery != "" {
		targetPath += "?" + r.URL.RawQuery
	}

	target, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", port))
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = targetPath
			req.URL.RawQuery = r.URL.RawQuery
			req.Host = target.Host
			// Forward basic auth if the OpenCode server uses it
			if user, pass, ok := r.BasicAuth(); ok {
				req.SetBasicAuth(user, pass)
			}
		},
		ModifyResponse: func(resp *http.Response) error {
			// If this is an SSE stream, ensure proper headers
			if strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream") {
				resp.Header.Set("Cache-Control", "no-cache")
				resp.Header.Set("Connection", "keep-alive")
			}
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			writeError(w, http.StatusBadGateway, fmt.Sprintf("proxy error: %v", err))
		},
	}

	proxy.ServeHTTP(w, r)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// CopyBody is a helper to read and close a request body.
func CopyBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}
