package server

import "net/http"

func (s *Server) handleProjectLinks(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	proj, err := s.projSvc.Get(name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	links := s.navSvc.ProjectLinks(*proj)
	out := make([]jumpLinkJSON, len(links))
	for i, l := range links {
		out[i] = toJumpLinkJSON(l)
	}
	writeJSON(w, http.StatusOK, out)
}
