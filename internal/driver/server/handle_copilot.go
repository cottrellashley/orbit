package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/cottrellashley/orbit/internal/domain"
)

// ---------------------------------------------------------------------------
// Copilot Task DTOs
// ---------------------------------------------------------------------------

type copilotTaskJSON struct {
	ID        string `json:"id"`
	Owner     string `json:"owner"`
	Repo      string `json:"repo"`
	PRNumber  int    `json:"prNumber"`
	Title     string `json:"title"`
	Prompt    string `json:"prompt,omitempty"`
	Status    string `json:"status"`
	HTMLURL   string `json:"htmlUrl"`
	Branch    string `json:"branch,omitempty"`
	Draft     bool   `json:"draft"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func toCopilotTaskJSON(t domain.CopilotTask) copilotTaskJSON {
	return copilotTaskJSON{
		ID:        t.ID,
		Owner:     t.Owner,
		Repo:      t.Repo,
		PRNumber:  t.PRNumber,
		Title:     t.Title,
		Prompt:    t.Prompt,
		Status:    t.Status.String(),
		HTMLURL:   t.HTMLURL,
		Branch:    t.Branch,
		Draft:     t.Draft,
		CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

type copilotAvailableJSON struct {
	Available bool `json:"available"`
}

type createCopilotTaskRequest struct {
	Owner              string `json:"owner"`
	Repo               string `json:"repo"`
	Prompt             string `json:"prompt"`
	BaseBranch         string `json:"baseBranch,omitempty"`
	CustomInstructions string `json:"customInstructions,omitempty"`
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

// handleCopilotAvailable reports whether the Copilot agent-task CLI is usable.
func (s *Server) handleCopilotAvailable(w http.ResponseWriter, r *http.Request) {
	if s.copilotSvc == nil {
		writeJSON(w, http.StatusOK, copilotAvailableJSON{Available: false})
		return
	}
	writeJSON(w, http.StatusOK, copilotAvailableJSON{Available: s.copilotSvc.IsAvailable()})
}

// handleListCopilotTasks returns Copilot agent tasks, optionally filtered
// by ?owner= and ?repo= query parameters.
func (s *Server) handleListCopilotTasks(w http.ResponseWriter, r *http.Request) {
	if s.copilotSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "copilot service not available")
		return
	}

	owner := r.URL.Query().Get("owner")
	repo := r.URL.Query().Get("repo")

	tasks, err := s.copilotSvc.ListTasks(r.Context(), owner, repo)
	if err != nil {
		if errors.Is(err, domain.ErrCopilotUnavailable) {
			writeError(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	out := make([]copilotTaskJSON, len(tasks))
	for i, t := range tasks {
		out[i] = toCopilotTaskJSON(t)
	}
	writeJSON(w, http.StatusOK, out)
}

// handleGetCopilotTask returns a single Copilot agent task by session ID.
func (s *Server) handleGetCopilotTask(w http.ResponseWriter, r *http.Request) {
	if s.copilotSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "copilot service not available")
		return
	}

	sessionID := r.PathValue("id")
	if sessionID == "" {
		writeError(w, http.StatusBadRequest, "session ID is required")
		return
	}

	task, err := s.copilotSvc.GetTask(r.Context(), sessionID)
	if err != nil {
		if errors.Is(err, domain.ErrCopilotUnavailable) {
			writeError(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toCopilotTaskJSON(*task))
}

// handleCreateCopilotTask starts a new Copilot coding agent task.
func (s *Server) handleCreateCopilotTask(w http.ResponseWriter, r *http.Request) {
	if s.copilotSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "copilot service not available")
		return
	}

	var req createCopilotTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	task, err := s.copilotSvc.CreateTask(r.Context(), domain.CopilotTaskCreateOpts{
		Owner:              req.Owner,
		Repo:               req.Repo,
		Prompt:             req.Prompt,
		BaseBranch:         req.BaseBranch,
		CustomInstructions: req.CustomInstructions,
	})
	if err != nil {
		if errors.Is(err, domain.ErrCopilotUnavailable) {
			writeError(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, toCopilotTaskJSON(*task))
}

// handleStopCopilotTask stops a running Copilot agent task by session ID.
func (s *Server) handleStopCopilotTask(w http.ResponseWriter, r *http.Request) {
	if s.copilotSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "copilot service not available")
		return
	}

	sessionID := r.PathValue("id")
	if sessionID == "" {
		writeError(w, http.StatusBadRequest, "session ID is required")
		return
	}

	if err := s.copilotSvc.StopTask(r.Context(), sessionID); err != nil {
		if errors.Is(err, domain.ErrCopilotUnavailable) {
			writeError(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleCopilotTaskLogs streams the session log for a Copilot agent task.
func (s *Server) handleCopilotTaskLogs(w http.ResponseWriter, r *http.Request) {
	if s.copilotSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "copilot service not available")
		return
	}

	sessionID := r.PathValue("id")
	if sessionID == "" {
		writeError(w, http.StatusBadRequest, "session ID is required")
		return
	}

	rc, err := s.copilotSvc.TaskLogs(r.Context(), sessionID)
	if err != nil {
		if errors.Is(err, domain.ErrCopilotUnavailable) {
			writeError(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, rc)
}
