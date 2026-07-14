import { ArrowRight, CheckCircle2, Cloud, ExternalLink, RefreshCw } from 'lucide-react'
import clsx from 'clsx'

export default function VercelSyncForm({
    t,
    syncStatus,
    pollPaused,
    pollFailures,
    onManualRefresh,
    preconfig,
    vercelToken,
    setVercelToken,
    projectId,
    setProjectId,
    teamId,
    setTeamId,
    saveCredentials,
    setSaveCredentials,
    loading,
    onSync,
}) {
    return (
        <div className="ds-card shadow-sm p-6 space-y-6">
            <div className="border-b pb-6" style={{ borderColor: 'var(--ds-border)' }}>
                <div className="flex items-center justify-between">
                    <h2 className="text-xl font-semibold flex items-center gap-2" style={{ color: 'var(--ds-text)' }}>
                        <Cloud className="w-6 h-6" style={{ color: 'var(--ds-blue)' }} />
                        {t('vercel.title')}
                    </h2>
                    {syncStatus && (
                        <div className="flex items-center gap-1.5 text-xs font-semibold px-2.5 py-1 rounded-full border transition-colors" style={{
                            color: syncStatus.synced ? 'var(--ds-success)' : syncStatus.has_synced_before ? 'var(--ds-warning)' : 'var(--ds-text-tertiary)',
                            backgroundColor: syncStatus.synced ? 'var(--ds-success-bg)' : syncStatus.has_synced_before ? 'var(--ds-warning-bg)' : 'var(--ds-surface)',
                            borderColor: syncStatus.synced ? 'var(--ds-success)' : syncStatus.has_synced_before ? 'var(--ds-warning)' : 'var(--ds-border)',
                        }}>
                            <span className="w-1.5 h-1.5 rounded-full" style={{ backgroundColor: syncStatus.synced ? 'var(--ds-success)' : syncStatus.has_synced_before ? 'var(--ds-warning)' : 'var(--ds-text-tertiary)' }} />
                            {syncStatus.synced
                                ? t('vercel.statusSynced')
                                : syncStatus.has_synced_before
                                    ? t('vercel.statusNotSynced')
                                    : t('vercel.statusNeverSynced')}
                        </div>
                    )}
                </div>
                <p className="text-sm mt-1" style={{ color: 'var(--ds-text-secondary)' }}>
                    {t('vercel.description')}
                </p>
                {pollPaused && (
                    <div className="mt-2 flex flex-wrap items-center gap-2">
                        <p className="text-xs" style={{ color: 'var(--ds-error)' }}>
                            {t('vercel.pollPaused', { count: pollFailures })}
                        </p>
                        <button
                            type="button"
                            onClick={onManualRefresh}
                            className="px-2 py-1 text-xs rounded border hover:bg-secondary/50"
                            style={{ borderColor: 'var(--ds-border)' }}
                        >
                            {t('vercel.manualRefresh')}
                        </button>
                    </div>
                )}
                {syncStatus?.last_sync_time && (
                    <p className="text-xs mt-1.5 flex items-center gap-1" style={{ color: 'var(--ds-text-tertiary)' }}>
                        <RefreshCw className="w-3 h-3" />
                        {t('vercel.lastSyncTime', { time: new Date(syncStatus.last_sync_time * 1000).toLocaleString() })}
                    </p>
                )}
                {syncStatus?.draft_differs && (
                    <p className="text-xs mt-2" style={{ color: 'var(--ds-warning)' }}>
                        {t('vercel.draftDiffers')}
                    </p>
                )}
            </div>

            <div className="space-y-4">
                <div className="space-y-2">
                    <label className="text-sm font-medium flex items-center justify-between" style={{ color: 'var(--ds-text-secondary)' }}>
                        {t('vercel.tokenLabel')}
                        <a href="https://vercel.com/account/tokens" target="_blank" rel="noopener noreferrer" className="text-xs hover:underline flex items-center gap-1" style={{ color: 'var(--ds-blue)' }}>
                            {t('vercel.getToken')} <ExternalLink className="w-3 h-3" />
                        </a>
                    </label>
                    <div className="relative">
                        <input
                            type="password"
                            className="ds-input text-sm"
                            placeholder={preconfig?.has_token ? t('vercel.tokenPlaceholderPreconfig') : t('vercel.tokenPlaceholder')}
                            value={vercelToken}
                            onChange={e => setVercelToken(e.target.value)}
                        />
                        {preconfig?.has_token && !vercelToken && (
                            <div className="absolute right-3 top-2.5" style={{ color: 'var(--ds-success)' }}>
                                <CheckCircle2 className="w-5 h-5" />
                            </div>
                        )}
                    </div>
                </div>

                <div className="space-y-2">
                    <label className="text-sm font-medium" style={{ color: 'var(--ds-text-secondary)' }}>{t('vercel.projectIdLabel')}</label>
                    <input
                        type="text"
                        className="ds-input text-sm"
                        placeholder="prj_xxxxxxxxxxxx or Project Name"
                        value={projectId}
                        onChange={e => setProjectId(e.target.value)}
                    />
                    <p className="text-xs" style={{ color: 'var(--ds-text-tertiary)' }}>{t('vercel.projectIdHint')}</p>
                </div>

                <div className="space-y-2">
                    <label className="text-sm font-medium flex items-center gap-2" style={{ color: 'var(--ds-text-secondary)' }}>
                        {t('vercel.teamIdLabel')} <span className="text-xs font-normal" style={{ color: 'var(--ds-text-tertiary)' }}>({t('vercel.optional')})</span>
                    </label>
                    <input
                        type="text"
                        className="ds-input text-sm"
                        placeholder="team_xxxxxxxxxxxx"
                        value={teamId}
                        onChange={e => setTeamId(e.target.value)}
                    />
                </div>

                <label className="flex items-start gap-3 text-sm">
                    <input
                        type="checkbox"
                        className="mt-1 h-4 w-4"
                        style={{ borderRadius: 'var(--radius-ctrl)', accentColor: 'var(--ds-blue)' }}
                        checked={saveCredentials}
                        onChange={e => setSaveCredentials(e.target.checked)}
                    />
                    <span className="space-y-1">
                        <span className="block font-medium" style={{ color: 'var(--ds-text)' }}>{t('vercel.saveCredentials')}</span>
                        <span className="block text-xs" style={{ color: 'var(--ds-text-tertiary)' }}>{t('vercel.saveCredentialsHint')}</span>
                    </span>
                </label>
            </div>

            <div className="pt-4">
                <button
                    onClick={onSync}
                    disabled={loading}
                    className="ds-btn-primary w-full flex items-center justify-center gap-2 py-3 text-sm font-medium"
                >
                    {loading ? (
                        <span className="flex items-center gap-2">
                            <span className="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin" />
                            {t('vercel.syncing')}
                        </span>
                    ) : (
                        <span className="flex items-center gap-2">
                            {t('vercel.syncRedeploy')} <ArrowRight className="w-4 h-4" />
                        </span>
                    )}
                </button>
                <p className="text-xs text-center mt-4" style={{ color: 'var(--ds-text-tertiary)' }}>
                    {t('vercel.redeployHint')}
                </p>
            </div>
        </div>
    )
}
