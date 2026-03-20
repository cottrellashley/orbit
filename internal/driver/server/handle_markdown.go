package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleMarkdownRender(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Source string `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if body.Source == "" {
		writeError(w, http.StatusBadRequest, "source is required")
		return
	}
	result, err := s.mdSvc.Render(r.Context(), body.Source)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"html":     result.HTML,
		"fallback": result.Fallback,
	})
}
