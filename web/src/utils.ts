export function formatTime(t: string | undefined | null): string {
  if (!t) return '-'
  const d = new Date(t)
  const now = Date.now()
  const diff = now - d.getTime()
  const sec = Math.floor(diff / 1000)
  if (sec < 60) return 'just now'
  const min = Math.floor(sec / 60)
  if (min < 60) return `${min}m ago`
  const hr = Math.floor(min / 60)
  if (hr < 24) return `${hr}h ago`
  return d.toLocaleDateString()
}
