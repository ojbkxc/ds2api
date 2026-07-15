import { useEffect } from 'react'
import { X } from 'lucide-react'
import { v4 as uuidv4 } from 'uuid'

import { maskSecret } from '../../utils/maskSecret'

export default function AddKeyModal({ show, t, editingKey, newKey, setNewKey, loading, onClose, onAdd }) {
    useEffect(() => {
        if (!show) return
        const onKey = (e) => {
            if (e.key === 'Escape') onClose()
        }
        window.addEventListener('keydown', onKey)
        return () => window.removeEventListener('keydown', onKey)
    }, [show, onClose])

    if (!show) {
        return null
    }

    const isEditing = Boolean(editingKey?.key)
    const displayKey = isEditing ? maskSecret(editingKey?.key || newKey.key) : newKey.key

    return (
        <div className="ds-modal-overlay" onClick={onClose}>
            <div className="ds-modal-card" style={{ maxWidth: 420 }} onClick={e => e.stopPropagation()}>
                <div className="flex items-center justify-between" style={{ marginBottom: 16 }}>
                    <h3 className="ds-modal-title">
                        {isEditing ? t('accountManager.modalEditKeyTitle') : t('accountManager.modalAddKeyTitle')}
                    </h3>
                    <button
                        onClick={onClose}
                        className="ds-action-btn"
                        style={{ borderRadius: 'var(--radius-ctrl)', padding: 4 }}
                    >
                        <X className="w-4 h-4" />
                    </button>
                </div>

                <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
                    <div>
                        <label
                            style={{
                                display: 'block',
                                fontSize: 12,
                                fontWeight: 600,
                                color: 'var(--ds-text-secondary)',
                                marginBottom: 6,
                                textTransform: 'uppercase',
                                letterSpacing: '0.04em',
                            }}
                        >
                            {isEditing ? t('accountManager.keyLabel') : t('accountManager.newKeyLabel')}
                        </label>
                        <div style={{ display: 'flex', gap: 8 }}>
                            <input
                                type="text"
                                className="ds-input"
                                style={{
                                    flex: 1,
                                    ...(isEditing ? { opacity: 0.5, cursor: 'not-allowed' } : { background: 'var(--ds-shell-bg)' }),
                                }}
                                placeholder={isEditing ? t('accountManager.keyReadonlyPlaceholder') : t('accountManager.newKeyPlaceholder')}
                                value={displayKey}
                                onChange={e => setNewKey({ ...newKey, key: e.target.value })}
                                autoFocus={!isEditing}
                                readOnly={isEditing}
                            />
                            {!isEditing && (
                                <button
                                    type="button"
                                    onClick={() => setNewKey({ ...newKey, key: 'sk-' + uuidv4().replace(/-/g, '') })}
                                    className="ds-btn-secondary"
                                    style={{ padding: '0.5rem 0.75rem', fontSize: 12, whiteSpace: 'nowrap' }}
                                >
                                    {t('accountManager.generate')}
                                </button>
                            )}
                        </div>
                        <p style={{ fontSize: 11, color: 'var(--ds-text-tertiary)', marginTop: 6, margin: '6px 0 0 0' }}>
                            {isEditing ? t('accountManager.keyReadonlyHint') : t('accountManager.generateHint')}
                        </p>
                    </div>
                    <div>
                        <label
                            style={{
                                display: 'block',
                                fontSize: 12,
                                fontWeight: 600,
                                color: 'var(--ds-text-secondary)',
                                marginBottom: 6,
                                textTransform: 'uppercase',
                                letterSpacing: '0.04em',
                            }}
                        >
                            {t('accountManager.nameOptional')}
                        </label>
                        <input
                            type="text"
                            className="ds-input"
                            placeholder={t('accountManager.namePlaceholder')}
                            value={newKey.name}
                            onChange={e => setNewKey({ ...newKey, name: e.target.value })}
                            autoFocus={isEditing}
                        />
                    </div>
                    <div>
                        <label
                            style={{
                                display: 'block',
                                fontSize: 12,
                                fontWeight: 600,
                                color: 'var(--ds-text-secondary)',
                                marginBottom: 6,
                                textTransform: 'uppercase',
                                letterSpacing: '0.04em',
                            }}
                        >
                            {t('accountManager.remarkOptional')}
                        </label>
                        <input
                            type="text"
                            className="ds-input"
                            placeholder={t('accountManager.remarkPlaceholder')}
                            value={newKey.remark}
                            onChange={e => setNewKey({ ...newKey, remark: e.target.value })}
                        />
                    </div>
                </div>

                <div className="ds-modal-actions">
                    <button
                        onClick={onClose}
                        className="ds-btn-secondary"
                        style={{ padding: '0.5rem 1rem', fontSize: 13 }}
                    >
                        {t('actions.cancel')}
                    </button>
                    <button
                        onClick={onAdd}
                        disabled={loading}
                        className="ds-btn-primary"
                        style={{ padding: '0.5rem 1rem', fontSize: 13 }}
                    >
                        {loading
                            ? (isEditing ? t('accountManager.editKeyLoading') : t('accountManager.addKeyLoading'))
                            : (isEditing ? t('accountManager.editKeyAction') : t('accountManager.addKeyAction'))}
                    </button>
                </div>
            </div>
        </div>
    )
}