import { useApp } from '../state/context'
import { MainHeader } from '../components/layout/MainHeader'
import { SummaryGrid, SummaryCard } from '../components/ui/SummaryCard'
import { Table, THead, TBody, TRow, TH, TD } from '../components/ui/Table'
import { StatusBadge } from '../components/ui/Badge'
import { ServerCard, ServerGrid } from '../components/ServerCard'
import { formatTime } from '../utils'

export function Dashboard() {
  const { state, navigate } = useApp()

  const healthStatus = state.doctor
    ? state.doctor.ok
      ? 'Passing'
      : 'Issues'
    : '—'

  const recentSessions = state.sessions.slice(0, 5)
  const displayServers = state.servers.slice(0, 4)

  return (
    <div className="flex flex-col flex-1 min-h-0">
      <MainHeader title="Dashboard" />
      <div className="flex-1 overflow-y-auto px-4 py-3">
        {/* Stats */}
        <SummaryGrid>
          <SummaryCard value={state.servers.length} label="Nodes" onClick={() => navigate('nodes')} />
          <SummaryCard value={state.sessions.length} label="Sessions" onClick={() => navigate('sessions')} />
          <SummaryCard value={state.environments.length} label="Projects" onClick={() => navigate('projects')} />
          <SummaryCard value={healthStatus} label="Health" onClick={() => navigate('health')} />
        </SummaryGrid>

        {/* Active servers */}
        {displayServers.length > 0 && (
          <div className="mb-4">
            <div className="text-caption-xs text-txt-quaternary uppercase tracking-[0.06em] px-3 mb-1">Nodes</div>
            <ServerGrid>
              {displayServers.map((s) => (
                <ServerCard
                  key={s.port}
                  server={s}
                  onClick={() => navigate('node-detail', { server: s })}
                />
              ))}
            </ServerGrid>
          </div>
        )}

        {/* Recent sessions */}
        {recentSessions.length > 0 && (
          <div>
            <div className="flex items-center justify-between px-3 mb-1">
              <span className="text-caption-xs text-txt-quaternary uppercase tracking-[0.06em]">Recent Sessions</span>
              <button
                onClick={() => navigate('sessions')}
                className="text-caption-xs text-txt-quaternary hover:text-txt-tertiary transition-colors duration-fast cursor-pointer bg-transparent border-none"
              >
                View all
              </button>
            </div>
            <Table>
              <THead>
                <TRow>
                  <TH>Title</TH>
                  <TH>Server</TH>
                  <TH>Status</TH>
                  <TH>Created</TH>
                </TRow>
              </THead>
              <TBody>
                {recentSessions.map((s) => (
                  <TRow
                    key={s.id}
                    clickable
                    onClick={() =>
                      navigate('tui', {
                        tuiOpts: {
                          serverURL: `http://127.0.0.1:${s.server_port}`,
                          sessionID: s.id,
                          title: s.title || 'Untitled',
                        },
                      })
                    }
                  >
                    <TD>
                      <span className="text-caption text-txt">{s.title || 'Untitled'}</span>
                      <span className="font-mono text-caption-xs text-txt-quaternary ml-1.5">{s.id.substring(0, 8)}</span>
                    </TD>
                    <TD>
                      <span className="font-mono text-caption-xs text-txt-tertiary tabular-nums">:{s.server_port}</span>
                    </TD>
                    <TD>
                      <StatusBadge status={s.status || 'unknown'} />
                    </TD>
                    <TD>
                      <span className="text-caption-xs text-txt-quaternary">{formatTime(s.created_at)}</span>
                    </TD>
                  </TRow>
                ))}
              </TBody>
            </Table>
          </div>
        )}
      </div>
    </div>
  )
}
