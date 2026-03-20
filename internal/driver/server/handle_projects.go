package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := s.projSvc.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]projectJSON, len(projects))
	for i, p := range projects {
		out[i] = toProjectJSON(p)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	p, err := s.projSvc.Get(name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toProjectJSON(p))
}

func (s *Server) handleRegisterProject(w http.ResponseWriter, r *http.Request) {
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
	p, err := s.projSvc.Register(body.Name, body.Path, body.Description)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, toProjectJSON(p))
}

func (s *Server) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := s.projSvc.Delete(name); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
