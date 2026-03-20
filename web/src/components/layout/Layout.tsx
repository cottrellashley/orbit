import type { ReactNode } from 'react'
import { Sidebar } from './Sidebar'
import { LoadingBar } from '../ui/LoadingBar'
import { useApp } from '../../state/context'

export function Layout({ children }: { children: ReactNode }) {
  const { state } = useApp()

  return (
    <div className="flex h-full w-full bg-bg">
      <Sidebar />
      <div className="flex-1 flex flex-col min-w-0 relative bg-bg-content">
        <LoadingBar visible={state.loading} />
        <div className="flex-1 flex flex-col min-h-0 animate-page-in">
          {children}
        </div>
      </div>
    </div>
  )
}
