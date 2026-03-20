// Domain types matching the Go backend DTOs

export interface Server {
  port: number
  pid: number
  healthy: boolean
  directory: string
  hostname: string
  version: string
}

export interface Session {
  id: string
  title: string
  status: string
  server_port: number
  environment_name: string
  created_at: string
  parentID?: string
}

export interface Environment {
  name: string
  path: string
  description?: string
  profile_name: string
  created_at: string
}

export interface DoctorCheck {
  name: string
  status: 'pass' | 'warn' | 'fail'
  message: string
  fix?: string
}

export interface DoctorReport {
  ok: boolean
  results: DoctorCheck[]
}

export interface SessionStatus {
  type: string
}

export type SessionStatusMap = Record<string, SessionStatus>

export interface GitHubAuthStatus {
  authenticated: boolean
  user?: string
  token_source?: string
  scopes?: string[]
}

export interface GitHubRepo {
  owner: string
  name: string
  full_name: string
  description?: string
  html_url: string
  clone_url: string
  ssh_url: string
  default_branch: string
  private: boolean
  fork: boolean
  archived: boolean
  updated_at: string
}

export interface GitHubIssue {
  number: number
  title: string
  state: string
  html_url: string
  user: string
  labels: string[]
  is_pull_request: boolean
  created_at: string
  updated_at: string
}

export type Page =
  | 'dashboard'
  | 'sessions'
  | 'chat'
  | 'usage'
  | 'repos'
  | 'projects'
  | 'nodes'
  | 'node-detail'
  | 'agents'
  | 'copilot'
  | 'config'
  | 'installs'
  | 'debug'
  | 'health'
  | 'logs'
  | 'tui'

export interface TUIOptions {
  serverURL?: string
  sessionID?: string
  title?: string
}

export interface ToolInfo {
  name: string
  description: string
  status: 'installed' | 'not_installed' | 'unknown'
  version?: string
}

export interface InstallResult {
  name: string
  success: boolean
  version?: string
  error?: string
}

export interface CopilotTask {
  id: string
  owner: string
  repo: string
  prNumber: number
  title: string
  prompt?: string
  status: 'running' | 'completed' | 'stopped' | 'failed' | 'unknown'
  htmlUrl: string
  branch?: string
  draft: boolean
  createdAt: string
  updatedAt: string
}

export interface CopilotAvailable {
  available: boolean
}
