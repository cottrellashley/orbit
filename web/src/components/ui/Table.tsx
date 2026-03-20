import type { ReactNode, TdHTMLAttributes, ThHTMLAttributes } from 'react'

export function Table({ children }: { children: ReactNode }) {
  return (
    <div className="overflow-x-auto rounded-lg border border-border-subtle">
      <table className="w-full border-collapse">{children}</table>
    </div>
  )
}

export function THead({ children }: { children: ReactNode }) {
  return <thead className="bg-bg-raised">{children}</thead>
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
        border-b border-border-subtle last:border-b-0 transition-colors duration-fast
        ${clickable ? 'cursor-pointer hover:bg-bg-hover' : ''}
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
        text-txt-tertiary font-medium
        px-3 py-2
        border-b border-border
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
    <td className={`px-3 py-2.5 text-caption ${className}`} {...props}>
      {children}
    </td>
  )
}
