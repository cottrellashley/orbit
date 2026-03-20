package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer(envSvc EnvironmentService, projSvc ProjectService, sessSvc SessionService) *Server {
	return New("127.0.0.1:0", envSvc, projSvc, sessSvc, &mockDoctorService{}, &mockOpenService{}, &mockGitHubService{}, &mockNavigationService{}, &mockMarkdownService{})
}

func doRequest(s *Server, method, path string, body string) *httptest.ResponseRecorder {
	var r *http.Request
	if body != "" {
		r = httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	} else {
		r = httptest.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, r)
	return w
}

func decodeJSON(t *testing.T, w *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.NewDecoder(w.Body).Decode(v); err != nil {
		t.Fatalf("decode JSON: %v\nbody: %s", err, w.Body.String())
	}
}
