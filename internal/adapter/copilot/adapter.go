// Package copilot implements port.CopilotTaskProvider by shelling out
// to the GitHub CLI's `gh agent-task` command set (available in gh v2.80.0+).
//
// The adapter is intentionally thin: it translates CLI JSON output into
// domain types and delegates all authentication to the gh CLI.
package copilot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/cottrellashley/orbit/internal/domain"
)

// ---------------------------------------------------------------------------
// gh agent-task JSON shapes (unexported — only used for parsing)
// ---------------------------------------------------------------------------

// ghTask is the JSON structure returned by `gh agent-task list --json` and
// `gh agent-task view --json`. Field names match the gh CLI v2.80+ output.
type ghTask struct {
	ID          string     `json:"id"`                // session UUID
	Name        string     `json:"name"`              // session name
	PRNumber    int        `json:"pullRequestNumber"` // PR number
	PRState     string     `json:"pullRequestState"`  // "OPEN", "CLOSED", "MERGED"
	PRTitle     string     `json:"pullRequestTitle"`  // PR title
	PRURL       string     `json:"pullRequestUrl"`    // PR URL
	Repo        string     `json:"repository"`        // "owner/repo"
	State       string     `json:"state"`             // "in_progress", "completed", etc.
	User        string     `json:"user"`              // username
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	CompletedAt *time.Time `json:"completedAt"`
}

// ---------------------------------------------------------------------------
// Adapter
// ---------------------------------------------------------------------------

// Adapter implements port.CopilotTaskProvider using the gh CLI.
type Adapter struct {
	lookPath func(string) (string, error)
	runCmd   func(ctx context.Context, name string, args ...string) ([]byte, error)
	ghBin    string
}

// Option configures an Adapter.
type Option func(*Adapter)

// WithLookPath overrides the path lookup function (for tests).
func WithLookPath(fn func(string) (string, error)) Option {
	return func(a *Adapter) { a.lookPath = fn }
}

// WithRunCmd overrides the command runner (for tests).
func WithRunCmd(fn func(ctx context.Context, name string, args ...string) ([]byte, error)) Option {
	return func(a *Adapter) { a.runCmd = fn }
}

// WithGHBinary overrides the gh binary name (for tests).
func WithGHBinary(bin string) Option {
	return func(a *Adapter) { a.ghBin = bin }
}

// NewAdapter creates a CopilotTaskProvider backed by the gh CLI.
func NewAdapter(opts ...Option) *Adapter {
	a := &Adapter{
		lookPath: exec.LookPath,
		ghBin:    "gh",
	}
	for _, o := range opts {
		o(a)
	}
	if a.runCmd == nil {
		a.runCmd = defaultRunCmd
	}
	return a
}

func defaultRunCmd(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

// ---------------------------------------------------------------------------
// CopilotTaskProvider implementation
// ---------------------------------------------------------------------------

// IsAvailable reports whether the gh CLI supports agent-task commands.
func (a *Adapter) IsAvailable() bool {
	_, err := a.lookPath(a.ghBin)
	if err != nil {
		return false
	}
	// Probe for the agent-task subcommand.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := a.runCmd(ctx, a.ghBin, "agent-task", "--help")
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "agent-task")
}

// ghJSONFields is the set of fields requested from gh agent-task --json.
const ghJSONFields = "id,name,pullRequestNumber,pullRequestState,pullRequestTitle,pullRequestUrl,repository,state,createdAt,updatedAt,completedAt,user"

// ListTasks returns Copilot agent tasks visible to the authenticated user.
func (a *Adapter) ListTasks(ctx context.Context, owner, repo string) ([]domain.CopilotTask, error) {
	if !a.IsAvailable() {
		return nil, domain.ErrCopilotUnavailable
	}

	args := []string{"agent-task", "list", "--json", ghJSONFields}

	if owner != "" && repo != "" {
		args = append(args, "--repo", owner+"/"+repo)
	}

	out, err := a.runCmd(ctx, a.ghBin, args...)
	if err != nil {
		return nil, fmt.Errorf("gh agent-task list: %w — %s", err, strings.TrimSpace(string(out)))
	}

	var tasks []ghTask
	if err := json.Unmarshal(out, &tasks); err != nil {
		return nil, fmt.Errorf("parsing agent-task list output: %w", err)
	}

	result := make([]domain.CopilotTask, len(tasks))
	for i, t := range tasks {
		result[i] = toDomainTask(t)
	}
	return result, nil
}

