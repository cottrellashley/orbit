package server

import (
	"encoding/json"
	"net/http"
)

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
