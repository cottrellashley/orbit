const statusStyles: Record<string, { text: string; dot: string; pulse?: boolean }> = {
  healthy:   { text: 'text-semantic-green', dot: 'bg-semantic-green' },
  pass:      { text: 'text-semantic-green', dot: 'bg-semantic-green' },
  idle:      { text: 'text-semantic-green', dot: 'bg-semantic-green' },
  completed: { text: 'text-semantic-green', dot: 'bg-semantic-green' },
  running:   { text: 'text-semantic-blue',  dot: 'bg-semantic-blue', pulse: true },
  busy:      { text: 'text-semantic-yellow', dot: 'bg-semantic-yellow', pulse: true },
  building:  { text: 'text-semantic-yellow', dot: 'bg-semantic-yellow', pulse: true },
  fail:      { text: 'text-semantic-red',   dot: 'bg-semantic-red' },
  failed:    { text: 'text-semantic-red',   dot: 'bg-semantic-red' },
  error:     { text: 'text-semantic-red',   dot: 'bg-semantic-red' },
  stopped:   { text: 'text-semantic-red',   dot: 'bg-semantic-red' },
  warn:      { text: 'text-semantic-orange', dot: 'bg-semantic-orange' },
  retry:     { text: 'text-semantic-orange', dot: 'bg-semantic-orange' },
}

const defaultStyle = { text: 'text-txt-quaternary', dot: 'bg-txt-quaternary' }

export function StatusBadge({ status }: { status: string }) {
  const s = statusStyles[status] ?? defaultStyle
  const label = status.charAt(0).toUpperCase() + status.slice(1)

  return (
    <span className={`inline-flex items-center gap-[5px] text-caption-xs ${s.text}`}>
      <span
        className={`w-[5px] h-[5px] rounded-full ${s.dot} ${s.pulse ? 'animate-pulse-dot' : ''}`}
      />
      {label}
    </span>
  )
}
