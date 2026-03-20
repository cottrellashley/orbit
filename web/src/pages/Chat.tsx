import { TerminalView } from '../components/Terminal'

export function Chat() {
  return (
    <div className="flex flex-col flex-1 min-h-0">
      {/* Minimal path */}
      <div className="flex items-center px-4 py-2 shrink-0">
        <span className="text-caption text-txt-secondary">Chat</span>
      </div>

      {/* Terminal fills everything */}
      <TerminalView opts={{}} />
    </div>
  )
}
