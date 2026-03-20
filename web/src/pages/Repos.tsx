import { useEffect, useCallback, useState } from 'react'
import { useApp } from '../state/context'
import { MainHeader } from '../components/layout/MainHeader'
import { Table, THead, TBody, TRow, TH, TD } from '../components/ui/Table'
import { EmptyState } from '../components/ui/EmptyState'
import { Button } from '../components/ui/Button'
import { StatusBadge } from '../components/ui/Badge'
import { Spinner } from '../components/ui/Spinner'
import { formatTime } from '../utils'
import * as api from '../api/client'

export function Repos() {
  const { state, dispatch } = useApp()
  const [loading, setLoading] = useState(false)

  const refresh = useCallback(async () => {
    setLoading(true)
    try {
      const [auth, repos] = await Promise.all([
        api.fetchGitHubAuth(),
        api.fetchGitHubRepos(),
      ])
      dispatch({ type: 'SET_GITHUB_AUTH', auth })
      dispatch({ type: 'SET_GITHUB_REPOS', repos })
    } catch (e) {
      console.error('Failed to fetch GitHub data:', e)
    } finally {
      setLoading(false)
    }
  }, [dispatch])

  useEffect(() => {
    if (!state.github.auth) {
      refresh()
    }
  }, [state.github.auth, refresh])

  const auth = state.github.auth
  const repos = state.github.repos

  return (
    <div className="flex flex-col flex-1 min-h-0">
      <MainHeader
        title="Repos"
        actions={
          <Button variant="ghost" onClick={refresh}>
            Refresh
          </Button>
        }
      />
      <div className="flex-1 overflow-y-auto px-5 py-4">
        {/* Auth status */}
        <div className="flex items-center gap-2 bg-bg-raised border border-border-subtle rounded-lg px-4 py-2.5 mb-4">
          {loading ? (
            <>
              <Spinner size={12} />
              <span className="text-caption-xs text-txt-quaternary">Loading...</span>
            </>
          ) : auth ? (
            <>
              <StatusBadge status={auth.authenticated ? 'healthy' : 'error'} />
              <span className="text-caption text-txt-secondary">
                {auth.authenticated ? auth.user : 'Not authenticated'}
              </span>
              {auth.token_source && (
                <>
                  <span className="text-caption-xs text-txt-quaternary">/</span>
                  <span className="text-caption-xs text-txt-quaternary">{auth.token_source}</span>
                </>
              )}
            </>
          ) : (
            <span className="text-caption-xs text-txt-quaternary">No auth data</span>
          )}
        </div>

        {/* Repo list */}
        {repos.length === 0 && !loading ? (
          <EmptyState
            title="No repositories"
            description={
              auth?.authenticated
                ? 'No repositories found for this account.'
                : 'Authenticate with GitHub to see your repositories. Set GITHUB_TOKEN or use gh auth login.'
            }
            action={<Button onClick={refresh}>Retry</Button>}
          />
        ) : repos.length > 0 ? (
          <>
            <div className="text-caption-xs text-txt-tertiary uppercase tracking-[0.06em] mb-2">
              {repos.length} repositories
            </div>
            <Table>
              <THead>
                <TRow>
                  <TH>Repository</TH>
                  <TH>Branch</TH>
                  <TH>Visibility</TH>
                  <TH>Updated</TH>
                </TRow>
              </THead>
              <TBody>
                {repos.map((repo) => (
                  <TRow key={repo.full_name}>
                    <TD>
                      <div>
                        <span className="text-caption text-txt">{repo.full_name}</span>
                        {repo.fork && (
                          <span className="text-caption-xs text-txt-quaternary ml-1.5">fork</span>
                        )}
                        {repo.archived && (
                          <span className="text-caption-xs text-txt-quaternary ml-1.5">archived</span>
                        )}
                      </div>
                      {repo.description && (
                        <div className="text-caption-xs text-txt-tertiary mt-0.5 truncate max-w-[400px]">
                          {repo.description}
                        </div>
                      )}
                    </TD>
                    <TD>
                      <span className="font-mono text-caption-xs text-txt-tertiary">{repo.default_branch}</span>
                    </TD>
                    <TD>
                      <span className="text-caption-xs text-txt-tertiary">
                        {repo.private ? 'private' : 'public'}
                      </span>
                    </TD>
                    <TD>
                      <span className="text-caption-xs text-txt-quaternary">{formatTime(repo.updated_at)}</span>
                    </TD>
                  </TRow>
                ))}
              </TBody>
            </Table>
          </>
        ) : null}
      </div>
    </div>
  )
}
