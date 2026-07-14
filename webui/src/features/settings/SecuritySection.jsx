import { Lock } from 'lucide-react'

export default function SecuritySection({
    t,
    form,
    setForm,
    newPassword,
    setNewPassword,
    changingPassword,
    onUpdatePassword,
}) {
    return (
        <div className="ds-card p-5 space-y-4">
            <h3 className="font-semibold" style={{ color: 'var(--ds-text)' }}>{t('settings.securityTitle')}</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <label className="text-sm space-y-2">
                    <span style={{ color: 'var(--ds-text-secondary)' }}>{t('settings.jwtExpireHours')}</span>
                    <input
                        type="number"
                        min={1}
                        max={720}
                        value={form.admin.jwt_expire_hours}
                        onChange={(e) => setForm((prev) => ({
                            ...prev,
                            admin: { ...prev.admin, jwt_expire_hours: Number(e.target.value || 1) },
                        }))}
                        className="ds-input"
                    />
                </label>
                <label className="text-sm space-y-2">
                    <span style={{ color: 'var(--ds-text-secondary)' }}>{t('settings.newPassword')}</span>
                    <div className="flex gap-2">
                        <input
                            type="password"
                            value={newPassword}
                            onChange={(e) => setNewPassword(e.target.value)}
                            placeholder={t('settings.newPasswordPlaceholder')}
                            className="ds-input"
                        />
                        <button
                            type="button"
                            onClick={onUpdatePassword}
                            disabled={changingPassword}
                            className="ds-btn-secondary text-sm flex items-center gap-1"
                        >
                            <Lock className="w-4 h-4" />
                            {changingPassword ? t('settings.updating') : t('settings.updatePassword')}
                        </button>
                    </div>
                </label>
            </div>
        </div>
    )
}
