import { Info } from 'lucide-react'

export default function VercelGuide({ t }) {
    return (
        <div className="p-6" style={{ borderRadius: 'var(--radius-card)', border: '1px solid var(--ds-border)', backgroundColor: 'var(--ds-surface)' }}>
            <h3 className="font-semibold flex items-center gap-2 mb-4" style={{ color: 'var(--ds-text)' }}>
                <Info className="w-5 h-5" style={{ color: 'var(--ds-info)' }} />
                {t('vercel.howItWorks')}
            </h3>
            <ul className="space-y-4">
                <li className="flex gap-3">
                    <span className="shrink-0 w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold" style={{ backgroundColor: 'var(--ds-surface)', border: '1px solid var(--ds-border)', color: 'var(--ds-text-tertiary)' }}>1</span>
                    <p className="text-sm" style={{ color: 'var(--ds-text-secondary)' }}>{t('vercel.steps.one')}</p>
                </li>
                <li className="flex gap-3">
                    <span className="shrink-0 w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold" style={{ backgroundColor: 'var(--ds-surface)', border: '1px solid var(--ds-border)', color: 'var(--ds-text-tertiary)' }}>2</span>
                    <p className="text-sm" style={{ color: 'var(--ds-text-secondary)' }}>{t('vercel.steps.two')}</p>
                </li>
                <li className="flex gap-3">
                    <span className="shrink-0 w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold" style={{ backgroundColor: 'var(--ds-surface)', border: '1px solid var(--ds-border)', color: 'var(--ds-text-tertiary)' }}>3</span>
                    <p className="text-sm" style={{ color: 'var(--ds-text-secondary)' }}>
                        {t('vercel.steps.three')} <code className="px-1 py-0.5 rounded text-xs" style={{ backgroundColor: 'var(--ds-bg)', border: '1px solid var(--ds-border)' }}>DS2API_CONFIG_JSON</code>
                    </p>
                </li>
                <li className="flex gap-3">
                    <span className="shrink-0 w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold" style={{ backgroundColor: 'var(--ds-surface)', border: '1px solid var(--ds-border)', color: 'var(--ds-text-tertiary)' }}>4</span>
                    <p className="text-sm" style={{ color: 'var(--ds-text-secondary)' }}>{t('vercel.steps.four')}</p>
                </li>
            </ul>
        </div>
    )
}
