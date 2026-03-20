import { useApp } from '../../state/context'
import type { Page } from '../../state/types'
import {
  OrbitLogo,
  MessageIcon,
  TerminalIcon,
  EyeIcon,
  BarChartIcon,
  GitRepoIcon,
  FolderIcon,
  NetworkIcon,
  CpuIcon,
  ActivityIcon,
  SettingsIcon,
  DownloadIcon,
  BugIcon,
  HeartPulseIcon,
  FileTextIcon,
} from '../icons'

interface NavItem {
  page: Page
  label: string
  icon: React.ReactNode
  badge?: number
}

export function Sidebar() {
  const { state, navigate } = useApp()

  const controlItems: NavItem[] = [
    {
      page: 'sessions',
      label: 'Sessions',
      icon: <MessageIcon size={14} />,
      badge: state.sessions.length,
    },
    { page: 'chat', label: 'Chat', icon: <TerminalIcon size={14} /> },
    { page: 'dashboard', label: 'Overview', icon: <EyeIcon size={14} /> },
    { page: 'usage', label: 'Usage', icon: <BarChartIcon size={14} /> },
    {
      page: 'repos',
      label: 'Repos',
      icon: <GitRepoIcon size={14} />,
      badge: state.github.repos.length,
    },
    {
      page: 'projects',
      label: 'Projects',
      icon: <FolderIcon size={14} />,
      badge: state.environments.length,
    },
  ]

  const agentsItems: NavItem[] = [
    {
      page: 'nodes',
      label: 'Nodes',
      icon: <NetworkIcon size={14} />,
      badge: state.servers.length,
    },
    { page: 'agents', label: 'Agents', icon: <CpuIcon size={14} /> },
    {
      page: 'copilot',
      label: 'Copilot',
      icon: <ActivityIcon size={14} />,
      badge: state.copilotTasks.filter((t) => t.status === 'running').length,
    },
  ]

  const settingsItems: NavItem[] = [
    { page: 'config', label: 'Config', icon: <SettingsIcon size={14} /> },
    { page: 'installs', label: 'Installs', icon: <DownloadIcon size={14} /> },
    { page: 'debug', label: 'Debug', icon: <BugIcon size={14} /> },
    { page: 'health', label: 'Health', icon: <HeartPulseIcon size={14} /> },
    { page: 'logs', label: 'Logs', icon: <FileTextIcon size={14} /> },
  ]

  return (
    <div className="w-[200px] h-full flex flex-col shrink-0 bg-bg-sidebar border-r border-border">
      {/* Logo */}
      <div className="px-4 py-3">
        <div className="flex items-center gap-2">
          <div className="text-accent">
            <OrbitLogo size={20} />
          </div>
          <span className="text-label text-txt font-semibold">Orbit</span>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto py-1 px-2">
        <NavGroup label="Control" items={controlItems} currentPage={state.page} onNavigate={navigate} />
        <NavGroup label="Agents" items={agentsItems} currentPage={state.page} onNavigate={navigate} />
        <NavGroup label="Settings" items={settingsItems} currentPage={state.page} onNavigate={navigate} />
      </nav>

      {/* Footer */}
      <div className="px-4 py-2 border-t border-border-subtle">
        <div className="text-caption-xs text-txt-quaternary">v0.1.0</div>
      </div>
    </div>
  )
}

function NavGroup({
  label,
  items,
  currentPage,
  onNavigate,
}: {
  label: string
  items: NavItem[]
  currentPage: Page
  onNavigate: (page: Page) => void
}) {
  return (
    <div className="mb-1">
      <div className="px-2 py-1.5 text-caption-xs uppercase tracking-[0.08em] text-txt-quaternary">
        {label}
      </div>
      {items.map((item) => {
        const isActive = currentPage === item.page
        return (
          <button
            key={item.page}
            onClick={() => onNavigate(item.page)}
            className={`
              w-full flex items-center gap-2 px-2 py-[5px] rounded
              text-caption font-medium transition-colors duration-fast
              cursor-pointer border-none
              ${isActive
                ? 'bg-accent-muted text-txt'
                : 'bg-transparent text-txt-tertiary hover:text-txt-secondary hover:bg-white/[0.03]'
              }
            `}
          >
            <span className={`shrink-0 ${isActive ? 'text-accent' : ''}`}>{item.icon}</span>
            <span className="flex-1 text-left truncate">{item.label}</span>
            {item.badge !== undefined && item.badge > 0 && (
              <span className="text-caption-xs tabular-nums text-txt-quaternary">
                {item.badge}
              </span>
            )}
          </button>
        )
      })}
    </div>
  )
}
