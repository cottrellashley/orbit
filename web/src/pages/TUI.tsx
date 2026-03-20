import { useApp } from '../state/context'
import { TerminalView } from '../components/Terminal'

export function TUI() {
  const { state, navigate } = useApp()
  const opts = state.tuiOpts

  const title = opts?.title || 'Session'

  return (
    <div className="flex flex-col flex-1 min-h-0">
      {/* Minimal breadcrumb path */}
      <div className="flex items-center gap-1.5 px-4 py-2 shrink-0">
        <button
          onClick={() => navigate('sessions')}
          className="text-caption text-txt-quaternary hover:text-txt-secondary transition-colors duration-fast cursor-pointer bg-transparent border-none"
        >
          Sessions
        </button>
        <span className="text-caption text-txt-quaternary">/</span>
        <span className="text-caption text-txt-secondary truncate">{title}</span>
      </div>

      {/* Terminal fills everything */}
      {opts ? (
        <TerminalView opts={opts} />
      ) : (
        <div className="flex items-center justify-center flex-1 text-txt-tertiary text-caption">
          No session selected
        </div>
      )}
    </div>
  )
}
