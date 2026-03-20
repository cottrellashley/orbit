import {
  createContext,
  useContext,
  useReducer,
  useCallback,
  type ReactNode,
  type Dispatch,
} from 'react'
import { reducer, initialState, type AppState, type Action } from './reducer'
import type { Page, Server, TUIOptions } from './types'
import * as api from '../api/client'

interface AppContextValue {
  state: AppState
  dispatch: Dispatch<Action>
  navigate: (page: Page, opts?: { server?: Server; tuiOpts?: TUIOptions }) => void
  refreshAll: () => Promise<void>
  refreshServersAndSessions: () => Promise<void>
}

const AppContext = createContext<AppContextValue | null>(null)

export function AppProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(reducer, initialState)

  const navigate = useCallback(
    (page: Page, opts?: { server?: Server; tuiOpts?: TUIOptions }) => {
      dispatch({
        type: 'SET_PAGE',
        page,
        server: opts?.server,
        tuiOpts: opts?.tuiOpts,
      })
    },
    [dispatch],
  )

  const refreshAll = useCallback(async () => {
    dispatch({ type: 'SET_LOADING', loading: true })
    try {
      const [servers, sessions, environments] = await Promise.all([
        api.fetchServers(),
        api.fetchSessions(),
        api.fetchEnvironments(),
      ])
      dispatch({ type: 'SET_SERVERS', servers })
      dispatch({ type: 'SET_SESSIONS', sessions })
      dispatch({ type: 'SET_ENVIRONMENTS', environments })

      // GitHub data (non-blocking — don't fail the whole refresh)
      api.fetchGitHubAuth()
        .then((auth) => dispatch({ type: 'SET_GITHUB_AUTH', auth }))
        .catch(() => {})
      api.fetchGitHubRepos()
        .then((repos) => dispatch({ type: 'SET_GITHUB_REPOS', repos }))
        .catch(() => {})

      // Install data (non-blocking)
      api.fetchInstalls()
        .then((installs) => dispatch({ type: 'SET_INSTALLS', installs }))
        .catch(() => {})

      // Copilot data (non-blocking)
      api.fetchCopilotAvailable()
        .then((r) => dispatch({ type: 'SET_COPILOT_AVAILABLE', available: r.available }))
        .catch(() => {})
      api.fetchCopilotTasks()
        .then((tasks) => dispatch({ type: 'SET_COPILOT_TASKS', tasks }))
        .catch(() => {})
    } catch (e) {
      console.error('Failed to refresh:', e)
    } finally {
      dispatch({ type: 'SET_LOADING', loading: false })
    }
  }, [dispatch])

  const refreshServersAndSessions = useCallback(async () => {
    try {
      const [servers, sessions] = await Promise.all([
        api.fetchServers(),
        api.fetchSessions(),
      ])
      dispatch({ type: 'SET_SERVERS', servers })
      dispatch({ type: 'SET_SESSIONS', sessions })
    } catch (e) {
      console.error('Failed to refresh:', e)
    }
  }, [dispatch])

  return (
    <AppContext.Provider
      value={{ state, dispatch, navigate, refreshAll, refreshServersAndSessions }}
    >
      {children}
    </AppContext.Provider>
  )
}

export function useApp(): AppContextValue {
  const ctx = useContext(AppContext)
  if (!ctx) throw new Error('useApp must be used within AppProvider')
  return ctx
}
