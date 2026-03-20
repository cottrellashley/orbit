package server

import "net/http"

func (s *Server) handleGitHubStatus(w http.ResponseWriter, r *http.Request) {
	status, err := s.ghSvc.AuthStatus(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, githubAuthStatusJSON{
		Authenticated: status.Authenticated,
		User:          status.User,
		TokenSource:   status.TokenSource,
		Scopes:        status.Scopes,
	})
}

func (s *Server) handleGitHubRepos(w http.ResponseWriter, r *http.Request) {
	repos, err := s.ghSvc.ListRepos(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]githubRepoJSON, len(repos))
	for i, repo := range repos {
		out[i] = toGitHubRepoJSON(repo)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleGitHubIssues(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("owner")
	repo := r.PathValue("repo")
	issues, err := s.ghSvc.ListIssues(r.Context(), owner, repo)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]githubIssueJSON, len(issues))
	for i, issue := range issues {
		out[i] = toGitHubIssueJSON(issue)
	}
	writeJSON(w, http.StatusOK, out)
}
