import { useCallback } from 'react'
import { useApp } from '../state/context'
import { MainHeader } from '../components/layout/MainHeader'
import { Button } from '../components/ui/Button'
import { Table, THead, TBody, TRow, TH, TD } from '../components/ui/Table'
import { StatusBadge } from '../components/ui/Badge'
import { EmptyState } from '../components/ui/EmptyState'
import { formatTime } from '../utils'
import * as api from '../api/client'

export function Sessions() {
  const { state, navigate, dispatch } = useApp()

  const handleAbort = useCallback(
    async (id: string) => {
      if (!confirm('Abort this session?')) return
      await api.abortSession(id)
      const sessions = await api.fetchSessions()
      dispatch({ type: 'SET_SESSIONS', sessions })
    },
    [dispatch],
  )

  const handleDelete = useCallback(
    async (id: string) => {
      if (!confirm('Delete this session?')) return
      await api.deleteSession(id)
      const sessions = await api.fetchSessions()
      dispatch({ type: 'SET_SESSIONS', sessions })
    },
    [dispatch],
  )

  return (
    <div className="flex flex-col flex-1 min-h-0">
      <MainHeader title="Sessions" />
      <div className="flex-1 overflow-y-auto px-4 py-3">
        {state.sessions.length === 0 ? (
          <EmptyState
            title="No active sessions"
            description="Sessions from discovered OpenCode servers will appear here."
          />
        ) : (
          <>
            <div className="text-caption-xs text-txt-quaternary uppercase tracking-[0.06em] px-3 mb-1">
              {state.sessions.length} total
            </div>
            <Table>
              <THead>
                <TRow>
                  <TH>Title</TH>
                  <TH>Environment</TH>
                  <TH>Server</TH>
                  <TH>Status</TH>
                  <TH>Created</TH>
                  <TH></TH>
                </TRow>
              </THead>
              <TBody>
                {state.sessions.map((s) => (
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
                      <span className="text-caption-xs text-txt-tertiary">{s.environment_name || '—'}</span>
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
                    <TD onClick={(e) => e.stopPropagation()}>
                      <div className="flex items-center gap-1">
                        {s.status === 'busy' && (
                          <Button size="sm" variant="danger" onClick={() => handleAbort(s.id)}>
                            Abort
                          </Button>
                        )}
                        <Button size="sm" variant="danger" onClick={() => handleDelete(s.id)}>
                          Delete
                        </Button>
                      </div>
                    </TD>
                  </TRow>
                ))}
              </TBody>
            </Table>
          </>
        )}
      </div>
    </div>
  )
}
