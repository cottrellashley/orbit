package app

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
)

// ---------------------------------------------------------------------------
// mockCopilotTaskProvider implements copilotTaskProvider for testing.
// ---------------------------------------------------------------------------

type mockCopilotTaskProvider struct {
	available bool
	tasks     []domain.CopilotTask
	task      *domain.CopilotTask
	logs      string
	listErr   error
	getErr    error
	createErr error
	stopErr   error
	logsErr   error
}

func (m *mockCopilotTaskProvider) IsAvailable() bool { return m.available }

func (m *mockCopilotTaskProvider) ListTasks(_ context.Context, _, _ string) ([]domain.CopilotTask, error) {
	return m.tasks, m.listErr
}

func (m *mockCopilotTaskProvider) GetTask(_ context.Context, _ string) (*domain.CopilotTask, error) {
	return m.task, m.getErr
}

func (m *mockCopilotTaskProvider) CreateTask(_ context.Context, _ domain.CopilotTaskCreateOpts) (*domain.CopilotTask, error) {
	return m.task, m.createErr
}

func (m *mockCopilotTaskProvider) StopTask(_ context.Context, _ string) error {
	return m.stopErr
}

func (m *mockCopilotTaskProvider) TaskLogs(_ context.Context, _ string) (io.ReadCloser, error) {
	if m.logsErr != nil {
		return nil, m.logsErr
	}
	return io.NopCloser(strings.NewReader(m.logs)), nil
}

// ---------------------------------------------------------------------------
// IsAvailable tests
// ---------------------------------------------------------------------------

func TestCopilotService_IsAvailable_True(t *testing.T) {
	svc := NewCopilotService(&mockCopilotTaskProvider{available: true})
	if !svc.IsAvailable() {
		t.Fatal("expected available")
	}
}

func TestCopilotService_IsAvailable_False(t *testing.T) {
	svc := NewCopilotService(&mockCopilotTaskProvider{available: false})
	if svc.IsAvailable() {
		t.Fatal("expected not available")
	}
}

// ---------------------------------------------------------------------------
// ListTasks tests
// ---------------------------------------------------------------------------

func TestCopilotService_ListTasks_Success(t *testing.T) {
	tasks := []domain.CopilotTask{
		{ID: "303acf6d-db68-4f0a-bcd2-0765bd7fbecd", Owner: "org", Repo: "repo", PRNumber: 1, Status: domain.CopilotTaskRunning},
		{ID: "16ebee43-ffd1-4ffd-8e5c-065e1ccb3184", Owner: "org", Repo: "repo", PRNumber: 2, Status: domain.CopilotTaskCompleted},
	}
	svc := NewCopilotService(&mockCopilotTaskProvider{available: true, tasks: tasks})

	got, err := svc.ListTasks(context.Background(), "org", "repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(got))
	}
	if got[0].ID != "303acf6d-db68-4f0a-bcd2-0765bd7fbecd" {
		t.Fatalf("expected session UUID, got %q", got[0].ID)
	}
}

func TestCopilotService_ListTasks_Unavailable(t *testing.T) {
	svc := NewCopilotService(&mockCopilotTaskProvider{available: false})

	_, err := svc.ListTasks(context.Background(), "", "")
	if !errors.Is(err, domain.ErrCopilotUnavailable) {
		t.Fatalf("expected ErrCopilotUnavailable, got %v", err)
	}
}

func TestCopilotService_ListTasks_ProviderError(t *testing.T) {
	svc := NewCopilotService(&mockCopilotTaskProvider{
		available: true,
		listErr:   errors.New("api error"),
	})

	_, err := svc.ListTasks(context.Background(), "org", "")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "api error" {
		t.Fatalf("expected 'api error', got %q", err.Error())
	}
}

// ---------------------------------------------------------------------------
// GetTask tests
// ---------------------------------------------------------------------------

func TestCopilotService_GetTask_Success(t *testing.T) {
	task := &domain.CopilotTask{
		ID: "303acf6d-db68-4f0a-bcd2-0765bd7fbecd", Owner: "org", Repo: "repo", PRNumber: 42,
		Title: "Fix the bug", Status: domain.CopilotTaskRunning,
		CreatedAt: time.Now(),
	}
	svc := NewCopilotService(&mockCopilotTaskProvider{available: true, task: task})

	got, err := svc.GetTask(context.Background(), "303acf6d-db68-4f0a-bcd2-0765bd7fbecd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.PRNumber != 42 {
		t.Fatalf("expected PR 42, got %d", got.PRNumber)
	}
	if got.Title != "Fix the bug" {
		t.Fatalf("expected title 'Fix the bug', got %q", got.Title)
	}
}

func TestCopilotService_GetTask_Unavailable(t *testing.T) {
	svc := NewCopilotService(&mockCopilotTaskProvider{available: false})

	_, err := svc.GetTask(context.Background(), "303acf6d-db68-4f0a-bcd2-0765bd7fbecd")
	if !errors.Is(err, domain.ErrCopilotUnavailable) {
		t.Fatalf("expected ErrCopilotUnavailable, got %v", err)
	}
}

