import type { ReactNode } from 'react'

type IconColor = 'accent' | 'green' | 'yellow' | 'red' | 'blue'

interface SummaryCardProps {
  icon?: ReactNode
  iconColor?: IconColor
  value: number | string
  label: string
  onClick?: () => void
}

export function SummaryCard({ value, label, onClick }: SummaryCardProps) {
  return (
    <div
      className={`${onClick ? 'cursor-pointer hover:bg-white/[0.02]' : ''} transition-colors duration-fast px-3 py-2 rounded`}
      onClick={onClick}
    >
      <div className="text-label text-txt tabular-nums">{value}</div>
      <div className="text-caption-xs text-txt-quaternary mt-0.5">{label}</div>
    </div>
  )
}

export function SummaryGrid({ children }: { children: ReactNode }) {
  return (
    <div className="flex gap-1 mb-4">{children}</div>
  )
}
