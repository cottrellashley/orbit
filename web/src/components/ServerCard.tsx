import type { Server } from '../state/types'
import { StatusBadge } from './ui/Badge'

interface ServerCardProps {
  server: Server
  onClick: () => void
}

export function ServerCard({ server, onClick }: ServerCardProps) {
  return (
    <div
      className="cursor-pointer bg-bg-raised border border-border-subtle rounded-lg px-4 py-3 hover:border-border-hover hover:bg-bg-hover transition-colors duration-fast"
      onClick={onClick}
    >
      <div className="flex items-center justify-between mb-1.5">
        <span className="font-mono text-caption text-txt tabular-nums">:{server.port}</span>
        <StatusBadge status={server.healthy ? 'healthy' : 'error'} />
      </div>
      <div className="font-mono text-caption-xs text-txt-tertiary truncate mb-1">
        {server.directory}
      </div>
      <div className="flex items-center gap-2 text-caption-xs text-txt-quaternary">
        <span className="tabular-nums">PID {server.pid}</span>
        {server.version && (
          <>
            <span>/</span>
            <span>v{server.version}</span>
          </>
        )}
      </div>
    </div>
  )
}

export function ServerGrid({ children }: { children: React.ReactNode }) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-2">{children}</div>
  )
}