// GetTask returns a single Copilot agent task by session ID (UUID).
func (a *Adapter) GetTask(ctx context.Context, sessionID string) (*domain.CopilotTask, error) {
	if !a.IsAvailable() {
		return nil, domain.ErrCopilotUnavailable
	}

	args := []string{"agent-task", "view", sessionID, "--json", ghJSONFields}

	out, err := a.runCmd(ctx, a.ghBin, args...)
	if err != nil {
		errStr := strings.TrimSpace(string(out))
		if strings.Contains(errStr, "not found") || strings.Contains(errStr, "Could not resolve") || strings.Contains(errStr, "session not found") {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("gh agent-task view: %w — %s", err, errStr)
	}

	var t ghTask
	if err := json.Unmarshal(out, &t); err != nil {
		return nil, fmt.Errorf("parsing agent-task view output: %w", err)
	}

	task := toDomainTask(t)
	return &task, nil
}

// CreateTask starts a new Copilot coding agent task.
func (a *Adapter) CreateTask(ctx context.Context, opts domain.CopilotTaskCreateOpts) (*domain.CopilotTask, error) {
	if !a.IsAvailable() {
		return nil, domain.ErrCopilotUnavailable
	}

	if opts.Owner == "" || opts.Repo == "" || opts.Prompt == "" {
		return nil, fmt.Errorf("owner, repo, and prompt are required")
	}

	args := []string{"agent-task", "create",
		opts.Prompt,
		"--repo", opts.Owner + "/" + opts.Repo,
	}
	if opts.BaseBranch != "" {
		args = append(args, "--base", opts.BaseBranch)
	}

	out, err := a.runCmd(ctx, a.ghBin, args...)
	if err != nil {
		return nil, fmt.Errorf("gh agent-task create: %w — %s", err, strings.TrimSpace(string(out)))
	}

	// The create command may return JSON or a URL. Try JSON first.
	var t ghTask
	if jsonErr := json.Unmarshal(out, &t); jsonErr == nil {
		task := toDomainTask(t)
		return &task, nil
	}

	// Fallback: parse the output as a URL or message containing PR info.
	task := &domain.CopilotTask{
		Owner:     opts.Owner,
		Repo:      opts.Repo,
		Prompt:    opts.Prompt,
		Status:    domain.CopilotTaskRunning,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Try to extract a URL from the output.
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "https://github.com/") {
			task.HTMLURL = line
			// Try to extract PR number from URL: .../pull/123
			if parts := strings.Split(line, "/pull/"); len(parts) == 2 {
				if n, err := strconv.Atoi(strings.TrimRight(parts[1], "/ \n")); err == nil {
					task.PRNumber = n
					task.ID = fmt.Sprintf("%s/%s#%d", opts.Owner, opts.Repo, n)
				}
			}
			break
		}
	}

	return task, nil
}

// StopTask stops a running Copilot agent task by closing its pull request.
// The session ID is used to look up the PR number via `gh agent-task view`,
// then `gh pr close` is called to terminate the agent session.
func (a *Adapter) StopTask(ctx context.Context, sessionID string) error {
	if !a.IsAvailable() {
		return domain.ErrCopilotUnavailable
	}

	// First, look up the task to get repo and PR number.
	task, err := a.GetTask(ctx, sessionID)
	if err != nil {
		return err
	}

	args := []string{"pr", "close",
		"--repo", task.Owner + "/" + task.Repo,
		strconv.Itoa(task.PRNumber),
	}

	out, err := a.runCmd(ctx, a.ghBin, args...)
	if err != nil {
		errStr := strings.TrimSpace(string(out))
		if strings.Contains(errStr, "not found") || strings.Contains(errStr, "Could not resolve") {
			return domain.ErrNotFound
		}
		return fmt.Errorf("gh pr close: %w — %s", err, errStr)
	}
	return nil
}

// TaskLogs returns the session log for a task as a stream.
func (a *Adapter) TaskLogs(ctx context.Context, sessionID string) (io.ReadCloser, error) {
	if !a.IsAvailable() {
		return nil, domain.ErrCopilotUnavailable
	}

	args := []string{"agent-task", "view", sessionID, "--log"}

	out, err := a.runCmd(ctx, a.ghBin, args...)
	if err != nil {
		errStr := strings.TrimSpace(string(out))
		if strings.Contains(errStr, "not found") || strings.Contains(errStr, "session not found") {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("gh agent-task view --log: %w — %s", err, errStr)
	}

	return io.NopCloser(bytes.NewReader(out)), nil
}

// ---------------------------------------------------------------------------
// Conversion helpers
// ---------------------------------------------------------------------------

// toDomainTask converts a ghTask (CLI JSON) to a domain.CopilotTask.
func toDomainTask(t ghTask) domain.CopilotTask {
	owner, repo := splitRepo(t.Repo)

	status := inferStatus(t)

	return domain.CopilotTask{
		ID:        t.ID,
		Owner:     owner,
		Repo:      repo,
		PRNumber:  t.PRNumber,
		Title:     t.PRTitle,
		Prompt:    t.Name, // session name is the closest to the original prompt
		Status:    status,
		HTMLURL:   t.PRURL,
		Branch:    "",    // not available in gh agent-task JSON
		Draft:     false, // not available in gh agent-task JSON
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

// splitRepo splits "owner/repo" into owner, repo.
func splitRepo(fullName string) (string, string) {
	parts := strings.SplitN(fullName, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return fullName, ""
}

// inferStatus maps gh CLI agent task state to a domain status.
func inferStatus(t ghTask) domain.CopilotTaskStatus {
	// The State field contains the agent task status directly.
	if t.State != "" {
		return domain.ParseCopilotTaskStatus(t.State)
	}

	// Fallback: infer from PR state.
	switch strings.ToUpper(t.PRState) {
	case "OPEN":
		return domain.CopilotTaskRunning
	case "CLOSED":
		return domain.CopilotTaskStopped
	case "MERGED":
		return domain.CopilotTaskCompleted
	default:
		return domain.CopilotTaskUnknown
	}
}
