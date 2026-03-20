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
      <div className="text-caption text-txt-tertiary mb-1">{title}</div>
      <div className="text-caption text-txt-quaternary max-w-[300px] mb-4">{description}</div>
      {action}
    </div>
  )
}
