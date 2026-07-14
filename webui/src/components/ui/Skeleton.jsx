export default function Skeleton({ className = '', width }) {
  return (
    <div
      className={`ds-skeleton rounded ${className}`}
      style={width ? { width } : undefined}
    />
  )
}

export function SkeletonList({ rows = 3 }) {
  return (
    <div className="space-y-2">
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="ds-surface-panel p-3 space-y-2">
          <Skeleton className="h-3" width="60%" />
          <Skeleton className="h-2.5" width="85%" />
        </div>
      ))}
    </div>
  )
}