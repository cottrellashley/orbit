package server

import (
	"net/http"

	"github.com/cottrellashley/orbit/internal/domain"
)

// ---------------------------------------------------------------------------
// Install DTOs
// ---------------------------------------------------------------------------

type toolInfoJSON struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Version     string `json:"version,omitempty"`
}

func toToolInfoJSON(t domain.ToolInfo) toolInfoJSON {
	return toolInfoJSON{
		Name:        t.Name,
		Description: t.Description,
		Status:      t.Status.String(),
		Version:     t.Version,
	}
}

type installResultJSON struct {
	Name    string `json:"name"`
	Success bool   `json:"success"`
	Version string `json:"version,omitempty"`
	Error   string `json:"error,omitempty"`
}

func toInstallResultJSON(r domain.InstallResult) installResultJSON {
	return installResultJSON{
		Name:    r.Name,
		Success: r.Success,
		Version: r.Version,
		Error:   r.Error,
	}
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

// handleListInstalls returns all registered tools and their install status.
func (s *Server) handleListInstalls(w http.ResponseWriter, r *http.Request) {
	if s.installSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "install service not available")
		return
	}

	tools, err := s.installSvc.ListAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	out := make([]toolInfoJSON, len(tools))
	for i, t := range tools {
		out[i] = toToolInfoJSON(t)
	}
	writeJSON(w, http.StatusOK, out)
}

// handleInstallTool triggers installation of a tool by name.
func (s *Server) handleInstallTool(w http.ResponseWriter, r *http.Request) {
	if s.installSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "install service not available")
		return
	}

	name := r.PathValue("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "tool name is required")
		return
	}

	result, err := s.installSvc.Install(r.Context(), name)
	if err != nil {
		writeJSON(w, http.StatusNotFound, toInstallResultJSON(result))
		return
	}

	status := http.StatusOK
	if !result.Success {
		status = http.StatusUnprocessableEntity
	}
	writeJSON(w, status, toInstallResultJSON(result))
}
