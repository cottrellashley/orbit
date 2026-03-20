import { useRef, useCallback, useEffect } from 'react'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebglAddon } from '@xterm/addon-webgl'
import { spawnTerminal, killTerminal, fetchManagedServerURL } from '../api/client'
import type { TUIOptions } from '../state/types'

export interface UseOpenCodeTUIReturn {
  open: (containerEl: HTMLDivElement, opts: TUIOptions) => Promise<void>
  destroy: () => void
  isActive: () => boolean
}

const XTERM_THEME = {
  background: '#0f1117',
  foreground: '#e8ecf4',
  cursor: '#4C8BF5',
  selectionBackground: 'rgba(76,139,245,0.3)',
  black: '#1c1f2b',
  red: '#F87171',
  green: '#34D399',
  yellow: '#FBBF24',
  blue: '#4C8BF5',
  magenta: '#a882ff',
  cyan: '#22D3EE',
  white: '#e8ecf4',
  brightBlack: '#5f6880',
  brightRed: '#FCA5A5',
  brightGreen: '#6EE7B7',
  brightYellow: '#FDE68A',
  brightBlue: '#93C5FD',
  brightMagenta: '#C4B5FD',
  brightCyan: '#67E8F9',
  brightWhite: '#F8FAFC',
}

export function useOpenCodeTUI(
  onStatusChange?: (status: 'connecting' | 'connected' | 'disconnected' | 'error', message?: string) => void,
): UseOpenCodeTUIReturn {
  const terminalIdRef = useRef<string | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const termRef = useRef<Terminal | null>(null)
  const fitAddonRef = useRef<FitAddon | null>(null)
  const webglAddonRef = useRef<WebglAddon | null>(null)
  const resizeObserverRef = useRef<ResizeObserver | null>(null)
  const destroyedRef = useRef(false)

  const destroy = useCallback(() => {
    destroyedRef.current = true

    if (resizeObserverRef.current) {
      resizeObserverRef.current.disconnect()
      resizeObserverRef.current = null
    }
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
    if (webglAddonRef.current) {
      webglAddonRef.current.dispose()
      webglAddonRef.current = null
    }
    if (fitAddonRef.current) {
      fitAddonRef.current = null
    }
    if (termRef.current) {
      termRef.current.dispose()
      termRef.current = null
    }
    if (terminalIdRef.current) {
      killTerminal(terminalIdRef.current)
      terminalIdRef.current = null
    }
  }, [])

  const open = useCallback(
    async (containerEl: HTMLDivElement, opts: TUIOptions) => {
      destroyedRef.current = false
      onStatusChange?.('connecting')

      // Create xterm instance
      const term = new Terminal({
        cursorBlink: true,
        fontSize: 14,
        fontFamily:
          "'JetBrains Mono', 'Cascadia Code', 'Fira Code', 'SF Mono', Consolas, monospace",
        allowProposedApi: true,
        theme: XTERM_THEME,
      })
      const fitAddon = new FitAddon()
      term.loadAddon(fitAddon)
      termRef.current = term
      fitAddonRef.current = fitAddon

      // Mount
      containerEl.innerHTML = ''
      term.open(containerEl)

      // Try WebGL addon
      try {
        const webgl = new WebglAddon()
        term.loadAddon(webgl)
        webglAddonRef.current = webgl
      } catch {
        // Canvas fallback
      }

      fitAddon.fit()

      // Resolve server URL
      let serverURL = opts.serverURL
      if (!serverURL) {
        try {
          serverURL = await fetchManagedServerURL()
        } catch (e) {
          if (destroyedRef.current) return
          onStatusChange?.('error', `Failed to discover server: ${e}`)
          return
        }
      }
      if (destroyedRef.current) return

      // Build command args
      const args: string[] = ['attach']
      if (opts.sessionID) {
        args.push('--session', opts.sessionID, '--continue')
      }
      args.push(serverURL)

      // Spawn PTY
      let termId: string
      try {
        const res = await spawnTerminal({
          command: 'opencode',
          args,
          cols: term.cols,
          rows: term.rows,
        })
        termId = res.id
      } catch (e) {
        if (destroyedRef.current) return
        onStatusChange?.('error', `Failed to spawn terminal: ${e}`)
        return
      }
      if (destroyedRef.current) return
      terminalIdRef.current = termId

      // WebSocket
      const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const ws = new WebSocket(`${proto}//${window.location.host}/api/terminal/${termId}/ws`)
      ws.binaryType = 'arraybuffer'
      wsRef.current = ws

      const encoder = new TextEncoder()

      ws.onopen = () => {
        if (destroyedRef.current) return
        onStatusChange?.('connected')
      }

      ws.onmessage = (ev) => {
        if (destroyedRef.current) return
        if (ev.data instanceof ArrayBuffer) {
          term.write(new Uint8Array(ev.data))
        }
      }

      ws.onclose = () => {
        if (destroyedRef.current) return
        onStatusChange?.('disconnected', 'Terminal session ended.')
      }

      ws.onerror = () => {
        if (destroyedRef.current) return
        onStatusChange?.('error', 'WebSocket connection error.')
      }

      // Input from user -> PTY
      term.onData((data) => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(encoder.encode(data))
        }
      })

      // Resize
      term.onResize(({ cols, rows }) => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({ type: 'resize', cols, rows }))
        }
      })

      // ResizeObserver for container
      const ro = new ResizeObserver(() => {
        if (!destroyedRef.current && fitAddonRef.current) {
          try {
            fitAddonRef.current.fit()
          } catch {
            // ignore
          }
        }
      })
      ro.observe(containerEl)
      resizeObserverRef.current = ro
    },
    [onStatusChange],
  )

  const isActive = useCallback(() => {
    return terminalIdRef.current !== null && !destroyedRef.current
  }, [])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      destroy()
    }
  }, [destroy])

  return { open, destroy, isActive }
}
