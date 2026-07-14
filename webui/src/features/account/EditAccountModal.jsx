import { X } from 'lucide-react'

export default function EditAccountModal({
    show,
    t,
    editingAccount,
    editAccount,
    setEditAccount,
    loading,
    onClose,
    onSave,
}) {
    if (!show || !editingAccount) {
        return null
    }

    return (
        <div className="ds-modal-overlay" onClick={onClose}>
            <div className="ds-modal-card" style={{ maxWidth: 420 }} onClick={e => e.stopPropagation()}>
                <div className="flex items-start justify-between" style={{ marginBottom: 16, gap: 16 }}>
                    <div style={{ minWidth: 0 }}>
                        <h3 className="ds-modal-title">{t('accountManager.modalEditAccountTitle')}</h3>
                        <p style={{ marginTop: 4, fontSize: 11, color: 'var(--ds-text-tertiary)' }}>
                            {t('accountManager.editAccountHint')}
                        </p>
                    </div>
                    <button
                        onClick={onClose}
                        className="ds-action-btn"
                        style={{ borderRadius: 'var(--radius-ctrl)', padding: 4, flexShrink: 0 }}
                    >
                        <X className="w-4 h-4" />
                    </button>
                </div>

                <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
                    <div
                        style={{
                            borderRadius: 'var(--radius-ctrl)',
                            border: '1px solid var(--ds-border)',
                            background: 'var(--ds-bg)',
                            padding: '8px 12px',
                        }}
                    >
                        <div style={{ fontSize: 11, fontWeight: 600, color: 'var(--ds-text-tertiary)', marginBottom: 4 }}>
                            {t('accountManager.accountIdentifierLabel')}
                        </div>
                        <code
                            style={{
                                fontSize: 13,
                                fontFamily: 'monospace',
                                color: 'var(--ds-text)',
                                wordBreak: 'break-all',
                            }}
                        >
                            {editingAccount.identifier}
                        </code>
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
                            value={editAccount.name}
                            onChange={e => setEditAccount({ ...editAccount, name: e.target.value })}
                            autoFocus
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
                            value={editAccount.remark}
                            onChange={e => setEditAccount({ ...editAccount, remark: e.target.value })}
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
                        onClick={onSave}
                        disabled={loading}
                        className="ds-btn-primary"
                        style={{ padding: '0.5rem 1rem', fontSize: 13 }}
                    >
                        {loading ? t('accountManager.editAccountLoading') : t('accountManager.editAccountAction')}
                    </button>
                </div>
            </div>
        </div>
    )
}