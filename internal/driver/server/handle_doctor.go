package server

import "net/http"

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
