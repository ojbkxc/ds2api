export default function BehaviorSection({ t, form, setForm }) {
    return (
        <div className="ds-card p-5 space-y-4">
            <h3 className="font-semibold" style={{ color: 'var(--ds-text)' }}>{t('settings.behaviorTitle')}</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <label className="text-sm space-y-2">
                    <span style={{ color: 'var(--ds-text-secondary)' }}>{t('settings.responsesTTL')}</span>
                    <input
                        type="number"
                        min={30}
                        value={form.responses.store_ttl_seconds}
                        onChange={(e) => setForm((prev) => ({
                            ...prev,
                            responses: { ...prev.responses, store_ttl_seconds: Number(e.target.value || 30) },
                        }))}
                        className="ds-input"
                    />
                </label>
                <label className="text-sm space-y-2">
                    <span style={{ color: 'var(--ds-text-secondary)' }}>{t('settings.embeddingsProvider')}</span>
                    <input
                        type="text"
                        value={form.embeddings.provider}
                        onChange={(e) => setForm((prev) => ({
                            ...prev,
                            embeddings: { ...prev.embeddings, provider: e.target.value },
                        }))}
                        className="ds-input"
                    />
                </label>
                <label className="flex items-start gap-3 p-4" style={{ borderRadius: 'var(--radius-ctrl)', border: '1px solid var(--ds-border)', background: 'var(--ds-bg)' }}>
                    <input
                        type="checkbox"
                        checked={Boolean(form.thinking_injection?.enabled ?? true)}
                        onChange={(e) => setForm((prev) => ({
                            ...prev,
                            thinking_injection: {
                                ...prev.thinking_injection,
                                enabled: e.target.checked,
                            },
                        }))}
                        className="mt-1 h-4 w-4"
                        style={{ borderRadius: 'var(--radius-ctrl)', accentColor: 'var(--ds-blue)' }}
                    />
                    <div className="space-y-1">
                        <span className="text-sm font-medium block" style={{ color: 'var(--ds-text)' }}>{t('settings.thinkingInjectionEnabled')}</span>
                        <span className="text-xs block" style={{ color: 'var(--ds-text-tertiary)' }}>{t('settings.thinkingInjectionDesc')}</span>
                    </div>
                </label>
                <label className="text-sm space-y-2 md:col-span-2">
                    <span style={{ color: 'var(--ds-text-secondary)' }}>{t('settings.thinkingInjectionPrompt')}</span>
                    <textarea
                        rows={5}
                        value={form.thinking_injection?.prompt || ''}
                        placeholder={form.thinking_injection?.default_prompt || ''}
                        onChange={(e) => setForm((prev) => ({
                            ...prev,
                            thinking_injection: {
                                ...prev.thinking_injection,
                                prompt: e.target.value,
                            },
                        }))}
                        className="ds-input font-mono text-xs"
                        style={{ resize: 'vertical', minHeight: '128px' }}
                    />
                    <p className="text-xs" style={{ color: 'var(--ds-text-tertiary)' }}>{t('settings.thinkingInjectionPromptHelp')}</p>
                </label>
            </div>
        </div>
    )
}
