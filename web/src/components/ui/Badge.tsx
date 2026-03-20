const statusStyles: Record<string, { text: string; bg: string; dot: string; pulse?: boolean }> = {
  healthy:   { text: 'text-semantic-green', bg: 'bg-semantic-green-soft', dot: 'bg-semantic-green' },
  pass:      { text: 'text-semantic-green', bg: 'bg-semantic-green-soft', dot: 'bg-semantic-green' },
  idle:      { text: 'text-semantic-green', bg: 'bg-semantic-green-soft', dot: 'bg-semantic-green' },
  completed: { text: 'text-semantic-green', bg: 'bg-semantic-green-soft', dot: 'bg-semantic-green' },
  running:   { text: 'text-semantic-blue',  bg: 'bg-semantic-blue-soft',  dot: 'bg-semantic-blue', pulse: true },
  busy:      { text: 'text-semantic-yellow', bg: 'bg-semantic-yellow-soft', dot: 'bg-semantic-yellow', pulse: true },
  building:  { text: 'text-semantic-yellow', bg: 'bg-semantic-yellow-soft', dot: 'bg-semantic-yellow', pulse: true },
  fail:      { text: 'text-semantic-red',   bg: 'bg-semantic-red-soft',   dot: 'bg-semantic-red' },
  failed:    { text: 'text-semantic-red',   bg: 'bg-semantic-red-soft',   dot: 'bg-semantic-red' },
  error:     { text: 'text-semantic-red',   bg: 'bg-semantic-red-soft',   dot: 'bg-semantic-red' },
  stopped:   { text: 'text-semantic-red',   bg: 'bg-semantic-red-soft',   dot: 'bg-semantic-red' },
  warn:      { text: 'text-semantic-orange', bg: 'bg-semantic-orange-soft', dot: 'bg-semantic-orange' },
  retry:     { text: 'text-semantic-orange', bg: 'bg-semantic-orange-soft', dot: 'bg-semantic-orange' },
}

const defaultStyle = { text: 'text-txt-quaternary', bg: 'bg-white/[0.04]', dot: 'bg-txt-quaternary' }

export function StatusBadge({ status }: { status: string }) {
  const s = statusStyles[status] ?? defaultStyle
  const label = status.charAt(0).toUpperCase() + status.slice(1)

  return (
    <span className={`inline-flex items-center gap-[5px] px-1.5 py-[1px] rounded ${s.bg} text-caption-xs ${s.text}`}>
      <span
        className={`w-[5px] h-[5px] rounded-full ${s.dot} ${s.pulse ? 'animate-pulse-dot' : ''}`}
      />
      {label}
    </span>
  )
}
