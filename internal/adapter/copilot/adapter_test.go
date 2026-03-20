package copilot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
)

// ---------------------------------------------------------------------------
// test helpers
// ---------------------------------------------------------------------------

func lookPathFound(path string) func(string) (string, error) {
	return func(_ string) (string, error) { return path, nil }
}

func lookPathNotFound() func(string) (string, error) {
	return func(name string) (string, error) {
		return "", fmt.Errorf("%s: not found", name)
	}
}

func runCmdReturns(out string, err error) func(context.Context, string, ...string) ([]byte, error) {
	return func(_ context.Context, _ string, _ ...string) ([]byte, error) {
		return []byte(out), err
	}
}

func runCmdDispatch(handlers map[string]func() ([]byte, error)) func(context.Context, string, ...string) ([]byte, error) {
	return func(_ context.Context, name string, args ...string) ([]byte, error) {
		key := name + " " + strings.Join(args, " ")
		for prefix, handler := range handlers {
			if strings.HasPrefix(key, prefix) {
				return handler()
			}
		}
		return nil, fmt.Errorf("unhandled command: %s", key)
	}
}

// sampleTasksJSON returns a JSON array of ghTask matching real gh CLI output.
func sampleTasksJSON() string {
	tasks := []ghTask{
		{
			ID:        "303acf6d-db68-4f0a-bcd2-0765bd7fbecd",
			Name:      "Working on user profile page layout updates",
			PRNumber:  42,
			PRState:   "OPEN",
			PRTitle:   "[WIP] Add dark mode toggle",
			PRURL:     "https://github.com/acme/app/pull/42",
			Repo:      "acme/app",
			State:     "in_progress",
			User:      "testuser",
			CreatedAt: time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 3, 20, 10, 5, 0, 0, time.UTC),
		},
		{
			ID:        "16ebee43-ffd1-4ffd-8e5c-065e1ccb3184",
			Name:      "Fixing null pointer in auth handler",
			PRNumber:  43,
			PRState:   "OPEN",
			PRTitle:   "Fix null pointer in auth handler",
			PRURL:     "https://github.com/acme/app/pull/43",
			Repo:      "acme/app",
			State:     "completed",
			User:      "testuser",
			CreatedAt: time.Date(2026, 3, 19, 8, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 3, 19, 9, 0, 0, 0, time.UTC),
		},
	}
	data, _ := json.Marshal(tasks)
	return string(data)
}

func sampleTaskJSON() string {
	t := ghTask{
		ID:        "303acf6d-db68-4f0a-bcd2-0765bd7fbecd",
		Name:      "Working on user profile page layout updates",
		PRNumber:  42,
		PRState:   "OPEN",
		PRTitle:   "[WIP] Add dark mode toggle",
		PRURL:     "https://github.com/acme/app/pull/42",
		Repo:      "acme/app",
		State:     "in_progress",
		User:      "testuser",
		CreatedAt: time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 3, 20, 10, 5, 0, 0, time.UTC),
	}
	data, _ := json.Marshal(t)
	return string(data)
}

// ---------------------------------------------------------------------------
// IsAvailable
// ---------------------------------------------------------------------------

func TestAdapter_IsAvailable_True(t *testing.T) {
	a := NewAdapter(
		WithLookPath(lookPathFound("/usr/local/bin/gh")),
		WithRunCmd(runCmdReturns("agent-task - Manage Copilot agent tasks", nil)),
	)
	if !a.IsAvailable() {
		t.Fatal("expected IsAvailable to return true")
	}
}

func TestAdapter_IsAvailable_NoGH(t *testing.T) {
	a := NewAdapter(
		WithLookPath(lookPathNotFound()),
	)
	if a.IsAvailable() {
		t.Fatal("expected IsAvailable to return false when gh is not on PATH")
	}
}

func TestAdapter_IsAvailable_OldGH(t *testing.T) {
	a := NewAdapter(
		WithLookPath(lookPathFound("/usr/local/bin/gh")),
		WithRunCmd(runCmdReturns("unknown command \"agent-task\" for \"gh\"", fmt.Errorf("exit 1"))),
	)
	if a.IsAvailable() {
		t.Fatal("expected IsAvailable to return false for old gh version")
	}
}

// ---------------------------------------------------------------------------
// ListTasks
// ---------------------------------------------------------------------------

