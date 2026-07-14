import { CheckCircle2, ExternalLink, XCircle } from 'lucide-react'

export default function VercelSyncStatus({ t, result }) {
    if (!result) {
        return null
    }

    return (
        <div className="p-6 border animate-in fade-in slide-in-from-right-4" style={{ borderRadius: 'var(--radius-card)', borderColor: result.success ? 'var(--ds-success)' : 'var(--ds-error)', backgroundColor: result.success ? 'var(--ds-success-bg)' : 'var(--ds-error-bg)' }}>
            <div className="flex items-start gap-4">
                {result.success ? (
                    <div className="p-2 rounded-full shadow-lg" style={{ backgroundColor: 'var(--ds-success)', color: 'var(--ds-text-on-primary)' }}>
                        <CheckCircle2 className="w-6 h-6" />
                    </div>
                ) : (
                    <div className="p-2 rounded-full shadow-lg" style={{ backgroundColor: 'var(--ds-error)', color: 'var(--ds-text-on-primary)' }}>
                        <XCircle className="w-6 h-6" />
                    </div>
                )}
                <div className="space-y-1">
                    <h3 className="font-semibold text-lg" style={{ color: result.success ? 'var(--ds-success)' : 'var(--ds-error)' }}>
                        {result.success ? t('vercel.syncSucceeded') : t('vercel.syncFailedLabel')}
                    </h3>
                    <p className="text-sm" style={{ color: 'var(--ds-text-secondary)' }}>{result.message}</p>

                    {result.deployment_url && (
                        <div className="pt-3 mt-3 border-t" style={{ borderColor: result.success ? 'var(--ds-success)' : 'var(--ds-error)' }}>
                            <a href={`https://${result.deployment_url}`} target="_blank" rel="noopener noreferrer" className="inline-flex items-center gap-1 text-sm font-medium hover:underline" style={{ color: 'var(--ds-blue)' }}>
                                {t('vercel.openDeployment')} <ExternalLink className="w-3 h-3" />
                            </a>
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}
