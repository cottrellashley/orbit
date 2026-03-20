import { useState, useCallback } from 'react'
import { useApp } from '../state/context'
import { MainHeader } from '../components/layout/MainHeader'
import { Button } from '../components/ui/Button'
import { Table, THead, TBody, TRow, TH, TD } from '../components/ui/Table'
import { EmptyState } from '../components/ui/EmptyState'
import { Modal } from '../components/ui/Modal'
import { formatTime } from '../utils'
import * as api from '../api/client'

export function Environments() {
  const { state, dispatch } = useApp()
  const [modalOpen, setModalOpen] = useState(false)
  const [formName, setFormName] = useState('')
  const [formPath, setFormPath] = useState('')
  const [formDesc, setFormDesc] = useState('')

  const handleDelete = useCallback(
    async (name: string) => {
      if (!confirm(`Delete environment "${name}"?`)) return
      await api.deleteEnvironment(name)
      const environments = await api.fetchEnvironments()
      dispatch({ type: 'SET_ENVIRONMENTS', environments })
    },
    [dispatch],
  )

  const handleSubmit = useCallback(async () => {
    if (!formName.trim() || !formPath.trim()) return
    await api.createEnvironment({
      name: formName.trim(),
      path: formPath.trim(),
      description: formDesc.trim(),
    })
    setModalOpen(false)
    setFormName('')
    setFormPath('')
    setFormDesc('')
    const environments = await api.fetchEnvironments()
    dispatch({ type: 'SET_ENVIRONMENTS', environments })
  }, [formName, formPath, formDesc, dispatch])

  return (
    <div className="flex flex-col flex-1 min-h-0">
      <MainHeader
        title="Projects"
        actions={
          <Button variant="ghost" onClick={() => setModalOpen(true)}>
            + Add
          </Button>
        }
      />
      <div className="flex-1 overflow-y-auto px-4 py-3">
        {state.environments.length === 0 ? (
          <EmptyState
            title="No projects"
            description="Register project directories to manage AI sessions."
            action={
              <Button onClick={() => setModalOpen(true)}>
                + Add Project
              </Button>
            }
          />
        ) : (
          <>
            <div className="text-caption-xs text-txt-quaternary uppercase tracking-[0.06em] px-3 mb-1">
              {state.environments.length} registered
            </div>
            <Table>
              <THead>
                <TRow>
                  <TH>Name</TH>
                  <TH>Path</TH>
                  <TH>Profile</TH>
                  <TH>Created</TH>
                  <TH></TH>
                </TRow>
              </THead>
              <TBody>
                {state.environments.map((env) => (
                  <TRow key={env.name}>
                    <TD>
                      <span className="text-caption text-txt">{env.name}</span>
                    </TD>
                    <TD>
                      <span className="font-mono text-caption-xs text-txt-tertiary truncate block max-w-[240px]">
                        {env.path}
                      </span>
                    </TD>
                    <TD>
                      <span className="text-caption-xs text-txt-tertiary">{env.profile_name || '—'}</span>
                    </TD>
                    <TD>
                      <span className="text-caption-xs text-txt-quaternary">{formatTime(env.created_at)}</span>
                    </TD>
                    <TD>
                      <Button size="sm" variant="danger" onClick={() => handleDelete(env.name)}>
                        Delete
                      </Button>
                    </TD>
                  </TRow>
                ))}
              </TBody>
            </Table>
          </>
        )}
      </div>

      <Modal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        title="Add Project"
        footer={
          <>
            <Button onClick={() => setModalOpen(false)}>Cancel</Button>
            <Button variant="primary" onClick={handleSubmit}>Add</Button>
          </>
        }
      >
        <div className="flex flex-col gap-3">
          <div>
            <label className="block text-caption-xs text-txt-tertiary mb-1">Name</label>
            <input
              type="text"
              value={formName}
              onChange={(e) => setFormName(e.target.value)}
              placeholder="my-project"
              autoFocus
              className="input-base"
            />
          </div>
          <div>
            <label className="block text-caption-xs text-txt-tertiary mb-1">Path</label>
            <input
              type="text"
              value={formPath}
              onChange={(e) => setFormPath(e.target.value)}
              placeholder="/home/user/my-project"
              className="input-base font-mono"
            />
          </div>
          <div>
            <label className="block text-caption-xs text-txt-tertiary mb-1">
              Description <span className="text-txt-quaternary">optional</span>
            </label>
            <textarea
              value={formDesc}
              onChange={(e) => setFormDesc(e.target.value)}
              placeholder="Brief description..."
              rows={2}
              className="input-base resize-none"
            />
          </div>
        </div>
      </Modal>
    </div>
  )
}