func TestCopilotService_GetTask_NotFound(t *testing.T) {
	svc := NewCopilotService(&mockCopilotTaskProvider{
		available: true,
		getErr:    domain.ErrNotFound,
	})

	_, err := svc.GetTask(context.Background(), "303acf6d-db68-4f0a-bcd2-0765bd7fbecd")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCopilotService_GetTask_EmptySessionID(t *testing.T) {
	svc := NewCopilotService(&mockCopilotTaskProvider{available: true})

	_, err := svc.GetTask(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty session ID")
	}
}

// ---------------------------------------------------------------------------
// CreateTask tests
// ---------------------------------------------------------------------------

func TestCopilotService_CreateTask_Success(t *testing.T) {
	task := &domain.CopilotTask{
		ID: "16ebee43-ffd1-4ffd-8e5c-065e1ccb3184", Owner: "org", Repo: "repo", PRNumber: 10,
		Status: domain.CopilotTaskRunning,
	}
	svc := NewCopilotService(&mockCopilotTaskProvider{available: true, task: task})

	got, err := svc.CreateTask(context.Background(), domain.CopilotTaskCreateOpts{
		Owner:  "org",
		Repo:   "repo",
		Prompt: "Add tests for the login flow",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.PRNumber != 10 {
		t.Fatalf("expected PR 10, got %d", got.PRNumber)
	}
}

func TestCopilotService_CreateTask_Unavailable(t *testing.T) {
	svc := NewCopilotService(&mockCopilotTaskProvider{available: false})

	_, err := svc.CreateTask(context.Background(), domain.CopilotTaskCreateOpts{
		Owner: "org", Repo: "repo", Prompt: "do stuff",
	})
	if !errors.Is(err, domain.ErrCopilotUnavailable) {
		t.Fatalf("expected ErrCopilotUnavailable, got %v", err)
	}
}

func TestCopilotService_CreateTask_MissingOwner(t *testing.T) {
	svc := NewCopilotService(&mockCopilotTaskProvider{available: true})

	_, err := svc.CreateTask(context.Background(), domain.CopilotTaskCreateOpts{
		Repo: "repo", Prompt: "do stuff",
	})
	if err == nil {
		t.Fatal("expected error for missing owner")
	}
}

func TestCopilotService_CreateTask_MissingRepo(t *testing.T) {
	svc := NewCopilotService(&mockCopilotTaskProvider{available: true})

	_, err := svc.CreateTask(context.Background(), domain.CopilotTaskCreateOpts{
		Owner: "org", Prompt: "do stuff",
	})
	if err == nil {
		t.Fatal("expected error for missing repo")
	}
}

func TestCopilotService_CreateTask_MissingPrompt(t *testing.T) {
	svc := NewCopilotService(&mockCopilotTaskProvider{available: true})

	_, err := svc.CreateTask(context.Background(), domain.CopilotTaskCreateOpts{
		Owner: "org", Repo: "repo",
	})
	if err == nil {
		t.Fatal("expected error for missing prompt")
	}
}

// ---------------------------------------------------------------------------
// StopTask tests
// ---------------------------------------------------------------------------

func TestCopilotService_StopTask_Success(t *testing.T) {
	svc := NewCopilotService(&mockCopilotTaskProvider{available: true})

	err := svc.StopTask(context.Background(), "303acf6d-db68-4f0a-bcd2-0765bd7fbecd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCopilotService_StopTask_Unavailable(t *testing.T) {
	svc := NewCopilotService(&mockCopilotTaskProvider{available: false})

	err := svc.StopTask(context.Background(), "303acf6d-db68-4f0a-bcd2-0765bd7fbecd")
	if !errors.Is(err, domain.ErrCopilotUnavailable) {
		t.Fatalf("expected ErrCopilotUnavailable, got %v", err)
	}
}

func TestCopilotService_StopTask_ProviderError(t *testing.T) {
	svc := NewCopilotService(&mockCopilotTaskProvider{
		available: true,
		stopErr:   errors.New("cannot stop"),
	})

	err := svc.StopTask(context.Background(), "303acf6d-db68-4f0a-bcd2-0765bd7fbecd")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "cannot stop" {
		t.Fatalf("expected 'cannot stop', got %q", err.Error())
	}
}

func TestCopilotService_StopTask_EmptySessionID(t *testing.T) {
	svc := NewCopilotService(&mockCopilotTaskProvider{available: true})

	err := svc.StopTask(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty session ID")
	}
}

// ---------------------------------------------------------------------------
// TaskLogs tests
// ---------------------------------------------------------------------------

func TestCopilotService_TaskLogs_Success(t *testing.T) {
	svc := NewCopilotService(&mockCopilotTaskProvider{
		available: true,
		logs:      "line 1\nline 2\n",
	})

	rc, err := svc.TaskLogs(context.Background(), "303acf6d-db68-4f0a-bcd2-0765bd7fbecd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer rc.Close()

	buf := new(strings.Builder)
	if _, err := io.Copy(buf, rc); err != nil {
		t.Fatalf("read error: %v", err)
	}
	if buf.String() != "line 1\nline 2\n" {
		t.Fatalf("unexpected logs: %q", buf.String())
	}
}

func TestCopilotService_TaskLogs_Unavailable(t *testing.T) {
	svc := NewCopilotService(&mockCopilotTaskProvider{available: false})

	_, err := svc.TaskLogs(context.Background(), "303acf6d-db68-4f0a-bcd2-0765bd7fbecd")
	if !errors.Is(err, domain.ErrCopilotUnavailable) {
		t.Fatalf("expected ErrCopilotUnavailable, got %v", err)
	}
}

func TestCopilotService_TaskLogs_ProviderError(t *testing.T) {
	svc := NewCopilotService(&mockCopilotTaskProvider{
		available: true,
		logsErr:   domain.ErrNotFound,
	})

	_, err := svc.TaskLogs(context.Background(), "303acf6d-db68-4f0a-bcd2-0765bd7fbecd")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCopilotService_TaskLogs_EmptySessionID(t *testing.T) {
	svc := NewCopilotService(&mockCopilotTaskProvider{available: true})

	_, err := svc.TaskLogs(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty session ID")
	}
}
