import { Trash2 } from 'lucide-react'

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
                <Trash2 className="w-4 h-4" style={{ color: 'var(--ds-text-tertiary)' }} />
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
            <p className="text-xs" style={{ color: mode === 'none' ? 'var(--ds-text-tertiary)' : 'var(--ds-warning)' }}>
                {t(descKey)}
            </p>
            {mode !== 'none' && (
                <p className="text-xs flex items-center gap-1" style={{ color: 'var(--ds-warning)' }}>
                    {t('settings.autoDeleteWarning')}
                </p>
            )}
        </div>
    )
}
