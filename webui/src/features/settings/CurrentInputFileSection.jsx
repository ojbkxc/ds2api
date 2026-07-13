export default function CurrentInputFileSection({ t, form, setForm }) {
    return (
        <div className="bg-card border border-border rounded-xl p-5 space-y-4">
            <div className="space-y-1">
                <h3 className="font-semibold">{t('settings.currentInputFileTitle')}</h3>
                <p className="text-sm text-muted-foreground">{t('settings.currentInputFileDesc')}</p>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <label className="flex items-start gap-3 rounded-lg border border-border bg-background/60 p-4">
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
                        className="mt-1 h-4 w-4 rounded border-border"
                    />
                    <div className="space-y-1">
                        <span className="text-sm font-medium block">{t('settings.currentInputFileEnabled')}</span>
                        <span className="text-xs text-muted-foreground block">{t('settings.currentInputFileDesc')}</span>
                    </div>
                </label>
                <label className="text-sm space-y-2">
                    <span className="text-muted-foreground">{t('settings.currentInputFileMinChars')}</span>
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
                        className="w-full bg-background border border-border rounded-lg px-3 py-2"
                    />
                    <p className="text-xs text-muted-foreground">{t('settings.currentInputFileHelp')}</p>
                </label>
                <label className="text-sm space-y-2">
                    <span className="text-muted-foreground">{t('settings.currentInputFileFilenameTemplate', { defaultValue: 'Filename Template' })}</span>
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
                        className="w-full bg-background border border-border rounded-lg px-3 py-2"
                    />
                    <p className="text-xs text-muted-foreground">{t('settings.currentInputFileFilenameTemplateHelp', { defaultValue: 'Use {time} for last 4 digits of timestamp, {timestamp} for full timestamp. Leave empty for default name.' })}</p>
                </label>
                <label className="text-sm space-y-2">
                    <span className="text-muted-foreground">{t('settings.currentInputFileDisabledModels', { defaultValue: 'Disabled Models' })}</span>
                    <input
                        type="text"
                        placeholder="deepseek-v4-pro, deepseek-v4-pro-search"
                        value={(form.current_input_file?.disabled_models ?? []).join(', ')}
                        onChange={(e) => setForm((prev) => ({
                            ...prev,
                            current_input_file: {
                                ...prev.current_input_file,
                                disabled_models: e.target.value.split(',').map(s => s.trim()).filter(Boolean),
                            },
                        }))}
                        className="w-full bg-background border border-border rounded-lg px-3 py-2"
                    />
                    <p className="text-xs text-muted-foreground">{t('settings.currentInputFileDisabledModelsHelp', { defaultValue: 'Comma-separated model names that cannot upload files.' })}</p>
                </label>
                <label className="text-sm space-y-2">
                    <span className="text-muted-foreground">{t('settings.currentInputFileVisionAccounts', { defaultValue: 'Vision Accounts' })}</span>
                    <input
                        type="text"
                        placeholder="4@email.lxseek.com, 12@email.lxseek.com"
                        value={(form.current_input_file?.vision_accounts ?? []).join(', ')}
                        onChange={(e) => setForm((prev) => ({
                            ...prev,
                            current_input_file: {
                                ...prev.current_input_file,
                                vision_accounts: e.target.value.split(',').map(s => s.trim()).filter(Boolean),
                            },
                        }))}
                        className="w-full bg-background border border-border rounded-lg px-3 py-2"
                    />
                    <p className="text-xs text-muted-foreground">{t('settings.currentInputFileVisionAccountsHelp', { defaultValue: 'Comma-separated account identifiers that support vision models.' })}</p>
                </label>
                <label className="text-sm space-y-2">
                    <span className="text-muted-foreground">{t('settings.currentInputFileDisabledAccounts', { defaultValue: 'Disabled Accounts' })}</span>
                    <input
                        type="text"
                        placeholder="2@email.lxseek.com, 5@email.lxseek.com"
                        value={(form.current_input_file?.disabled_accounts ?? []).join(', ')}
                        onChange={(e) => setForm((prev) => ({
                            ...prev,
                            current_input_file: {
                                ...prev.current_input_file,
                                disabled_accounts: e.target.value.split(',').map(s => s.trim()).filter(Boolean),
                            },
                        }))}
                        className="w-full bg-background border border-border rounded-lg px-3 py-2"
                    />
                    <p className="text-xs text-muted-foreground">{t('settings.currentInputFileDisabledAccountsHelp', { defaultValue: 'Comma-separated account identifiers that are excluded from file upload.' })}</p>
                </label>
            </div>
        </div>
    )
}
