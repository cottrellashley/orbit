import { useState, useEffect, useCallback } from 'react'
import { useApp } from '../state/context'
import { MainHeader } from '../components/layout/MainHeader'
import { Button } from '../components/ui/Button'
import { Table, THead, TBody, TRow, TH, TD } from '../components/ui/Table'
import { StatusBadge } from '../components/ui/Badge'
import { Spinner } from '../components/ui/Spinner'
import { EmptyState } from '../components/ui/EmptyState'
import { formatTime } from '../utils'
import * as api from '../api/client'
import type { Session, SessionStatusMap } from '../state/types'

export function ServerDetail() {
  const { state, navigate } = useApp()
  const server = state.selectedServer

  const [sessions, setSessions] = useState<Session[]>([])
  const [statuses, setStatuses] = useState<SessionStatusMap>({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const loadSessions = useCallback(async () => {
    if (!server) return
    setLoading(true)
    setError(null)
    try {
      const [rawSessions, statusMap] = await Promise.all([
        api.fetchServerSessions(server.port),
        api.fetchSessionStatus(server.port),
      ])
      const topLevel = (rawSessions || []).filter((s) => !s.parentID)
      setSessions(topLevel)
      setStatuses(statusMap || {})
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load sessions')
    } finally {
      setLoading(false)
    }
  }, [server])

  useEffect(() => {
    loadSessions()
  }, [loadSessions])

  const handleDelete = useCallback(
    async (id: string) => {
      if (!confirm('Delete this session?')) return
      if (!server) return
      await api.deleteSessionDirect(server.port, id)
      loadSessions()
    },
    [server, loadSessions],
  )

  if (!server) {
    return (
      <div className="flex flex-col flex-1 min-h-0">
        <MainHeader title="Server" />
        <div className="flex-1 flex items-center justify-center px-5 py-4">
          <EmptyState
            title="No server selected"
            description="Go back to the servers page and select a server."
            action={<Button onClick={() => navigate('nodes')}>Back to Nodes</Button>}
          />
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col flex-1 min-h-0">
      <MainHeader
        title={`:${server.port}`}
        breadcrumb={{ label: 'Nodes', onClick: () => navigate('nodes') }}
      />
      <div className="flex-1 overflow-y-auto px-5 py-4">
        {/* Server info */}
        <div className="flex items-center gap-3 bg-bg-raised border border-border-subtle rounded-lg px-4 py-2.5 mb-4">
          <StatusBadge status={server.healthy ? 'healthy' : 'error'} />
          <span className="text-caption-xs text-txt-quaternary">/</span>
          <span className="font-mono text-caption-xs text-txt-tertiary">PID {server.pid}</span>
          {server.version && (
            <>
              <span className="text-caption-xs text-txt-quaternary">/</span>
              <span className="text-caption-xs text-txt-tertiary">v{server.version}</span>
            </>
          )}
          {server.directory && (
            <>
              <span className="text-caption-xs text-txt-quaternary">/</span>
              <span className="font-mono text-caption-xs text-txt-tertiary truncate">{server.directory}</span>
            </>
          )}
        </div>

        {/* Sessions */}
        {loading ? (
          <div className="flex items-center gap-2 py-8 text-txt-quaternary justify-center">
            <Spinner size={12} />
            <span className="text-caption-xs">Loading sessions...</span>
          </div>
        ) : error ? (
          <EmptyState title="Error" description={error} />
        ) : sessions.length === 0 ? (
          <EmptyState
            title="No sessions"
            description="This server has no sessions yet."
          />
        ) : (
          <>
            <div className="text-caption-xs text-txt-tertiary uppercase tracking-[0.06em] mb-2">
              Sessions ({sessions.length})
            </div>
            <Table>
              <THead>
                <TRow>
                  <TH>Title</TH>
                  <TH>Status</TH>
                  <TH>Created</TH>
                  <TH></TH>
                </TRow>
              </THead>
              <TBody>
                {sessions.map((s) => {
                  const st = statuses[s.id] || { type: 'unknown' }
                  return (
                    <TRow
                      key={s.id}
                      clickable
                      onClick={() =>
                        navigate('tui', {
                          tuiOpts: {
                            serverURL: `http://127.0.0.1:${server.port}`,
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
                        <StatusBadge status={st.type} />
                      </TD>
                      <TD>
                        <span className="text-caption-xs text-txt-quaternary">{formatTime(s.created_at)}</span>
                      </TD>
                      <TD onClick={(e) => e.stopPropagation()}>
                        <Button size="sm" variant="danger" onClick={() => handleDelete(s.id)}>
                          Delete
                        </Button>
                      </TD>
                    </TRow>
                  )
                })}
              </TBody>
            </Table>
          </>
        )}
      </div>
    </div>
  )
}
