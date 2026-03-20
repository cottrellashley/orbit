import { useState, useCallback, useEffect } from 'react'
import { useApp } from '../state/context'
import { MainHeader } from '../components/layout/MainHeader'
import { Table, THead, TBody, TRow, TH, TD } from '../components/ui/Table'
import { StatusBadge } from '../components/ui/Badge'
import { EmptyState } from '../components/ui/EmptyState'
import { Button } from '../components/ui/Button'
import { formatTime } from '../utils'
import * as api from '../api/client'
import type { CopilotTask } from '../state/types'

// ---------------------------------------------------------------------------
// Task List
// ---------------------------------------------------------------------------

function TaskList({
  tasks,
  available,
  onSelect,
  onShowCreate,
}: {
  tasks: CopilotTask[]
  available: boolean
  onSelect: (t: CopilotTask) => void
  onShowCreate: () => void
}) {
  const running = tasks.filter((t) => t.status === 'running')
  const rest = tasks.filter((t) => t.status !== 'running')

  if (!available) {
    return (
      <div className="flex flex-col flex-1 min-h-0">
        <MainHeader title="Copilot Tasks" />
        <div className="flex-1 overflow-y-auto px-4 py-3">
          <EmptyState
            title="Copilot unavailable"
            description="gh CLI v2.80.0+ with agent-task support is required. Run `gh --version` to check."
          />
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col flex-1 min-h-0">
      <MainHeader
        title="Copilot Tasks"
        actions={<Button variant="primary" size="sm" onClick={onShowCreate}>New Task</Button>}
      />
      <div className="flex-1 overflow-y-auto px-4 py-3">
        {tasks.length === 0 ? (
          <EmptyState
            title="No Copilot tasks"
            description="Create a task to start a GitHub Copilot coding agent session."
            action={<Button variant="primary" onClick={onShowCreate}>New Task</Button>}
          />
        ) : (
          <div className="space-y-6">
            {running.length > 0 && (
              <TaskTable label={`Running (${running.length})`} tasks={running} onSelect={onSelect} />
            )}
            {rest.length > 0 && (
              <TaskTable label={`History (${rest.length})`} tasks={rest} onSelect={onSelect} />
            )}
          </div>
        )}
      </div>
    </div>
  )
}

function TaskTable({
  label,
  tasks,
  onSelect,
}: {
  label: string
  tasks: CopilotTask[]
  onSelect: (t: CopilotTask) => void
}) {
  return (
    <section>
      <div className="text-caption-xs text-txt-quaternary uppercase tracking-[0.06em] px-3 mb-1">
        {label}
      </div>
      <Table>
        <THead>
          <TRow>
            <TH>Task</TH>
            <TH>Repo</TH>
            <TH>Status</TH>
            <TH>PR</TH>
            <TH>Created</TH>
          </TRow>
        </THead>
        <TBody>
          {tasks.map((t) => (
            <TRow key={t.id} clickable onClick={() => onSelect(t)}>
              <TD>
                <span className="text-caption text-txt">{t.title || 'Untitled'}</span>
                <span className="font-mono text-caption-xs text-txt-quaternary ml-1.5">#{t.prNumber}</span>
              </TD>
              <TD>
                <span className="font-mono text-caption-xs text-txt-tertiary">{t.owner}/{t.repo}</span>
              </TD>
              <TD>
                <StatusBadge status={t.status} />
              </TD>
              <TD>
                <a
                  href={t.htmlUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-caption-xs text-accent hover:text-accent/80 transition-colors"
                  onClick={(e) => e.stopPropagation()}
                >
                  View PR
                </a>
              </TD>
              <TD>
                <span className="text-caption-xs text-txt-quaternary">{formatTime(t.createdAt)}</span>
              </TD>
            </TRow>
          ))}
        </TBody>
      </Table>
    </section>
  )
}

// ---------------------------------------------------------------------------
// Task Detail
// ---------------------------------------------------------------------------

function TaskDetail({
  task,
  onBack,
  onRefresh,
}: {
  task: CopilotTask
  onBack: () => void
  onRefresh: () => void
}) {
  const [logs, setLogs] = useState<string | null>(null)
  const [logsLoading, setLogsLoading] = useState(false)
  const [logsError, setLogsError] = useState<string | null>(null)
  const [stopping, setStopping] = useState(false)

  const loadLogs = useCallback(async () => {
    setLogsLoading(true)
    setLogsError(null)
    try {
      const text = await api.fetchCopilotTaskLogs(task.id)
      setLogs(text)
    } catch (e) {
      setLogsError(e instanceof Error ? e.message : 'Failed to load logs')
    } finally {
      setLogsLoading(false)
    }
  }, [task.id])

  const handleStop = useCallback(async () => {
    if (!confirm(`Stop task ${task.id}?`)) return
    setStopping(true)
    try {
      await api.stopCopilotTask(task.id)
      onRefresh()
    } catch {
      // ignore — refresh will show current state
    } finally {
      setStopping(false)
    }
  }, [task.id, onRefresh])

  return (
    <div className="flex flex-col flex-1 min-h-0">
        <MainHeader
        title={task.title || task.id}
        breadcrumb={{ label: 'Copilot Tasks', onClick: onBack }}
        actions={
          <div className="flex items-center gap-1.5">
            {task.status === 'running' && (
              <Button variant="danger" size="sm" disabled={stopping} onClick={handleStop}>
                {stopping ? 'Stopping...' : 'Stop'}
              </Button>
            )}
            <a
              href={task.htmlUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="text-caption-xs text-accent hover:text-accent/80 transition-colors"
            >
              View PR
            </a>
          </div>
        }
      />
      <div className="flex-1 overflow-y-auto px-4 py-3 space-y-4">
        {/* Metadata */}
        <div className="grid grid-cols-[auto_1fr] gap-x-6 gap-y-1 text-caption">
          <span className="text-txt-quaternary">Repo</span>
          <span className="font-mono text-caption-xs text-txt-secondary">{task.owner}/{task.repo}</span>
          <span className="text-txt-quaternary">PR</span>
          <span className="font-mono text-caption-xs text-txt-secondary">#{task.prNumber}</span>
          <span className="text-txt-quaternary">Status</span>
          <StatusBadge status={task.status} />
          <span className="text-txt-quaternary">Created</span>
          <span className="text-caption-xs text-txt-tertiary">{formatTime(task.createdAt)}</span>
          <span className="text-txt-quaternary">Updated</span>
          <span className="text-caption-xs text-txt-tertiary">{formatTime(task.updatedAt)}</span>
        </div>

        {/* Prompt */}
        {task.prompt && (
          <section>
            <div className="text-caption-xs text-txt-quaternary uppercase tracking-wider mb-1">Prompt</div>
            <div className="text-caption text-txt-secondary whitespace-pre-wrap">{task.prompt}</div>
          </section>
        )}

        {/* Logs */}
        <section>
          <div className="flex items-center gap-2 mb-1">
            <span className="text-caption-xs text-txt-quaternary uppercase tracking-wider">Logs</span>
            <Button size="sm" onClick={loadLogs} disabled={logsLoading}>
              {logsLoading ? 'Loading...' : logs !== null ? 'Reload' : 'Load Logs'}
            </Button>
          </div>
          {logsError && (
            <div className="text-caption-xs text-red-400 bg-red-400/10 px-3 py-2 mb-1">{logsError}</div>
          )}
          {logs !== null && (
            <pre className="text-caption-xs text-txt-tertiary font-mono whitespace-pre-wrap bg-white/[0.02] p-3 overflow-x-auto max-h-[400px] overflow-y-auto">
              {logs || '(empty)'}
            </pre>
          )}
        </section>
      </div>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Create Task Form
// ---------------------------------------------------------------------------

function CreateTask({
  repos,
  onBack,
  onCreated,
}: {
  repos: { owner: string; name: string }[]
  onBack: () => void
  onCreated: () => void
}) {
  const [owner, setOwner] = useState('')
  const [repo, setRepo] = useState('')
  const [prompt, setPrompt] = useState('')
  const [baseBranch, setBaseBranch] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Auto-fill owner/repo from first available repo
  useEffect(() => {
    if (repos.length > 0 && !owner && !repo) {
      setOwner(repos[0].owner)
      setRepo(repos[0].name)
    }
  }, [repos, owner, repo])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!owner || !repo || !prompt.trim()) return
    setSubmitting(true)
    setError(null)
    try {
      await api.createCopilotTask({
        owner,
        repo,
        prompt: prompt.trim(),
        baseBranch: baseBranch || undefined,
      })
      onCreated()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create task')
    } finally {
      setSubmitting(false)
    }
  }

  const inputClasses =
    'w-full bg-white/[0.04] text-caption text-txt px-2.5 py-[6px] border border-border/50 focus:border-accent/50 focus:outline-none transition-colors'

  return (
    <div className="flex flex-col flex-1 min-h-0">
      <MainHeader title="New Copilot Task" breadcrumb={{ label: 'Copilot Tasks', onClick: onBack }} />
      <div className="flex-1 overflow-y-auto px-4 py-3">
        <form onSubmit={handleSubmit} className="max-w-[480px] space-y-3">
          {error && (
            <div className="text-caption-xs text-red-400 bg-red-400/10 px-3 py-2">{error}</div>
          )}

          {/* Repo selector — if we have repos, show a select; otherwise manual input */}
          {repos.length > 0 ? (
            <div>
              <label className="block text-caption-xs text-txt-quaternary mb-1">Repository</label>
              <select
                value={`${owner}/${repo}`}
                onChange={(e) => {
                  const [o, r] = e.target.value.split('/')
                  setOwner(o)
                  setRepo(r)
                }}
                className={inputClasses}
              >
                {repos.map((r) => (
                  <option key={`${r.owner}/${r.name}`} value={`${r.owner}/${r.name}`}>
                    {r.owner}/{r.name}
                  </option>
                ))}
              </select>
            </div>
          ) : (
            <div className="flex gap-2">
              <div className="flex-1">
                <label className="block text-caption-xs text-txt-quaternary mb-1">Owner</label>
                <input
                  value={owner}
                  onChange={(e) => setOwner(e.target.value)}
                  placeholder="owner"
                  className={inputClasses}
                  required
                />
              </div>
              <div className="flex-1">
                <label className="block text-caption-xs text-txt-quaternary mb-1">Repo</label>
                <input
                  value={repo}
                  onChange={(e) => setRepo(e.target.value)}
                  placeholder="repo"
                  className={inputClasses}
                  required
                />
              </div>
            </div>
          )}

          <div>
            <label className="block text-caption-xs text-txt-quaternary mb-1">Prompt</label>
            <textarea
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              placeholder="Describe what Copilot should do..."
              rows={4}
              className={`${inputClasses} resize-y`}
              required
            />
          </div>

          <div>
            <label className="block text-caption-xs text-txt-quaternary mb-1">
              Base Branch <span className="text-txt-quaternary">(optional)</span>
            </label>
            <input
              value={baseBranch}
              onChange={(e) => setBaseBranch(e.target.value)}
              placeholder="main"
              className={inputClasses}
            />
          </div>

          <div className="flex items-center gap-2 pt-1">
            <Button type="submit" variant="primary" disabled={submitting || !prompt.trim()}>
              {submitting ? 'Creating...' : 'Create Task'}
            </Button>
            <Button type="button" variant="ghost" onClick={onBack}>
              Cancel
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Root Component
// ---------------------------------------------------------------------------

type View = { kind: 'list' } | { kind: 'detail'; task: CopilotTask } | { kind: 'create' }

export function CopilotTasks() {
  const { state, dispatch } = useApp()
  const [view, setView] = useState<View>({ kind: 'list' })

  const refresh = useCallback(async () => {
    try {
      const tasks = await api.fetchCopilotTasks()
      dispatch({ type: 'SET_COPILOT_TASKS', tasks })
    } catch {
      // ignore
    }
    setView({ kind: 'list' })
  }, [dispatch])

  // Derive repos list from github state for the create form
  const repos = state.github.repos.map((r) => ({ owner: r.owner, name: r.name }))

  switch (view.kind) {
    case 'detail':
      return (
        <TaskDetail
          task={view.task}
          onBack={() => setView({ kind: 'list' })}
          onRefresh={refresh}
        />
      )
    case 'create':
      return (
        <CreateTask
          repos={repos}
          onBack={() => setView({ kind: 'list' })}
          onCreated={refresh}
        />
      )
    default:
      return (
        <TaskList
          tasks={state.copilotTasks}
          available={state.copilotAvailable}
          onSelect={(t) => setView({ kind: 'detail', task: t })}
          onShowCreate={() => setView({ kind: 'create' })}
        />
      )
  }
}
