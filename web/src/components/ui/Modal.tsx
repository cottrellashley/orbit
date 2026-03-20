import { useEffect, useCallback, type ReactNode } from 'react'
import { XIcon } from '../icons'

interface ModalProps {
  open: boolean
  onClose: () => void
  title: string
  children: ReactNode
  footer?: ReactNode
}

export function Modal({ open, onClose, title, children, footer }: ModalProps) {
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    },
    [onClose],
  )

  useEffect(() => {
    if (open) {
      document.addEventListener('keydown', handleKeyDown)
      return () => document.removeEventListener('keydown', handleKeyDown)
    }
  }, [open, handleKeyDown])

  if (!open) return null

  return (
    <div
      className="fixed inset-0 z-[1000] flex items-center justify-center animate-fade-in"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose()
      }}
      style={{ background: 'rgba(0, 0, 0, 0.70)', backdropFilter: 'blur(4px)' }}
    >
      <div className="bg-[#111113] border border-border rounded-lg w-full max-w-[400px] animate-modal-in">
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-border/50">
          <span className="text-caption text-txt-secondary">{title}</span>
          <button
            onClick={onClose}
            className="text-txt-quaternary hover:text-txt-secondary transition-colors duration-fast cursor-pointer bg-transparent border-none p-0.5"
          >
            <XIcon size={14} />
          </button>
        </div>

        {/* Body */}
        <div className="px-4 py-3">{children}</div>

        {/* Footer */}
        {footer && (
          <div className="flex items-center justify-end gap-2 px-4 py-3 border-t border-border/50">
            {footer}
          </div>
        )}
      </div>
    </div>
  )
}