func TestAdapter_ListTasks_Success(t *testing.T) {
	a := NewAdapter(
		WithLookPath(lookPathFound("/usr/local/bin/gh")),
		WithRunCmd(runCmdDispatch(map[string]func() ([]byte, error){
			"gh agent-task --help": func() ([]byte, error) {
				return []byte("agent-task - Manage Copilot agent tasks"), nil
			},
			"gh agent-task list": func() ([]byte, error) {
				return []byte(sampleTasksJSON()), nil
			},
		})),
	)

	tasks, err := a.ListTasks(context.Background(), "acme", "app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	if tasks[0].PRNumber != 42 {
		t.Fatalf("expected PR #42, got #%d", tasks[0].PRNumber)
	}
	if tasks[0].Owner != "acme" {
		t.Fatalf("expected owner 'acme', got %q", tasks[0].Owner)
	}
	if tasks[0].Repo != "app" {
		t.Fatalf("expected repo 'app', got %q", tasks[0].Repo)
	}
	if tasks[0].Status != domain.CopilotTaskRunning {
		t.Fatalf("expected running (in_progress state), got %s", tasks[0].Status)
	}
	if tasks[1].Status != domain.CopilotTaskCompleted {
		t.Fatalf("expected completed, got %s", tasks[1].Status)
	}
	if tasks[0].ID != "303acf6d-db68-4f0a-bcd2-0765bd7fbecd" {
		t.Fatalf("expected session UUID as ID, got %q", tasks[0].ID)
	}
	if tasks[0].Prompt != "Working on user profile page layout updates" {
		t.Fatalf("expected session name as prompt, got %q", tasks[0].Prompt)
	}
}

func TestAdapter_ListTasks_Unavailable(t *testing.T) {
	a := NewAdapter(
		WithLookPath(lookPathNotFound()),
	)

	_, err := a.ListTasks(context.Background(), "", "")
	if !errors.Is(err, domain.ErrCopilotUnavailable) {
		t.Fatalf("expected ErrCopilotUnavailable, got %v", err)
	}
}

func TestAdapter_ListTasks_CLIError(t *testing.T) {
	a := NewAdapter(
		WithLookPath(lookPathFound("/usr/local/bin/gh")),
		WithRunCmd(runCmdDispatch(map[string]func() ([]byte, error){
			"gh agent-task --help": func() ([]byte, error) {
				return []byte("agent-task - Manage Copilot agent tasks"), nil
			},
			"gh agent-task list": func() ([]byte, error) {
				return []byte("authentication required"), fmt.Errorf("exit 1")
			},
		})),
	)

	_, err := a.ListTasks(context.Background(), "", "")
	if err == nil {
		t.Fatal("expected error")
	}
	if errors.Is(err, domain.ErrCopilotUnavailable) {
		t.Fatal("expected CLI error, not unavailable")
	}
}

// ---------------------------------------------------------------------------
// GetTask
// ---------------------------------------------------------------------------

func TestAdapter_GetTask_Success(t *testing.T) {
	a := NewAdapter(
		WithLookPath(lookPathFound("/usr/local/bin/gh")),
		WithRunCmd(runCmdDispatch(map[string]func() ([]byte, error){
			"gh agent-task --help": func() ([]byte, error) {
				return []byte("agent-task - Manage Copilot agent tasks"), nil
			},
			"gh agent-task view": func() ([]byte, error) {
				return []byte(sampleTaskJSON()), nil
			},
		})),
	)

	task, err := a.GetTask(context.Background(), "303acf6d-db68-4f0a-bcd2-0765bd7fbecd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.PRNumber != 42 {
		t.Fatalf("expected PR #42, got #%d", task.PRNumber)
	}
	if task.Title != "[WIP] Add dark mode toggle" {
		t.Fatalf("unexpected title: %q", task.Title)
	}
	if task.ID != "303acf6d-db68-4f0a-bcd2-0765bd7fbecd" {
		t.Fatalf("unexpected ID: %q", task.ID)
	}
}

