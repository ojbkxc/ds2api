export default function EmptyState({ title, description, icon, actions }) {
  return (
    <div className="ds-empty-state">
      {icon && <div className="ds-empty-state-icon">{icon}</div>}
      <p className="ds-empty-state-title">{title}</p>
      {description && <p className="ds-empty-state-description">{description}</p>}
      {actions && <div className="flex flex-wrap gap-2 justify-center mt-1">{actions}</div>}
    </div>
  )
}