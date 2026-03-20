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
      <div className="flex-1 overflow-y-auto p-4">
        {state.installs.length === 0 ? (
          <div className="flex items-center justify-center h-full">
            <div className="text-center">
              <div className="text-caption text-txt-tertiary">No tools registered</div>
              <div className="text-caption-xs text-txt-quaternary mt-1">
                Install service is not available.
              </div>
            </div>
          </div>
        ) : (
          <div className="space-y-6">
            {error && (
              <div className="text-caption-xs text-red-400 bg-red-400/10 px-3 py-2">
                {error}
              </div>
            )}

            {installed.length > 0 && (
              <section>
                <div className="text-caption-xs text-txt-quaternary uppercase tracking-wider mb-2">
                  Installed
                </div>
                <table className="w-full text-caption">
                  <thead>
                    <tr className="text-left text-caption-xs text-txt-quaternary">
                      <th className="pb-1 font-normal">Name</th>
                      <th className="pb-1 font-normal">Version</th>
                      <th className="pb-1 font-normal">Description</th>
                    </tr>
                  </thead>
                  <tbody>
                    {installed.map((t) => (
                      <tr key={t.name} className="text-txt-secondary">
                        <td className="py-1 pr-4 text-txt-primary">{t.name}</td>
                        <td className="py-1 pr-4 text-txt-tertiary font-mono text-caption-xs">
                          {t.version || '-'}
                        </td>
                        <td className="py-1 text-txt-quaternary">{t.description}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </section>
            )}

            {notInstalled.length > 0 && (
              <section>
                <div className="text-caption-xs text-txt-quaternary uppercase tracking-wider mb-2">
                  Not Installed
                </div>
                <table className="w-full text-caption">
                  <thead>
                    <tr className="text-left text-caption-xs text-txt-quaternary">
                      <th className="pb-1 font-normal">Name</th>
                      <th className="pb-1 font-normal">Description</th>
                      <th className="pb-1 font-normal text-right">Action</th>
                    </tr>
                  </thead>
                  <tbody>
                    {notInstalled.map((t) => (
                      <tr key={t.name} className="text-txt-secondary">
                        <td className="py-1 pr-4 text-txt-primary">{t.name}</td>
                        <td className="py-1 pr-4 text-txt-quaternary">{t.description}</td>
                        <td className="py-1 text-right">
                          <button
                            onClick={() => handleInstall(t.name)}
                            disabled={installing !== null}
                            className="text-caption-xs text-accent hover:text-accent/80 disabled:text-txt-quaternary disabled:cursor-not-allowed transition-colors"
                          >
                            {installing === t.name ? 'Installing...' : 'Install'}
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </section>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
