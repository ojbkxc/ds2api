import { AlertTriangle, Save } from 'lucide-react'

import { useI18n } from '../../i18n'
import { useSettingsForm } from './useSettingsForm'
import Button from '../../components/ui/Button'
import SecuritySection from './SecuritySection'
import RuntimeSection from './RuntimeSection'
import BehaviorSection from './BehaviorSection'
import CurrentInputFileSection from './CurrentInputFileSection'
import AutoDeleteSection from './AutoDeleteSection'
import ModelSection from './ModelSection'
import BackupSection from './BackupSection'

export default function SettingsContainer({ onRefresh, onMessage, authFetch, onForceLogout, isVercel = false }) {
    const { t } = useI18n()
    const apiFetch = authFetch || fetch

    const {
        form,
        setForm,
        loading,
        saving,
        changingPassword,
        importing,
        exportData,
        importMode,
        setImportMode,
        importText,
        setImportText,
        newPassword,
        setNewPassword,
        consecutiveFailures,
        autoFetchPaused,
        lastError,
        settingsMeta,
        syncHintVisible,
        retryLoadSettings,
        saveSettings,
        updatePassword,
        loadExportData,
        downloadExportFile,
        loadImportFile,
        doImport,
    } = useSettingsForm({
        apiFetch,
        t,
        onMessage,
        onRefresh,
        onForceLogout,
        isVercel,
    })

    return (
        <div className="space-y-6">
            {autoFetchPaused && (
                <div
                    className="p-4 flex items-center justify-between gap-4"
                    style={{
                        background: 'var(--ds-danger-bg)',
                        border: '1px solid var(--ds-danger-border)',
                        borderRadius: 'var(--radius-ctrl)',
                        color: 'var(--ds-danger)',
                    }}
                >
                    <div className="flex items-center gap-2">
                        <AlertTriangle className="w-4 h-4" />
                        <span className="text-sm">
                            {t('settings.autoFetchPaused', { count: consecutiveFailures, error: lastError || t('settings.loadFailed') })}
                        </span>
                    </div>
                    <button
                        type="button"
                        onClick={retryLoadSettings}
                        className="px-3 py-1.5 text-xs font-medium"
                        style={{
                            borderRadius: 'var(--radius-ctrl)',
                            border: '1px solid var(--ds-danger-border)',
                            background: 'transparent',
                            color: 'var(--ds-danger)',
                            cursor: 'pointer',
                        }}
                    >
                        {t('settings.retryLoad')}
                    </button>
                </div>
            )}
            {settingsMeta.default_password_warning && (
                <div
                    className="p-4 flex items-center gap-2"
                    style={{
                        background: 'var(--ds-warning-bg)',
                        border: '1px solid var(--ds-warning-border)',
                        borderRadius: 'var(--radius-ctrl)',
                        color: 'var(--ds-warning)',
                    }}
                >
                    <AlertTriangle className="w-4 h-4" />
                    <span className="text-sm">{t('settings.defaultPasswordWarning')}</span>
                </div>
            )}
            {syncHintVisible && (
                <div
                    className="p-4 flex items-center gap-2"
                    style={{
                        background: 'var(--ds-warning-bg)',
                        border: '1px solid var(--ds-warning-border)',
                        borderRadius: 'var(--radius-ctrl)',
                        color: 'var(--ds-warning)',
                    }}
                >
                    <AlertTriangle className="w-4 h-4" />
                    <span className="text-sm">{t('settings.vercelSyncHint')}</span>
                </div>
            )}

            <SecuritySection
                t={t}
                form={form}
                setForm={setForm}
                newPassword={newPassword}
                setNewPassword={setNewPassword}
                changingPassword={changingPassword}
                onUpdatePassword={updatePassword}
            />

            <RuntimeSection t={t} form={form} setForm={setForm} />

            <BehaviorSection t={t} form={form} setForm={setForm} />

            <CurrentInputFileSection t={t} form={form} setForm={setForm} />

            <AutoDeleteSection t={t} form={form} setForm={setForm} />

            <ModelSection t={t} form={form} setForm={setForm} />

            <BackupSection
                t={t}
                importMode={importMode}
                setImportMode={setImportMode}
                importing={importing}
                onLoadExportData={loadExportData}
                onDownloadExportFile={downloadExportFile}
                onImport={doImport}
                onImportFileChange={loadImportFile}
                importText={importText}
                setImportText={setImportText}
                exportData={exportData}
            />

            <div className="flex justify-end">
                <Button
                    type="button"
                    variant="primary"
                    size="md"
                    onClick={saveSettings}
                    disabled={loading || saving}
                >
                    <Save className="w-4 h-4" />
                    <span className="ml-1.5">{saving ? t('settings.saving') : t('settings.save')}</span>
                </Button>
            </div>
        </div>
    )
}