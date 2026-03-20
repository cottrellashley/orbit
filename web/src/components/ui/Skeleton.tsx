export function Skeleton({ className = '', width, height }: { className?: string; width?: string; height?: string }) {
  return (
    <div
      className={`rounded-lg bg-white animate-skeleton ${className}`}
      style={{ width, height }}
    />
  )
}

export function SkeletonRow() {
  return (
    <div className="flex items-center gap-3 py-2.5 px-3">
      <Skeleton width="120px" height="14px" />
      <Skeleton width="80px" height="14px" />
      <Skeleton width="60px" height="14px" />
      <Skeleton width="100px" height="14px" />
    </div>
  )
}

export function SkeletonCard() {
  return (
    <div className="bg-bg-raised border border-border-subtle rounded-lg p-4">
      <div className="flex items-center gap-3 mb-3">
        <Skeleton width="36px" height="36px" className="rounded-lg" />
        <div>
          <Skeleton width="48px" height="20px" className="mb-1" />
          <Skeleton width="72px" height="12px" />
        </div>
      </div>
    </div>
  )
}
