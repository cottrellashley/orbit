import { MainHeader } from '../components/layout/MainHeader'

export function ComingSoon({ title }: { title: string }) {
  return (
    <div className="flex flex-col flex-1 min-h-0">
      <MainHeader title={title} />
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center">
          <div className="text-caption text-txt-tertiary">Coming soon</div>
          <div className="text-caption-xs text-txt-quaternary mt-1">{title} is under development.</div>
        </div>
      </div>
    </div>
  )
}
