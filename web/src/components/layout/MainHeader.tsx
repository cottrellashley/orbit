import type { ReactNode } from 'react'

interface MainHeaderProps {
  title: string
  breadcrumb?: { label: string; onClick: () => void }
  actions?: ReactNode
}

export function MainHeader({ title, breadcrumb, actions }: MainHeaderProps) {
  return (
    <div className="flex items-center justify-between px-5 py-2.5 shrink-0 border-b border-border-subtle">
      <div className="flex items-center gap-1.5 min-w-0">
        {breadcrumb && (
          <>
            <button
              onClick={breadcrumb.onClick}
              className="text-caption text-txt-quaternary hover:text-txt-secondary transition-colors duration-fast cursor-pointer bg-transparent border-none"
            >
              {breadcrumb.label}
            </button>
            <span className="text-caption text-txt-quaternary">/</span>
          </>
        )}
        <span className="text-heading-sm text-txt truncate">{title}</span>
      </div>
      {actions && <div className="flex items-center gap-2 shrink-0">{actions}</div>}
    </div>
  )
}
