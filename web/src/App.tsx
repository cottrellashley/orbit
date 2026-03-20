import { useEffect, useCallback } from 'react'
import { AppProvider, useApp } from './state/context'
import { Layout } from './components/layout/Layout'
import { usePolling } from './hooks/usePolling'
import { Dashboard } from './pages/Dashboard'
import { Servers } from './pages/Servers'
import { ServerDetail } from './pages/ServerDetail'
import { Sessions } from './pages/Sessions'
import { Chat } from './pages/Chat'
import { TUI } from './pages/TUI'
import { Environments } from './pages/Environments'
import { Doctor } from './pages/Doctor'
import { Repos } from './pages/Repos'
import { Installs } from './pages/Installs'
import { CopilotTasks } from './pages/CopilotTasks'
import { ComingSoon } from './pages/ComingSoon'

function Router() {
  const { state, refreshAll } = useApp()
  usePolling()

  // Initial data load
  useEffect(() => {
    refreshAll()
  }, [refreshAll])

  // Global keyboard shortcuts
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'r') {
        e.preventDefault()
        refreshAll()
      }
    },
    [refreshAll],
  )

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])

  switch (state.page) {
    case 'dashboard':
      return <Dashboard />

    // Control
    case 'sessions':
      return <Sessions />
    case 'chat':
      return <Chat />
    case 'usage':
      return <ComingSoon title="Usage" />
    case 'repos':
      return <Repos />
    case 'projects':
      return <Environments />

    // Agents
    case 'nodes':
      return <Servers />
    case 'node-detail':
      return <ServerDetail />
    case 'agents':
      return <ComingSoon title="Agents" />
    case 'copilot':
      return <CopilotTasks />

    // Settings
    case 'config':
      return <ComingSoon title="Config" />
    case 'installs':
      return <Installs />
    case 'debug':
      return <ComingSoon title="Debug" />
    case 'health':
      return <Doctor />
    case 'logs':
      return <ComingSoon title="Logs" />

    // Hidden (no sidebar entry)
    case 'tui':
      return <TUI />

    default:
      return <Dashboard />
  }
}

export default function App() {
  return (
    <AppProvider>
      <Layout>
        <Router />
      </Layout>
    </AppProvider>
  )
}
