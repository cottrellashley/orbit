import { useRef, useState, useCallback } from 'react'
import { useOpenCodeTUI } from '../hooks/useOpenCodeTUI'
import type { TUIOptions } from '../state/types'
import { Spinner } from './ui/Spinner'
import { Button } from './ui/Button'
import { RefreshIcon } from './icons'

interface TerminalViewProps {
  opts: TUIOptions
  onReconnect?: () => void
}

export function TerminalView({ opts, onReconnect }: TerminalViewProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const [status, setStatus] = useState<'connecting' | 'connected' | 'disconnected' | 'error'>('connecting')
  const [statusMessage, setStatusMessage] = useState<string>('')
  const optsRef = useRef(opts)
  optsRef.current = opts

  const handleStatusChange = useCallback(
    (newStatus: 'connecting' | 'connected' | 'disconnected' | 'error', message?: string) => {
      setStatus(newStatus)
      setStatusMessage(message ?? '')
    },
    [],
  )

  const { open, destroy } = useOpenCodeTUI(handleStatusChange)

  const openRef = useRef(open)
  openRef.current = open
  const destroyRef = useRef(destroy)
  destroyRef.current = destroy

  // Open terminal on mount
  const mountedRef = useRef(false)
  const containerCallbackRef = useCallback(
    (el: HTMLDivElement | null) => {
      (containerRef as React.MutableRefObject<HTMLDivElement | null>).current = el
      if (el && !mountedRef.current) {
        mountedRef.current = true
        openRef.current(el, optsRef.current)
      }
    },
    [], // eslint-disable-line react-hooks/exhaustive-deps
  )

  const handleReconnect = useCallback(() => {
    destroyRef.current()
    mountedRef.current = false
    setStatus('connecting')
    setStatusMessage('')
    if (onReconnect) {
      onReconnect()
    } else if (containerRef.current) {
      openRef.current(containerRef.current, optsRef.current)
    }
  }, [onReconnect])

  return (
    <div className="flex flex-col flex-1 min-h-0 relative overflow-hidden">
      {/* Connecting overlay */}
      {status === 'connecting' && (
        <div className="absolute inset-0 z-10 flex flex-col items-center justify-center bg-[#0a0a0a]/90">
          <Spinner size={16} className="mb-2" />
          <div className="text-caption text-txt-tertiary">Connecting...</div>
        </div>
      )}

      {/* Disconnected/error overlay */}
      {(status === 'disconnected' || status === 'error') && (
        <div className="absolute inset-0 z-10 flex flex-col items-center justify-center bg-[#0a0a0a]/90">
          <div className="text-label-sm text-txt-secondary mb-1">
            {status === 'error' ? 'Connection error' : 'Session ended'}
          </div>
          <div className="text-caption text-txt-quaternary mb-3 max-w-[260px] text-center">
            {statusMessage || 'The terminal session has ended.'}
          </div>
          <Button variant="ghost" onClick={handleReconnect}>
            <RefreshIcon size={12} />
            Reconnect
          </Button>
        </div>
      )}

      {/* xterm.js container — flush, no padding */}
      <div
        ref={containerCallbackRef}
        className="flex-1 min-h-0"
      />
    </div>
  )
}
