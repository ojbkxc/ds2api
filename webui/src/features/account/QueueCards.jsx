import { CheckCircle2, Server, ShieldCheck } from 'lucide-react'

export default function QueueCards({ queueStatus, t }) {
    if (!queueStatus) {
        return null
    }

    const cards = [
        { icon: CheckCircle2, label: t('accountManager.available'), value: queueStatus.available, unit: t('accountManager.accountsUnit'), color: 'var(--ds-success)' },
        { icon: Server, label: t('accountManager.inUse'), value: queueStatus.in_use, unit: t('accountManager.threadsUnit'), color: 'var(--ds-blue)' },
        { icon: ShieldCheck, label: t('accountManager.totalPool'), value: queueStatus.total, unit: t('accountManager.accountsUnit'), color: 'var(--ds-purple)' },
    ]

    return (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {cards.map((card, idx) => {
                const Icon = card.icon
                return (
                    <div
                        key={idx}
                        className="p-4 flex flex-col justify-between relative overflow-hidden group transition-colors"
                        style={{
                            background: 'var(--ds-card)',
                            border: '1px solid var(--ds-border)',
                            borderRadius: 'var(--radius-card)',
                        }}
                        onMouseEnter={e => { e.currentTarget.style.borderColor = 'var(--ds-border-hover)' }}
                        onMouseLeave={e => { e.currentTarget.style.borderColor = 'var(--ds-border)' }}
                    >
                        <div className="absolute right-0 top-0 p-4 opacity-5 group-hover:opacity-10 transition-opacity" style={{ color: card.color }}>
                            <Icon className="w-16 h-16" />
                        </div>
                        <p className="text-[10px] font-bold uppercase tracking-widest" style={{ color: 'var(--ds-text-tertiary)' }}>{card.label}</p>
                        <div className="mt-2 flex items-baseline gap-2">
                            <span className="text-3xl font-bold" style={{ color: 'var(--ds-text)' }}>{card.value}</span>
                            <span className="text-xs" style={{ color: 'var(--ds-text-tertiary)' }}>{card.unit}</span>
                        </div>
                    </div>
                )
            })}
        </div>
    )
}
