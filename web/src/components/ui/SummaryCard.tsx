import type { ReactNode } from 'react'

interface SummaryCardProps {
  icon?: ReactNode
  iconColor?: 'accent' | 'green' | 'yellow' | 'red' | 'blue'
  value: number | string
  label: string
  onClick?: () => void
}

export function SummaryCard({ value, label, onClick }: SummaryCardProps) {
  return (
    <div
      className={`
        bg-bg-raised border border-border-subtle rounded-lg px-4 py-3
        transition-colors duration-fast
        ${onClick ? 'cursor-pointer hover:border-border-hover hover:bg-bg-hover' : ''}
      `}
      onClick={onClick}
    >
      <div className="text-heading-lg text-txt tabular-nums">{value}</div>
      <div className="text-caption-xs text-txt-tertiary mt-0.5">{label}</div>
    </div>
  )
}

export function SummaryGrid({ children }: { children: ReactNode }) {
  return (
    <div className="grid grid-cols-2 sm:grid-cols-4 gap-2 mb-5">{children}</div>
  )
}
