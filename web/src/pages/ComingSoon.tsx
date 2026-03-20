import { MainHeader } from '../components/layout/MainHeader'

export function ComingSoon({ title }: { title: string }) {
  return (
    <div className="flex flex-col flex-1 min-h-0">
      <MainHeader title={title} />
      <div className="flex-1 flex items-center justify-center">
        <div className="bg-bg-raised border border-border-subtle rounded-lg px-8 py-8 text-center">
          <div className="text-label text-txt-secondary">Coming soon</div>
          <div className="text-caption-xs text-txt-quaternary mt-1">{title} is under development.</div>
        </div>
      </div>
    </div>
  )
}
