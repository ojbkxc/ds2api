import Input from '../../components/ui/Input'

export default function ModelSection({ t, form, setForm }) {
    return (
        <div className="ds-card p-5" style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            <h3 className="font-semibold" style={{ color: 'var(--ds-text)' }}>{t('settings.modelTitle')}</h3>
            <label className="text-sm" style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                <span style={{ color: 'var(--ds-text-secondary)' }}>{t('settings.modelAliases')}</span>
                <textarea
                    value={form.model_aliases_text}
                    onChange={(e) => setForm((prev) => ({ ...prev, model_aliases_text: e.target.value }))}
                    rows={12}
                    className="ds-input font-mono text-xs"
                    style={{ resize: 'vertical' }}
                />
            </label>
        </div>
    )
}