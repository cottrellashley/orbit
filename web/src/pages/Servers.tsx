import { useApp } from '../state/context'
import { MainHeader } from '../components/layout/MainHeader'
import { EmptyState } from '../components/ui/EmptyState'
import { ServerCard, ServerGrid } from '../components/ServerCard'

export function Servers() {
  const { state, navigate } = useApp()

  return (
    <div className="flex flex-col flex-1 min-h-0">
      <MainHeader title="Nodes" />
      <div className="flex-1 overflow-y-auto px-5 py-4">
        {state.servers.length === 0 ? (
          <EmptyState
            title="No nodes found"
            description="Start an OpenCode server in a project directory, and it will be discovered automatically."
          />
        ) : (
          <>
            <div className="text-caption-xs text-txt-tertiary uppercase tracking-[0.06em] mb-2">
              {state.servers.length} discovered
            </div>
            <ServerGrid>
              {state.servers.map((s) => (
                <ServerCard
                  key={s.port}
                  server={s}
                  onClick={() => navigate('node-detail', { server: s })}
                />
              ))}
            </ServerGrid>
          </>
        )}
      </div>
    </div>
  )
}
