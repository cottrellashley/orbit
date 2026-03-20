export function Spinner({ size = 16, className = '' }: { size?: number; className?: string }) {
  return (
    <span
      className={`inline-block animate-spin rounded-full border-2 border-border-hover border-t-accent ${className}`}
      style={{ width: size, height: size }}
    />
  )
}
