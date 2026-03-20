package server

import "net/http"

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
