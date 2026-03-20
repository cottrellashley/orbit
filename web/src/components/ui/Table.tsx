import type { ReactNode, TdHTMLAttributes, ThHTMLAttributes } from 'react'

export function Table({ children }: { children: ReactNode }) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full border-collapse">{children}</table>
    </div>
  )
}

export function THead({ children }: { children: ReactNode }) {
  return <thead>{children}</thead>
}

export function TBody({ children }: { children: ReactNode }) {
  return <tbody>{children}</tbody>
}

interface TRowProps {
  children: ReactNode
  clickable?: boolean
  onClick?: () => void
  className?: string
}

export function TRow({ children, clickable, onClick, className = '' }: TRowProps) {
  return (
    <tr
      className={`
        border-b border-border/50 last:border-b-0 transition-colors duration-fast
        ${clickable ? 'cursor-pointer hover:bg-white/[0.02]' : ''}
        ${className}
      `}
      onClick={onClick}
    >
      {children}
    </tr>
  )
}

export function TH({ children, className = '', ...props }: ThHTMLAttributes<HTMLTableCellElement> & { children?: ReactNode }) {
  return (
    <th
      className={`
        text-left text-caption-xs uppercase tracking-[0.06em]
        text-txt-quaternary font-normal
        px-3 py-1.5
        border-b border-border/50
        ${className}
      `}
      {...props}
    >
      {children}
    </th>
  )
}

export function TD({ children, className = '', ...props }: TdHTMLAttributes<HTMLTableCellElement> & { children?: ReactNode }) {
  return (
    <td className={`px-3 py-2 text-caption ${className}`} {...props}>
      {children}
    </td>
  )
}
