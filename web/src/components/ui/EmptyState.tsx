import type { ReactNode } from 'react'

interface EmptyStateProps {
  icon?: ReactNode
  title: string
  description: string
  action?: ReactNode
}

export function EmptyState({ title, description, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <div className="bg-bg-raised border border-border-subtle rounded-lg px-8 py-8 max-w-[360px]">
        <div className="text-label text-txt-secondary mb-1">{title}</div>
        <div className="text-caption text-txt-quaternary mb-4">{description}</div>
        {action}
      </div>
    </div>
  )
}
