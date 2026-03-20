export function LoadingBar({ visible }: { visible: boolean }) {
  if (!visible) return null

  return (
    <div className="absolute top-0 left-0 right-0 h-[2px] overflow-hidden z-10">
      <div
        className="h-full w-full animate-loading-slide"
        style={{
          background: 'linear-gradient(90deg, transparent, #F97316 40%, #FB923C 60%, transparent)',
        }}
      />
    </div>
  )
}
