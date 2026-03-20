import type {
  Server,
  Session,
  Environment,
  DoctorReport,
  SessionStatusMap,
  GitHubAuthStatus,
  GitHubRepo,
  ToolInfo,
  InstallResult,
  CopilotTask,
  CopilotAvailable,
} from '../state/types'

async function api<T>(path: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    headers: { 'Content-Type': 'application/json' },
    ...opts,
  })
  if (!res.ok) {
    const body = await res.text().catch(() => '')
    throw new Error(body || `HTTP ${res.status}`)
  }
  return res.json()
}

async function proxyGet<T>(port: number, path: string): Promise<T> {
  return api<T>(`/api/proxy/${port}${path}`)
}

export async function fetchServers(): Promise<Server[]> {
  return api<Server[]>('/api/servers')
}

export async function fetchSessions(): Promise<Session[]> {
  return api<Session[]>('/api/sessions')
}

export async function fetchEnvironments(): Promise<Environment[]> {
  return api<Environment[]>('/api/environments')
}

export async function fetchDoctor(): Promise<DoctorReport> {
  return api<DoctorReport>('/api/doctor')
}

export async function fetchSessionStatus(port: number): Promise<SessionStatusMap> {
  return proxyGet<SessionStatusMap>(port, '/session/status')
}

export async function fetchServerSessions(port: number): Promise<Session[]> {
  return proxyGet<Session[]>(port, '/session')
}

export async function abortSession(id: string): Promise<void> {
  await api<void>(`/api/sessions/${id}/abort`, { method: 'POST' })
}

export async function deleteSession(id: string): Promise<void> {
  await api<void>(`/api/sessions/${id}`, { method: 'DELETE' })
}

export async function deleteSessionDirect(port: number, id: string): Promise<void> {
  await fetch(`/api/proxy/${port}/session/${id}`, { method: 'DELETE' })
}

export async function deleteEnvironment(name: string): Promise<void> {
  await api<void>(`/api/environments/${name}`, { method: 'DELETE' })
}

export async function createEnvironment(data: {
  name: string
  path: string
  description: string
}): Promise<void> {
  await api<void>('/api/environments', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function fetchManagedServerURL(): Promise<string> {
  const res = await api<{ url: string }>('/api/managed-server')
  return res.url
}

export interface SpawnTerminalOpts {
  command: string
  args: string[]
  cols: number
  rows: number
}

export async function spawnTerminal(opts: SpawnTerminalOpts): Promise<{ id: string }> {
  return api<{ id: string }>('/api/terminal', {
    method: 'POST',
    body: JSON.stringify(opts),
  })
}

export async function killTerminal(id: string): Promise<void> {
  await fetch(`/api/terminal/${id}`, { method: 'DELETE' }).catch(() => {})
}

// GitHub

export async function fetchGitHubAuth(): Promise<GitHubAuthStatus> {
  return api<GitHubAuthStatus>('/api/github/status')
}

export async function fetchGitHubRepos(): Promise<GitHubRepo[]> {
  return api<GitHubRepo[]>('/api/github/repos')
}

// Installs

export async function fetchInstalls(): Promise<ToolInfo[]> {
  return api<ToolInfo[]>('/api/installs')
}

export async function installTool(name: string): Promise<InstallResult> {
  return api<InstallResult>(`/api/installs/${name}`, { method: 'POST' })
}

// Copilot

export async function fetchCopilotAvailable(): Promise<CopilotAvailable> {
  return api<CopilotAvailable>('/api/copilot/available')
}

export async function fetchCopilotTasks(owner?: string, repo?: string): Promise<CopilotTask[]> {
  const params = new URLSearchParams()
  if (owner) params.set('owner', owner)
  if (repo) params.set('repo', repo)
  const qs = params.toString()
  return api<CopilotTask[]>(`/api/copilot/tasks${qs ? `?${qs}` : ''}`)
}

export async function fetchCopilotTask(id: string): Promise<CopilotTask> {
  return api<CopilotTask>(`/api/copilot/tasks/${id}`)
}

export async function createCopilotTask(data: {
  owner: string
  repo: string
  prompt: string
  baseBranch?: string
  customInstructions?: string
}): Promise<CopilotTask> {
  return api<CopilotTask>('/api/copilot/tasks', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function stopCopilotTask(id: string): Promise<void> {
  await api<void>(`/api/copilot/tasks/${id}`, { method: 'DELETE' })
}

export async function fetchCopilotTaskLogs(id: string): Promise<string> {
  const res = await fetch(`/api/copilot/tasks/${id}/logs`)
  if (!res.ok) {
    const body = await res.text().catch(() => '')
    throw new Error(body || `HTTP ${res.status}`)
  }
  return res.text()
}
