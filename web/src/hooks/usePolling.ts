import { useEffect, useRef } from 'react'
import { useApp } from '../state/context'

const POLL_PAGES = ['dashboard', 'servers', 'sessions']
const POLL_INTERVAL = 10_000

export function usePolling() {
  const { state, refreshServersAndSessions } = useApp()
  const pageRef = useRef(state.page)
  pageRef.current = state.page

  useEffect(() => {
    const id = setInterval(async () => {
      if (POLL_PAGES.includes(pageRef.current)) {
        await refreshServersAndSessions()
      }
    }, POLL_INTERVAL)
    return () => clearInterval(id)
  }, [refreshServersAndSessions])
}