func TestAdapter_GetTask_NotFound(t *testing.T) {
	a := NewAdapter(
		WithLookPath(lookPathFound("/usr/local/bin/gh")),
		WithRunCmd(runCmdDispatch(map[string]func() ([]byte, error){
			"gh agent-task --help": func() ([]byte, error) {
				return []byte("agent-task - Manage Copilot agent tasks"), nil
			},
			"gh agent-task view": func() ([]byte, error) {
				return []byte("Could not resolve to a pull request"), fmt.Errorf("exit 1")
			},
		})),
	)

	_, err := a.GetTask(context.Background(), "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestAdapter_GetTask_Unavailable(t *testing.T) {
	a := NewAdapter(
		WithLookPath(lookPathNotFound()),
	)

	_, err := a.GetTask(context.Background(), "303acf6d-db68-4f0a-bcd2-0765bd7fbecd")
	if !errors.Is(err, domain.ErrCopilotUnavailable) {
		t.Fatalf("expected ErrCopilotUnavailable, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// CreateTask
// ---------------------------------------------------------------------------

func TestAdapter_CreateTask_JSONResponse(t *testing.T) {
	a := NewAdapter(
		WithLookPath(lookPathFound("/usr/local/bin/gh")),
		WithRunCmd(runCmdDispatch(map[string]func() ([]byte, error){
			"gh agent-task --help": func() ([]byte, error) {
				return []byte("agent-task - Manage Copilot agent tasks"), nil
			},
			"gh agent-task create": func() ([]byte, error) {
				return []byte(sampleTaskJSON()), nil
			},
		})),
	)

	task, err := a.CreateTask(context.Background(), domain.CopilotTaskCreateOpts{
		Owner:  "acme",
		Repo:   "app",
		Prompt: "Add dark mode support",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.PRNumber != 42 {
		t.Fatalf("expected PR #42, got #%d", task.PRNumber)
	}
}

func TestAdapter_CreateTask_URLResponse(t *testing.T) {
	a := NewAdapter(
		WithLookPath(lookPathFound("/usr/local/bin/gh")),
		WithRunCmd(runCmdDispatch(map[string]func() ([]byte, error){
			"gh agent-task --help": func() ([]byte, error) {
				return []byte("agent-task - Manage Copilot agent tasks"), nil
			},
			"gh agent-task create": func() ([]byte, error) {
				return []byte("Creating task...\nhttps://github.com/acme/app/pull/55\n"), nil
			},
		})),
	)

	task, err := a.CreateTask(context.Background(), domain.CopilotTaskCreateOpts{
		Owner:  "acme",
		Repo:   "app",
		Prompt: "Fix the bug",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.PRNumber != 55 {
		t.Fatalf("expected PR #55, got #%d", task.PRNumber)
	}
	if task.HTMLURL != "https://github.com/acme/app/pull/55" {
		t.Fatalf("unexpected URL: %q", task.HTMLURL)
	}
}

func TestAdapter_CreateTask_MissingParams(t *testing.T) {
	a := NewAdapter(
		WithLookPath(lookPathFound("/usr/local/bin/gh")),
		WithRunCmd(runCmdDispatch(map[string]func() ([]byte, error){
			"gh agent-task --help": func() ([]byte, error) {
				return []byte("agent-task - Manage Copilot agent tasks"), nil
			},
		})),
	)

	_, err := a.CreateTask(context.Background(), domain.CopilotTaskCreateOpts{})
	if err == nil {
		t.Fatal("expected error for missing params")
	}
}

func TestAdapter_CreateTask_Unavailable(t *testing.T) {
	a := NewAdapter(
		WithLookPath(lookPathNotFound()),
	)

	_, err := a.CreateTask(context.Background(), domain.CopilotTaskCreateOpts{
		Owner:  "acme",
		Repo:   "app",
		Prompt: "Fix bug",
	})
	if !errors.Is(err, domain.ErrCopilotUnavailable) {
		t.Fatalf("expected ErrCopilotUnavailable, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// StopTask (resolves via GetTask, then calls gh pr close)
// ---------------------------------------------------------------------------

func TestAdapter_StopTask_Success(t *testing.T) {
	a := NewAdapter(
		WithLookPath(lookPathFound("/usr/local/bin/gh")),
		WithRunCmd(runCmdDispatch(map[string]func() ([]byte, error){
			"gh agent-task --help": func() ([]byte, error) {
				return []byte("agent-task - Manage Copilot agent tasks"), nil
			},
			"gh agent-task view": func() ([]byte, error) {
				return []byte(sampleTaskJSON()), nil
			},
			"gh pr close": func() ([]byte, error) {
				return []byte("Closed pull request #42"), nil
			},
		})),
	)

	err := a.StopTask(context.Background(), "303acf6d-db68-4f0a-bcd2-0765bd7fbecd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAdapter_StopTask_SessionNotFound(t *testing.T) {
	a := NewAdapter(
		WithLookPath(lookPathFound("/usr/local/bin/gh")),
		WithRunCmd(runCmdDispatch(map[string]func() ([]byte, error){
			"gh agent-task --help": func() ([]byte, error) {
				return []byte("agent-task - Manage Copilot agent tasks"), nil
			},
			"gh agent-task view": func() ([]byte, error) {
				return []byte("session not found"), fmt.Errorf("exit 1")
			},
		})),
	)

	err := a.StopTask(context.Background(), "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestAdapter_StopTask_PRCloseNotFound(t *testing.T) {
	a := NewAdapter(
		WithLookPath(lookPathFound("/usr/local/bin/gh")),
		WithRunCmd(runCmdDispatch(map[string]func() ([]byte, error){
			"gh agent-task --help": func() ([]byte, error) {
				return []byte("agent-task - Manage Copilot agent tasks"), nil
			},
			"gh agent-task view": func() ([]byte, error) {
				return []byte(sampleTaskJSON()), nil
			},
			"gh pr close": func() ([]byte, error) {
				return []byte("Could not resolve to a pull request"), fmt.Errorf("exit 1")
			},
		})),
	)

	err := a.StopTask(context.Background(), "303acf6d-db68-4f0a-bcd2-0765bd7fbecd")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestAdapter_StopTask_Unavailable(t *testing.T) {
	a := NewAdapter(
		WithLookPath(lookPathNotFound()),
	)

	err := a.StopTask(context.Background(), "303acf6d-db68-4f0a-bcd2-0765bd7fbecd")
	if !errors.Is(err, domain.ErrCopilotUnavailable) {
		t.Fatalf("expected ErrCopilotUnavailable, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// TaskLogs
// ---------------------------------------------------------------------------

func TestAdapter_TaskLogs_Success(t *testing.T) {
	logContent := "Step 1: Analyzing code...\nStep 2: Making changes...\nStep 3: Running tests...\n"
	a := NewAdapter(
		WithLookPath(lookPathFound("/usr/local/bin/gh")),
		WithRunCmd(runCmdDispatch(map[string]func() ([]byte, error){
			"gh agent-task --help": func() ([]byte, error) {
				return []byte("agent-task - Manage Copilot agent tasks"), nil
			},
			"gh agent-task view": func() ([]byte, error) {
				return []byte(logContent), nil
			},
		})),
	)

	rc, err := a.TaskLogs(context.Background(), "303acf6d-db68-4f0a-bcd2-0765bd7fbecd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("error reading logs: %v", err)
	}
	if string(data) != logContent {
		t.Fatalf("unexpected logs: %q", string(data))
	}
}

func TestAdapter_TaskLogs_NotFound(t *testing.T) {
	a := NewAdapter(
		WithLookPath(lookPathFound("/usr/local/bin/gh")),
		WithRunCmd(runCmdDispatch(map[string]func() ([]byte, error){
			"gh agent-task --help": func() ([]byte, error) {
				return []byte("agent-task - Manage Copilot agent tasks"), nil
			},
			"gh agent-task view": func() ([]byte, error) {
				return []byte("not found"), fmt.Errorf("exit 1")
			},
		})),
	)

	_, err := a.TaskLogs(context.Background(), "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestAdapter_TaskLogs_Unavailable(t *testing.T) {
	a := NewAdapter(
		WithLookPath(lookPathNotFound()),
	)

	_, err := a.TaskLogs(context.Background(), "303acf6d-db68-4f0a-bcd2-0765bd7fbecd")
	if !errors.Is(err, domain.ErrCopilotUnavailable) {
		t.Fatalf("expected ErrCopilotUnavailable, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// toDomainTask / inferStatus
// ---------------------------------------------------------------------------

func TestInferStatus_Variants(t *testing.T) {
	tests := []struct {
		name   string
		task   ghTask
		expect domain.CopilotTaskStatus
	}{
		{"state running", ghTask{State: "running"}, domain.CopilotTaskRunning},
		{"state in_progress", ghTask{State: "in_progress"}, domain.CopilotTaskRunning},
		{"state completed", ghTask{State: "completed"}, domain.CopilotTaskCompleted},
		{"state stopped", ghTask{State: "stopped"}, domain.CopilotTaskStopped},
		{"state failed", ghTask{State: "failed"}, domain.CopilotTaskFailed},
		{"fallback OPEN", ghTask{PRState: "OPEN"}, domain.CopilotTaskRunning},
		{"fallback CLOSED", ghTask{PRState: "CLOSED"}, domain.CopilotTaskStopped},
		{"fallback MERGED", ghTask{PRState: "MERGED"}, domain.CopilotTaskCompleted},
		{"unknown", ghTask{PRState: "something"}, domain.CopilotTaskUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferStatus(tt.task)
			if got != tt.expect {
				t.Fatalf("expected %s, got %s", tt.expect, got)
			}
		})
	}
}

func TestSplitRepo(t *testing.T) {
	owner, repo := splitRepo("acme/app")
	if owner != "acme" || repo != "app" {
		t.Fatalf("expected acme/app, got %s/%s", owner, repo)
	}

	owner, repo = splitRepo("single")
	if owner != "single" || repo != "" {
		t.Fatalf("expected single/'', got %s/%s", owner, repo)
	}
}

func TestParseCopilotTaskStatus(t *testing.T) {
	tests := []struct {
		input  string
		expect domain.CopilotTaskStatus
	}{
		{"running", domain.CopilotTaskRunning},
		{"in_progress", domain.CopilotTaskRunning},
		{"completed", domain.CopilotTaskCompleted},
		{"stopped", domain.CopilotTaskStopped},
		{"failed", domain.CopilotTaskFailed},
		{"", domain.CopilotTaskUnknown},
		{"garbage", domain.CopilotTaskUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := domain.ParseCopilotTaskStatus(tt.input)
			if got != tt.expect {
				t.Fatalf("expected %s, got %s", tt.expect, got)
			}
		})
	}
}
