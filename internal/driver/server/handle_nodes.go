package server

import (
	"encoding/json"
	"net/http"

	"github.com/cottrellashley/orbit/internal/domain"
)

func (s *Server) handleListNodes(w http.ResponseWriter, r *http.Request) {
	if s.nodeSvc == nil {
		writeError(w, http.StatusNotImplemented, "node service not configured")
		return
	}
	nodes, err := s.nodeSvc.ListNodes(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]nodeJSON, len(nodes))
	for i, n := range nodes {
		out[i] = toNodeJSON(n)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleGetNode(w http.ResponseWriter, r *http.Request) {
	if s.nodeSvc == nil {
		writeError(w, http.StatusNotImplemented, "node service not configured")
		return
	}
	id := r.PathValue("id")
	node, err := s.nodeSvc.GetNode(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toNodeJSON(node))
}

// registerNodeRequest is the JSON body for POST /api/nodes.
type registerNodeRequest struct {
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
	Provider string `json:"provider"`
	Name     string `json:"name"`
}

func (s *Server) handleRegisterNode(w http.ResponseWriter, r *http.Request) {
	if s.nodeSvc == nil {
		writeError(w, http.StatusNotImplemented, "node service not configured")
		return
	}
	var req registerNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	provider := domain.ParseNodeProvider(req.Provider)

	node, err := s.nodeSvc.RegisterNode(r.Context(), req.Hostname, req.Port, provider, req.Name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, toNodeJSON(node))
}

func (s *Server) handleRemoveNode(w http.ResponseWriter, r *http.Request) {
	if s.nodeSvc == nil {
		writeError(w, http.StatusNotImplemented, "node service not configured")
		return
	}
	id := r.PathValue("id")
	if err := s.nodeSvc.RemoveNode(r.Context(), id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleSyncNodes(w http.ResponseWriter, r *http.Request) {
	if s.nodeSvc == nil {
		writeError(w, http.StatusNotImplemented, "node service not configured")
		return
	}
	nodes, err := s.nodeSvc.SyncDiscoveredNodes(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]nodeJSON, len(nodes))
	for i, n := range nodes {
		out[i] = toNodeJSON(n)
	}
	writeJSON(w, http.StatusOK, out)
}
