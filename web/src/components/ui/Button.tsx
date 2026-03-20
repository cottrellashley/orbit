import type { ButtonHTMLAttributes, ReactNode } from 'react'

type Variant = 'default' | 'primary' | 'danger' | 'ghost'
type Size = 'default' | 'sm' | 'icon'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant
  size?: Size
  children: ReactNode
}

const baseClasses =
  'inline-flex items-center justify-center gap-1.5 font-medium transition-colors duration-fast cursor-pointer border rounded'

const variantClasses: Record<Variant, string> = {
  default:
    'bg-bg-raised border-border text-txt-secondary hover:bg-bg-hover hover:text-txt hover:border-border-hover',
  primary:
    'bg-accent border-accent text-white hover:bg-accent-hover hover:border-accent-hover',
  danger:
    'bg-transparent border-transparent text-semantic-red hover:bg-semantic-red/10 hover:border-semantic-red/20',
  ghost:
    'bg-transparent border-transparent text-txt-tertiary hover:text-txt-secondary hover:bg-white/[0.04]',
}

const sizeClasses: Record<Size, string> = {
  default: 'px-2.5 py-[4px] text-caption',
  sm:      'px-2 py-[3px] text-caption-xs',
  icon:    'p-1 text-caption',
}

export function Button({
  variant = 'default',
  size = 'default',
  className = '',
  children,
  ...props
}: ButtonProps) {
  return (
    <button
      className={`${baseClasses} ${variantClasses[variant]} ${sizeClasses[size]} ${className}`}
      {...props}
    >
      {children}
    </button>
  )
}

export function ButtonGroup({ children }: { children: ReactNode }) {
  return <div className="flex items-center gap-1">{children}</div>
}
