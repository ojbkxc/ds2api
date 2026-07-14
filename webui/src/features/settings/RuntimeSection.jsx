export default function RuntimeSection({ t, form, setForm }) {
    return (
        <div className="ds-card p-5 space-y-4">
            <h3 className="font-semibold" style={{ color: 'var(--ds-text)' }}>{t('settings.runtimeTitle')}</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <label className="text-sm space-y-2">
                    <span style={{ color: 'var(--ds-text-secondary)' }}>{t('settings.accountMaxInflight')}</span>
                    <input
                        type="number"
                        min={1}
                        value={form.runtime.account_max_inflight}
                        onChange={(e) => setForm((prev) => ({
                            ...prev,
                            runtime: { ...prev.runtime, account_max_inflight: Number(e.target.value || 1) },
                        }))}
                        className="ds-input"
                    />
                </label>
                <label className="text-sm space-y-2">
                    <span style={{ color: 'var(--ds-text-secondary)' }}>{t('settings.accountMaxQueue')}</span>
                    <input
                        type="number"
                        min={1}
                        value={form.runtime.account_max_queue}
                        onChange={(e) => setForm((prev) => ({
                            ...prev,
                            runtime: { ...prev.runtime, account_max_queue: Number(e.target.value || 1) },
                        }))}
                        className="ds-input"
                    />
                </label>
                <label className="text-sm space-y-2">
                    <span style={{ color: 'var(--ds-text-secondary)' }}>{t('settings.globalMaxInflight')}</span>
                    <input
                        type="number"
                        min={1}
                        value={form.runtime.global_max_inflight}
                        onChange={(e) => setForm((prev) => ({
                            ...prev,
                            runtime: { ...prev.runtime, global_max_inflight: Number(e.target.value || 1) },
                        }))}
                        className="ds-input"
                    />
                </label>
                <label className="text-sm space-y-2">
                    <span style={{ color: 'var(--ds-text-secondary)' }}>{t('settings.tokenRefreshIntervalHours')}</span>
                    <input
                        type="number"
                        min={1}
                        max={720}
                        step={1}
                        value={form.runtime.token_refresh_interval_hours}
                        onChange={(e) => setForm((prev) => ({
                            ...prev,
                            runtime: { ...prev.runtime, token_refresh_interval_hours: Number(e.target.value || 1) },
                        }))}
                        className="ds-input"
                    />
                </label>
            </div>
        </div>
    )
}
