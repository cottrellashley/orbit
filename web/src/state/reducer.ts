import type {
  Server,
  Session,
  Environment,
  DoctorReport,
  GitHubAuthStatus,
  GitHubRepo,
  ToolInfo,
  CopilotTask,
  Page,
  TUIOptions,
} from './types'

export interface AppState {
  page: Page
  servers: Server[]
  sessions: Session[]
  environments: Environment[]
  doctor: DoctorReport | null
  github: {
    auth: GitHubAuthStatus | null
    repos: GitHubRepo[]
  }
  installs: ToolInfo[]
  copilotAvailable: boolean
  copilotTasks: CopilotTask[]
  selectedServer: Server | null
  tuiOpts: TUIOptions | null
  loading: boolean
}

export const initialState: AppState = {
  page: 'dashboard',
  servers: [],
  sessions: [],
  environments: [],
  doctor: null,
  github: { auth: null, repos: [] },
  installs: [],
  copilotAvailable: false,
  copilotTasks: [],
  selectedServer: null,
  tuiOpts: null,
  loading: false,
}

export type Action =
  | { type: 'SET_PAGE'; page: Page; server?: Server | null; tuiOpts?: TUIOptions | null }
  | { type: 'SET_SERVERS'; servers: Server[] }
  | { type: 'SET_SESSIONS'; sessions: Session[] }
  | { type: 'SET_ENVIRONMENTS'; environments: Environment[] }
  | { type: 'SET_DOCTOR'; doctor: DoctorReport }
  | { type: 'SET_GITHUB_AUTH'; auth: GitHubAuthStatus }
  | { type: 'SET_GITHUB_REPOS'; repos: GitHubRepo[] }
  | { type: 'SET_INSTALLS'; installs: ToolInfo[] }
  | { type: 'UPDATE_INSTALL'; tool: ToolInfo }
  | { type: 'SET_COPILOT_AVAILABLE'; available: boolean }
  | { type: 'SET_COPILOT_TASKS'; tasks: CopilotTask[] }
  | { type: 'SET_LOADING'; loading: boolean }

export function reducer(state: AppState, action: Action): AppState {
  switch (action.type) {
    case 'SET_PAGE':
      return {
        ...state,
        page: action.page,
        selectedServer: action.server ?? null,
        tuiOpts: action.tuiOpts ?? null,
      }
    case 'SET_SERVERS':
      return { ...state, servers: action.servers }
    case 'SET_SESSIONS':
      return { ...state, sessions: action.sessions }
    case 'SET_ENVIRONMENTS':
      return { ...state, environments: action.environments }
    case 'SET_DOCTOR':
      return { ...state, doctor: action.doctor }
    case 'SET_GITHUB_AUTH':
      return { ...state, github: { ...state.github, auth: action.auth } }
    case 'SET_GITHUB_REPOS':
      return { ...state, github: { ...state.github, repos: action.repos } }
    case 'SET_INSTALLS':
      return { ...state, installs: action.installs }
    case 'UPDATE_INSTALL':
      return {
        ...state,
        installs: state.installs.map((t) =>
          t.name === action.tool.name ? action.tool : t,
        ),
      }
    case 'SET_COPILOT_AVAILABLE':
      return { ...state, copilotAvailable: action.available }
    case 'SET_COPILOT_TASKS':
      return { ...state, copilotTasks: action.tasks }
    case 'SET_LOADING':
      return { ...state, loading: action.loading }
    default:
      return state
  }
}
