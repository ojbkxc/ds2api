import { Trash2, AlertTriangle } from 'lucide-react'

export default function AutoDeleteSection({ t, form, setForm }) {
    const mode = form.auto_delete?.mode || 'none'
    const descKey = mode === 'single'
        ? 'settings.autoDeleteSingleDesc'
        : mode === 'all'
            ? 'settings.autoDeleteAllDesc'
            : 'settings.autoDeleteNoneDesc'

    return (
        <div className="ds-card p-5 space-y-4">
            <div className="flex items-center gap-2">
                <div style={{
                    padding: '6px',
                    borderRadius: 'var(--radius-ctrl)',
                    background: 'var(--ds-danger-bg)',
                    border: '1px solid var(--ds-danger-border)',
                }}>
                    <Trash2 className="w-4 h-4" style={{ color: 'var(--ds-danger)' }} />
                </div>
                <h3 className="font-semibold" style={{ color: 'var(--ds-text)' }}>{t('settings.autoDeleteTitle')}</h3>
            </div>
            <p className="text-sm" style={{ color: 'var(--ds-text-secondary)' }}>{t('settings.autoDeleteDesc')}</p>
            <div className="space-y-1.5">
                <label className="text-sm font-medium leading-6" style={{ color: 'var(--ds-text-secondary)' }}>{t('settings.autoDeleteMode')}</label>
                <select
                    value={mode}
                    onChange={(e) => setForm((prev) => ({
                        ...prev,
                        auto_delete: { ...(prev.auto_delete || {}), mode: e.target.value },
                    }))}
                    className="ds-input text-sm"
                >
                    <option value="none">{t('settings.autoDeleteNone')}</option>
                    <option value="single">{t('settings.autoDeleteSingle')}</option>
                    <option value="all">{t('settings.autoDeleteAll')}</option>
                </select>
            </div>
            <div className="p-3 border" style={{
                borderRadius: 'var(--radius-ctrl)',
                borderColor: mode === 'none' ? 'var(--ds-border)' : 'var(--ds-warning-border)',
                background: mode === 'none' ? 'var(--ds-bg)' : 'var(--ds-warning-bg)',
            }}>
                <p className="text-xs" style={{ color: mode === 'none' ? 'var(--ds-text-tertiary)' : 'var(--ds-warning)' }}>
                    {t(descKey)}
                </p>
            </div>
            {mode !== 'none' && (
                <div className="flex items-start gap-2 p-3 border" style={{
                    borderRadius: 'var(--radius-ctrl)',
                    borderColor: 'var(--ds-danger-border)',
                    background: 'var(--ds-danger-bg)',
                }}>
                    <AlertTriangle className="w-3.5 h-3.5 shrink-0 mt-0.5" style={{ color: 'var(--ds-danger)' }} />
                    <p className="text-xs" style={{ color: 'var(--ds-danger)' }}>
                        {t('settings.autoDeleteWarning')}
                    </p>
                </div>
            )}
        </div>
    )
}
