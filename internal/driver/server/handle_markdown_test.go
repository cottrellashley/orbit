package server

import (
	"net/http"
	"testing"
)

func TestHandleMarkdownRender(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})

	body := `{"source":"# Hello"}`
	w := doRequest(s, "POST", "/api/markdown/render", body)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var result struct {
		HTML     string `json:"html"`
		Fallback bool   `json:"fallback"`
	}
	decodeJSON(t, w, &result)
	if result.HTML == "" {
		t.Fatal("expected non-empty HTML")
	}
	if !result.Fallback {
		t.Fatal("expected fallback to be true")
	}
}

func TestHandleMarkdownRender_EmptySource(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	w := doRequest(s, "POST", "/api/markdown/render", `{"source":""}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleMarkdownRender_InvalidJSON(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	w := doRequest(s, "POST", "/api/markdown/render", `{bad}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
