import { useState } from 'react'
import { useApp } from '../state/context'
import { MainHeader } from '../components/layout/MainHeader'
import * as api from '../api/client'
import type { ToolInfo } from '../state/types'

export function Installs() {
  const { state, dispatch } = useApp()
  const [installing, setInstalling] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const installed = state.installs.filter((t) => t.status === 'installed')
  const notInstalled = state.installs.filter((t) => t.status !== 'installed')

  const handleInstall = async (name: string) => {
    setInstalling(name)
    setError(null)
    try {
      const result = await api.installTool(name)
      if (result.success) {
        const updated: ToolInfo = {
          name: result.name,
          description: state.installs.find((t) => t.name === name)?.description ?? '',
          status: 'installed',
          version: result.version,
        }
        dispatch({ type: 'UPDATE_INSTALL', tool: updated })
      } else {
        setError(`${name}: ${result.error}`)
      }
    } catch (e) {
      setError(`${name}: ${e instanceof Error ? e.message : 'unknown error'}`)
    } finally {
      setInstalling(null)
    }
  }

  return (
    <div className="flex flex-col flex-1 min-h-0">
      <MainHeader title="Installs" />
      <div className="flex-1 overflow-y-auto px-5 py-4">
        {state.installs.length === 0 ? (
          <div className="flex items-center justify-center h-full">
            <div className="bg-bg-raised border border-border-subtle rounded-lg px-8 py-8 text-center">
              <div className="text-label text-txt-secondary">No tools registered</div>
              <div className="text-caption-xs text-txt-quaternary mt-1">
                Install service is not available.
              </div>
            </div>
          </div>
        ) : (
          <div className="space-y-5">
            {error && (
              <div className="text-caption-xs text-semantic-red bg-semantic-red-soft border border-semantic-red/20 rounded-lg px-4 py-2.5">
                {error}
              </div>
            )}

            {installed.length > 0 && (
              <section>
                <div className="text-caption-xs text-txt-tertiary uppercase tracking-wider mb-2">
                  Installed
                </div>
                <div className="rounded-lg border border-border-subtle overflow-hidden">
                  {installed.map((t, i) => (
                    <div
                      key={t.name}
                      className={`flex items-center px-4 py-2.5 text-caption ${
                        i < installed.length - 1 ? 'border-b border-border-subtle' : ''
                      }`}
                    >
                      <span className="text-txt w-[140px] shrink-0">{t.name}</span>
                      <span className="text-txt-tertiary font-mono text-caption-xs w-[100px] shrink-0">
                        {t.version || '-'}
                      </span>
                      <span className="text-txt-quaternary flex-1 truncate">{t.description}</span>
                    </div>
                  ))}
                </div>
              </section>
            )}

            {notInstalled.length > 0 && (
              <section>
                <div className="text-caption-xs text-txt-tertiary uppercase tracking-wider mb-2">
                  Not Installed
                </div>
                <div className="rounded-lg border border-border-subtle overflow-hidden">
                  {notInstalled.map((t, i) => (
                    <div
                      key={t.name}
                      className={`flex items-center px-4 py-2.5 text-caption ${
                        i < notInstalled.length - 1 ? 'border-b border-border-subtle' : ''
                      }`}
                    >
                      <span className="text-txt w-[140px] shrink-0">{t.name}</span>
                      <span className="text-txt-quaternary flex-1 truncate">{t.description}</span>
                      <button
                        onClick={() => handleInstall(t.name)}
                        disabled={installing !== null}
                        className="text-caption-xs text-accent hover:text-accent-hover disabled:text-txt-quaternary disabled:cursor-not-allowed transition-colors shrink-0 ml-3 bg-transparent border-none cursor-pointer"
                      >
                        {installing === t.name ? 'Installing...' : 'Install'}
                      </button>
                    </div>
                  ))}
                </div>
              </section>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
