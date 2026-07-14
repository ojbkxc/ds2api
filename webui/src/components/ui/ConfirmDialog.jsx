import { useEffect } from 'react'
import { X } from 'lucide-react'

export default function ConfirmDialog({ open, title, message, confirmLabel, cancelLabel, onConfirm, onCancel }) {
  useEffect(() => {
    if (!open) return
    const onKey = (e) => {
      if (e.key === 'Escape') onCancel()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [open, onCancel])

  if (!open) return null

  return (
    <div className="ds-modal-overlay" role="dialog" aria-modal="true" onClick={onCancel}>
      <div className="ds-modal-card max-w-sm" onClick={e => e.stopPropagation()}>
        <div className="flex items-start justify-between gap-3">
          <div>
            <h3 className="ds-modal-title">{title}</h3>
            <p className="ds-modal-message">{message}</p>
          </div>
          <button onClick={onCancel} className="ds-action-btn p-1 rounded" style={{ borderRadius: 'var(--radius-ctrl)' }}>
            <X className="w-4 h-4" />
          </button>
        </div>
        <div className="ds-modal-actions">
          <button type="button" className="ds-btn-cancel px-3 py-2 text-[11px] font-medium" style={{ borderRadius: 'var(--radius-ctrl)' }} onClick={onCancel}>
            {cancelLabel}
          </button>
          <button type="button" className="ds-btn-danger px-3 py-2 text-[11px] font-medium" style={{ borderRadius: 'var(--radius-ctrl)' }} onClick={onConfirm} autoFocus>
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  )
}