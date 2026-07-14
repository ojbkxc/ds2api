export default function CurrentInputFileSection({ t, form, setForm }) {
    return (
        <div className="ds-card p-5 space-y-4">
            <div className="space-y-1">
                <h3 className="font-semibold" style={{ color: 'var(--ds-text)' }}>{t('settings.currentInputFileTitle')}</h3>
                <p className="text-sm" style={{ color: 'var(--ds-text-secondary)' }}>{t('settings.currentInputFileDesc')}</p>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <label className="flex items-start gap-3 p-4" style={{ borderRadius: 'var(--radius-ctrl)', border: '1px solid var(--ds-border)', background: 'var(--ds-bg)' }}>
                    <input
                        type="checkbox"
                        checked={Boolean(form.current_input_file?.enabled)}
                        onChange={(e) => setForm((prev) => ({
                            ...prev,
                            current_input_file: {
                                ...prev.current_input_file,
                                enabled: e.target.checked,
                            },
                        }))}
                        className="mt-1 h-4 w-4"
                        style={{ borderRadius: 'var(--radius-ctrl)', accentColor: 'var(--ds-blue)' }}
                    />
                    <div className="space-y-1">
                        <span className="text-sm font-medium block" style={{ color: 'var(--ds-text)' }}>{t('settings.currentInputFileEnabled')}</span>
                        <span className="text-xs block" style={{ color: 'var(--ds-text-tertiary)' }}>{t('settings.currentInputFileDesc')}</span>
                    </div>
                </label>
                <label className="text-sm space-y-2">
                    <span style={{ color: 'var(--ds-text-secondary)' }}>{t('settings.currentInputFileMinChars')}</span>
                    <input
                        type="number"
                        min={0}
                        max={100000000}
                        value={form.current_input_file?.min_chars ?? 0}
                        onChange={(e) => setForm((prev) => ({
                            ...prev,
                            current_input_file: {
                                ...prev.current_input_file,
                                min_chars: Number(e.target.value || 0),
                            },
                        }))}
                        className="ds-input"
                    />
                    <p className="text-xs" style={{ color: 'var(--ds-text-tertiary)' }}>{t('settings.currentInputFileHelp')}</p>
                </label>
                <label className="text-sm space-y-2">
                    <span style={{ color: 'var(--ds-text-secondary)' }}>{t('settings.currentInputFileFilenameTemplate', { defaultValue: 'Filename Template' })}</span>
                    <input
                        type="text"
                        placeholder="deepseek{time}.txt"
                        value={form.current_input_file?.filename_template ?? ''}
                        onChange={(e) => setForm((prev) => ({
                            ...prev,
                            current_input_file: {
                                ...prev.current_input_file,
                                filename_template: e.target.value,
                            },
                        }))}
                        className="ds-input"
                    />
                    <p className="text-xs" style={{ color: 'var(--ds-text-tertiary)' }}>{t('settings.currentInputFileFilenameTemplateHelp', { defaultValue: 'Use {time} for last 4 digits of timestamp, {timestamp} for full timestamp. Leave empty for default name.' })}</p>
                </label>
            </div>
        </div>
    )
}
