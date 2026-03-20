import { useCallback } from 'react'
import { useApp } from '../state/context'
import { MainHeader } from '../components/layout/MainHeader'
import { Button } from '../components/ui/Button'
import { EmptyState } from '../components/ui/EmptyState'
import { CheckIcon, AlertIcon, XIcon } from '../components/icons'
import * as api from '../api/client'

export function Doctor() {
  const { state, dispatch } = useApp()

  const runDoctor = useCallback(async () => {
    dispatch({ type: 'SET_LOADING', loading: true })
    try {
      const doctor = await api.fetchDoctor()
      dispatch({ type: 'SET_DOCTOR', doctor })
    } catch (e) {
      console.error('Doctor check failed:', e)
    } finally {
      dispatch({ type: 'SET_LOADING', loading: false })
    }
  }, [dispatch])

  const report = state.doctor

  return (
    <div className="flex flex-col flex-1 min-h-0">
      <MainHeader
        title="Health"
        actions={
          <Button variant="ghost" onClick={runDoctor}>
            Run Checks
          </Button>
        }
      />
      <div className="flex-1 overflow-y-auto px-5 py-4">
        {!report ? (
          <EmptyState
            title="No diagnostics yet"
            description="Run system checks to verify your environment is configured correctly."
            action={<Button onClick={runDoctor}>Run Checks</Button>}
          />
        ) : (
          <>
            {/* Summary line */}
            <div className="flex items-center gap-2 bg-bg-raised border border-border-subtle rounded-lg px-4 py-2.5 mb-4">
              <span className={`text-label ${report.ok ? 'text-semantic-green' : 'text-semantic-red'}`}>
                {report.ok ? 'All checks passed' : 'Issues found'}
              </span>
              <span className="text-caption-xs text-txt-quaternary">
                {report.results.filter((r) => r.status === 'pass').length}/{report.results.length} passing
              </span>
            </div>

            {/* Check list */}
            <div className="rounded-lg border border-border-subtle overflow-hidden">
              {report.results.map((check, i) => (
                <div
                  key={i}
                  className={`flex items-start gap-3 px-4 py-3 ${
                    i < report.results.length - 1 ? 'border-b border-border-subtle' : ''
                  }`}
                >
                  <div className="mt-[2px] shrink-0">
                    {check.status === 'pass' ? (
                      <CheckIcon size={13} className="text-semantic-green" />
                    ) : check.status === 'warn' ? (
                      <AlertIcon size={13} className="text-semantic-orange" />
                    ) : (
                      <XIcon size={13} className="text-semantic-red" />
                    )}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="text-caption text-txt">{check.name}</div>
                    <div className="text-caption-xs text-txt-tertiary mt-0.5">{check.message}</div>
                    {check.fix && (
                      <code className="block mt-1 font-mono text-caption-xs text-accent bg-accent-muted px-2 py-1 rounded">
                        {check.fix}
                      </code>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </>
        )}
      </div>
    </div>
  )
}
