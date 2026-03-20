package server

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
)

// ---------------------------------------------------------------------------
// handleCopilotAvailable
// ---------------------------------------------------------------------------

func TestHandleCopilotAvailable_True(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	s.SetCopilotService(&mockCopilotService{available: true})

	w := doRequest(s, "GET", "/api/copilot/available", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var out copilotAvailableJSON
	decodeJSON(t, w, &out)
	if !out.Available {
		t.Fatal("expected available=true")
	}
}

func TestHandleCopilotAvailable_False(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	s.SetCopilotService(&mockCopilotService{available: false})

	w := doRequest(s, "GET", "/api/copilot/available", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var out copilotAvailableJSON
	decodeJSON(t, w, &out)
	if out.Available {
		t.Fatal("expected available=false")
	}
}

func TestHandleCopilotAvailable_NilService(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	// copilotSvc not set.

	w := doRequest(s, "GET", "/api/copilot/available", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var out copilotAvailableJSON
	decodeJSON(t, w, &out)
	if out.Available {
		t.Fatal("expected available=false when service is nil")
	}
}

// ---------------------------------------------------------------------------
// handleListCopilotTasks
// ---------------------------------------------------------------------------

func TestHandleListCopilotTasks_OK(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	s.SetCopilotService(&mockCopilotService{
		available: true,
		tasks: []domain.CopilotTask{
			{
				ID: "303acf6d-db68-4f0a-bcd2-0765bd7fbecd", Owner: "org", Repo: "repo", PRNumber: 1,
				Title: "Fix bug", Status: domain.CopilotTaskRunning,
				HTMLURL:   "https://github.com/org/repo/pull/1",
				CreatedAt: time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
			},
		},
	})

	w := doRequest(s, "GET", "/api/copilot/tasks", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var tasks []copilotTaskJSON
	decodeJSON(t, w, &tasks)
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].ID != "303acf6d-db68-4f0a-bcd2-0765bd7fbecd" {
		t.Fatalf("expected session UUID, got %q", tasks[0].ID)
	}
	if tasks[0].Status != "running" {
		t.Fatalf("expected status 'running', got %q", tasks[0].Status)
	}
}

func TestHandleListCopilotTasks_Empty(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	s.SetCopilotService(&mockCopilotService{
		available: true,
		tasks:     []domain.CopilotTask{},
	})

	w := doRequest(s, "GET", "/api/copilot/tasks", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var tasks []copilotTaskJSON
	decodeJSON(t, w, &tasks)
	if len(tasks) != 0 {
		t.Fatalf("expected 0 tasks, got %d", len(tasks))
	}
}

func TestHandleListCopilotTasks_ServiceUnavailable(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	// copilotSvc not set.

	w := doRequest(s, "GET", "/api/copilot/tasks", "")
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestHandleListCopilotTasks_ProviderError(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	s.SetCopilotService(&mockCopilotService{
		available: true,
		listErr:   fmt.Errorf("api error"),
	})

	w := doRequest(s, "GET", "/api/copilot/tasks", "")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// handleGetCopilotTask
// ---------------------------------------------------------------------------

func TestHandleGetCopilotTask_OK(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	s.SetCopilotService(&mockCopilotService{
		available: true,
		task: &domain.CopilotTask{
			ID: "303acf6d-db68-4f0a-bcd2-0765bd7fbecd", Owner: "org", Repo: "repo", PRNumber: 42,
			Title: "Add feature", Status: domain.CopilotTaskCompleted,
			HTMLURL:   "https://github.com/org/repo/pull/42",
			CreatedAt: time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		},
	})

	w := doRequest(s, "GET", "/api/copilot/tasks/303acf6d-db68-4f0a-bcd2-0765bd7fbecd", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var task copilotTaskJSON
	decodeJSON(t, w, &task)
	if task.ID != "303acf6d-db68-4f0a-bcd2-0765bd7fbecd" {
		t.Fatalf("expected session UUID, got %q", task.ID)
	}
	if task.Status != "completed" {
		t.Fatalf("expected status 'completed', got %q", task.Status)
	}
}

func TestHandleGetCopilotTask_EmptyID(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	s.SetCopilotService(&mockCopilotService{
		available: true,
		getErr:    fmt.Errorf("session ID is required"),
	})

	// The mux won't match an empty segment, so this hits the handler with whatever
	// error the service returns. We test the service-level empty-ID guard via
	// the app test; here we just ensure a non-empty ID reaches the handler.
	w := doRequest(s, "GET", "/api/copilot/tasks/some-id", "")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleGetCopilotTask_NotFound(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	s.SetCopilotService(&mockCopilotService{
		available: true,
		getErr:    domain.ErrNotFound,
	})

	w := doRequest(s, "GET", "/api/copilot/tasks/303acf6d-db68-4f0a-bcd2-0765bd7fbecd", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleGetCopilotTask_ServiceUnavailable(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	// copilotSvc not set.

	w := doRequest(s, "GET", "/api/copilot/tasks/303acf6d-db68-4f0a-bcd2-0765bd7fbecd", "")
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// handleCreateCopilotTask
// ---------------------------------------------------------------------------

func TestHandleCreateCopilotTask_OK(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	s.SetCopilotService(&mockCopilotService{
		available: true,
		task: &domain.CopilotTask{
			ID: "16ebee43-ffd1-4ffd-8e5c-065e1ccb3184", Owner: "org", Repo: "repo", PRNumber: 10,
			Title: "New task", Status: domain.CopilotTaskRunning,
			HTMLURL:   "https://github.com/org/repo/pull/10",
			CreatedAt: time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		},
	})

	body := `{"owner":"org","repo":"repo","prompt":"Add tests"}`
	w := doRequest(s, "POST", "/api/copilot/tasks", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var task copilotTaskJSON
	decodeJSON(t, w, &task)
	if task.PRNumber != 10 {
		t.Fatalf("expected PR 10, got %d", task.PRNumber)
	}
}

func TestHandleCreateCopilotTask_BadJSON(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	s.SetCopilotService(&mockCopilotService{available: true})

	w := doRequest(s, "POST", "/api/copilot/tasks", "not json")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleCreateCopilotTask_CreateError(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	s.SetCopilotService(&mockCopilotService{
		available: true,
		createErr: fmt.Errorf("owner, repo, and prompt are required"),
	})

	body := `{"owner":"","repo":"","prompt":""}`
	w := doRequest(s, "POST", "/api/copilot/tasks", body)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleCreateCopilotTask_ServiceUnavailable(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	// copilotSvc not set.

	body := `{"owner":"org","repo":"repo","prompt":"do stuff"}`
	w := doRequest(s, "POST", "/api/copilot/tasks", body)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// handleStopCopilotTask
// ---------------------------------------------------------------------------

func TestHandleStopCopilotTask_OK(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	s.SetCopilotService(&mockCopilotService{available: true})

	w := doRequest(s, "DELETE", "/api/copilot/tasks/303acf6d-db68-4f0a-bcd2-0765bd7fbecd", "")
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleStopCopilotTask_NotFound(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	s.SetCopilotService(&mockCopilotService{
		available: true,
		stopErr:   domain.ErrNotFound,
	})

	w := doRequest(s, "DELETE", "/api/copilot/tasks/303acf6d-db68-4f0a-bcd2-0765bd7fbecd", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleStopCopilotTask_ServiceUnavailable(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})

	w := doRequest(s, "DELETE", "/api/copilot/tasks/303acf6d-db68-4f0a-bcd2-0765bd7fbecd", "")
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// handleCopilotTaskLogs
// ---------------------------------------------------------------------------

func TestHandleCopilotTaskLogs_OK(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	s.SetCopilotService(&mockCopilotService{
		available: true,
		logs:      "step 1: cloning\nstep 2: editing\n",
	})

	w := doRequest(s, "GET", "/api/copilot/tasks/303acf6d-db68-4f0a-bcd2-0765bd7fbecd/logs", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Fatalf("expected text/plain content type, got %q", ct)
	}
	if w.Body.String() != "step 1: cloning\nstep 2: editing\n" {
		t.Fatalf("unexpected body: %q", w.Body.String())
	}
}

func TestHandleCopilotTaskLogs_NotFound(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})
	s.SetCopilotService(&mockCopilotService{
		available: true,
		logsErr:   domain.ErrNotFound,
	})

	w := doRequest(s, "GET", "/api/copilot/tasks/303acf6d-db68-4f0a-bcd2-0765bd7fbecd/logs", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleCopilotTaskLogs_ServiceUnavailable(t *testing.T) {
	s := newTestServer(&mockEnvironmentService{}, &mockProjectService{}, &mockSessionService{})

	w := doRequest(s, "GET", "/api/copilot/tasks/303acf6d-db68-4f0a-bcd2-0765bd7fbecd/logs", "")
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}
