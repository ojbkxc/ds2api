import { X } from 'lucide-react'

export default function AddAccountModal({
    show,
    t,
    newAccount,
    setNewAccount,
    loading,
    onClose,
    onAdd,
}) {
    if (!show) {
        return null
    }

    return (
        <div className="ds-modal-overlay" onClick={onClose}>
            <div className="ds-modal-card" style={{ maxWidth: 420 }} onClick={e => e.stopPropagation()}>
                <div className="flex items-center justify-between" style={{ marginBottom: 16 }}>
                    <h3 className="ds-modal-title">{t('accountManager.modalAddAccountTitle')}</h3>
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
                            {t('accountManager.nameOptional')}
                        </label>
                        <input
                            type="text"
                            className="ds-input"
                            placeholder={t('accountManager.namePlaceholder')}
                            value={newAccount.name}
                            onChange={e => setNewAccount({ ...newAccount, name: e.target.value })}
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
                            value={newAccount.remark}
                            onChange={e => setNewAccount({ ...newAccount, remark: e.target.value })}
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
                            {t('accountManager.emailOptional')}
                        </label>
                        <input
                            type="email"
                            className="ds-input"
                            placeholder="user@example.com"
                            value={newAccount.email}
                            onChange={e => setNewAccount({ ...newAccount, email: e.target.value })}
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
                            {t('accountManager.mobileOptional')}
                        </label>
                        <input
                            type="text"
                            className="ds-input"
                            placeholder="+86..."
                            value={newAccount.mobile}
                            onChange={e => setNewAccount({ ...newAccount, mobile: e.target.value })}
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
                            {t('accountManager.passwordLabel')}{' '}
                            <span style={{ color: 'var(--ds-danger)' }}>*</span>
                        </label>
                        <input
                            type="password"
                            className="ds-input"
                            style={{ background: 'var(--ds-shell-bg)' }}
                            placeholder={t('accountManager.passwordPlaceholder')}
                            value={newAccount.password}
                            onChange={e => setNewAccount({ ...newAccount, password: e.target.value })}
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
                        {loading ? t('accountManager.addAccountLoading') : t('accountManager.addAccountAction')}
                    </button>
                </div>
            </div>
        </div>
    )
}