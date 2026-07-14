import { useState, useCallback, useMemo } from 'react'
import { Trash2, Edit3, Copy, Plus, Key, Ban, Search, ToggleLeft, ToggleRight } from 'lucide-react'
import clsx from 'clsx'
import { useI18n } from '../../i18n'
import StatusDot from '../../components/ui/StatusDot'
import Badge from '../../components/ui/Badge'
import ConfirmDialog from '../../components/ui/ConfirmDialog'

const PAGE_SIZE = 20

export default function AccountsTable({ accounts, onEdit, onDelete, onAddKey, onAddAccount, onToggleDisabled, togglingDisabled }) {
    const { t } = useI18n()
    const [search, setSearch] = useState('')
    const [deleteTarget, setDeleteTarget] = useState(null)
    const [page, setPage] = useState(0)

    const filtered = useMemo(() => {
        if (!search.trim()) return accounts
        const q = search.toLowerCase()
        return accounts.filter(a => a.email?.toLowerCase().includes(q) || a.plan_type?.toLowerCase().includes(q))
    }, [accounts, search])

    const totalPages = Math.max(1, Math.ceil(filtered.length / PAGE_SIZE))
    const safePage = Math.min(page, totalPages - 1)
    const paged = filtered.slice(safePage * PAGE_SIZE, (safePage + 1) * PAGE_SIZE)

    const statusTone = (status) => {
        switch (status) {
            case 'active': return 'success'
            case 'banned': return 'purple'
            case 'failed': return 'danger'
            case 'disabled': return 'muted'
            default: return 'info'
        }
    }

    const handleCopy = useCallback(async (text) => {
        try {
            await navigator.clipboard.writeText(text)
        } catch { /* ignore */ }
    }, [])

    return (
        <div className="space-y-4">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                <div className="relative flex-1 max-w-sm">
                    <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none" style={{ color: 'var(--ds-text-tertiary)' }}>
                        <Search className="w-3.5 h-3.5" />
                    </div>
                    <input
                        className="w-full pl-9 pr-3 py-2 text-xs"
                        style={{
                            background: 'var(--ds-bg)',
                            border: '1px solid var(--ds-border)',
                            borderRadius: 'var(--radius-ctrl)',
                            color: 'var(--ds-text)',
                        }}
                        placeholder={t('accounts.searchPlaceholder')}
                        value={search}
                        onChange={e => { setSearch(e.target.value); setPage(0) }}
                    />
                </div>
                <button
                    onClick={onAddAccount}
                    className="ds-btn-primary text-xs"
                >
                    <Plus className="w-3.5 h-3.5 mr-1.5" />
                    {t('accounts.addAccount')}
                </button>
            </div>

            <div className="border overflow-hidden" style={{ borderColor: 'var(--ds-border)', borderRadius: 'var(--radius-card)', background: 'var(--ds-card)' }}>
                <div className="overflow-x-auto">
                    <table className="w-full text-xs">
                        <thead>
                            <tr style={{ borderBottom: '1px solid var(--ds-border)', background: 'var(--ds-bg)' }}>
                                <th className="text-left px-4 py-3 font-semibold uppercase tracking-wider" style={{ color: 'var(--ds-text-secondary)' }}>{t('accounts.status')}</th>
                                <th className="text-left px-4 py-3 font-semibold uppercase tracking-wider" style={{ color: 'var(--ds-text-secondary)' }}>{t('accounts.email')}</th>
                                <th className="text-left px-4 py-3 font-semibold uppercase tracking-wider" style={{ color: 'var(--ds-text-secondary)' }}>{t('accounts.plan')}</th>
                                <th className="text-left px-4 py-3 font-semibold uppercase tracking-wider" style={{ color: 'var(--ds-text-secondary)' }}>{t('accounts.keys')}</th>
                                <th className="text-right px-4 py-3 font-semibold uppercase tracking-wider" style={{ color: 'var(--ds-text-secondary)' }}>{t('accounts.actions')}</th>
                            </tr>
                        </thead>
                        <tbody>
                            {paged.length === 0 ? (
                                <tr>
                                    <td colSpan={5} className="px-4 py-16 text-center" style={{ color: 'var(--ds-text-tertiary)' }}>
                                        {search ? t('accounts.noResults') : t('accounts.noAccounts')}
                                    </td>
                                </tr>
                            ) : (
                                paged.map((account, i) => (
                                    <tr
                                        key={account.email || i}
                                        className="transition-colors"
                                        style={{ borderBottom: '1px solid var(--ds-border)' }}
                                        onMouseEnter={e => { e.currentTarget.style.background = 'var(--ds-surface)' }}
                                        onMouseLeave={e => { e.currentTarget.style.background = 'transparent' }}
                                    >
                                        <td className="px-4 py-3">
                                            <div className="flex items-center gap-2">
                                                <StatusDot status={account.status} />
                                                <Badge tone={statusTone(account.status)}>
                                                    {account.status}
                                                </Badge>
                                            </div>
                                        </td>
                                        <td className="px-4 py-3 font-medium" style={{ color: 'var(--ds-text)' }}>
                                            {account.email}
                                        </td>
                                        <td className="px-4 py-3" style={{ color: 'var(--ds-text-secondary)' }}>
                                            {account.plan_type || '-'}
                                        </td>
                                        <td className="px-4 py-3">
                                            <span className="font-mono" style={{ color: 'var(--ds-text-secondary)' }}>
                                                {account.session_token ? (
                                                    <span className="flex items-center gap-1">
                                                        <Key className="w-3 h-3" />
                                                        <span className="font-mono" style={{ color: 'var(--ds-text-tertiary)' }}>{account.session_token.substring(0, 8)}...</span>
                                                    </span>
                                                ) : (
                                                    <span className="text-xs" style={{ color: 'var(--ds-text-tertiary)' }}>-</span>
                                                )}
                                            </span>
                                        </td>
                                        <td className="px-4 py-3">
                                            <div className="flex items-center justify-end gap-1">
                                                <button
                                                    onClick={() => handleCopy(account.session_token)}
                                                    className="ds-action-btn p-1.5"
                                                    title={t('accounts.copyToken')}
                                                    style={{ borderRadius: 'var(--radius-ctrl)' }}
                                                >
                                                    <Copy className="w-3.5 h-3.5" />
                                                </button>
                                                <button
                                                    onClick={() => onAddKey(account)}
                                                    className="ds-action-btn p-1.5"
                                                    title={t('accounts.addKey')}
                                                    style={{ borderRadius: 'var(--radius-ctrl)' }}
                                                >
                                                    <Key className="w-3.5 h-3.5" />
                                                </button>
                                                <button
                                                    onClick={() => onEdit(account)}
                                                    className="ds-action-btn p-1.5"
                                                    title={t('accounts.edit')}
                                                    style={{ borderRadius: 'var(--radius-ctrl)' }}
                                                >
                                                    <Edit3 className="w-3.5 h-3.5" />
                                                </button>
                                                {onToggleDisabled && (
                                                    <button
                                                        onClick={() => onToggleDisabled(account)}
                                                        disabled={togglingDisabled === account.email}
                                                        className="ds-action-btn p-1.5"
                                                        title={account.disabled ? t('accounts.enable') : t('accounts.disable')}
                                                        style={{
                                                            borderRadius: 'var(--radius-ctrl)',
                                                            color: account.disabled ? 'var(--ds-text-tertiary)' : 'var(--ds-warning)',
                                                        }}
                                                    >
                                                        {account.disabled ? <ToggleLeft className="w-3.5 h-3.5" /> : <ToggleRight className="w-3.5 h-3.5" />}
                                                    </button>
                                                )}
                                                <button
                                                    onClick={() => setDeleteTarget(account)}
                                                    className="ds-action-btn p-1.5"
                                                    title={t('accounts.delete')}
                                                    style={{ borderRadius: 'var(--radius-ctrl)', color: 'var(--ds-danger)' }}
                                                >
                                                    <Trash2 className="w-3.5 h-3.5" />
                                                </button>
                                            </div>
                                        </td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>

                {totalPages > 1 && (
                    <div className="flex items-center justify-between px-4 py-3 border-t" style={{ borderColor: 'var(--ds-border)' }}>
                        <span className="text-xs" style={{ color: 'var(--ds-text-tertiary)' }}>
                            {filtered.length} {t('accounts.total')} · {t('accounts.page')} {safePage + 1}/{totalPages}
                        </span>
                        <div className="flex gap-1">
                            <button
                                className="ds-btn-secondary px-2.5 py-1 text-[10px]"
                                disabled={safePage === 0}
                                onClick={() => setPage(p => Math.max(0, p - 1))}
                            >
                                {t('common.prev')}
                            </button>
                            <button
                                className="ds-btn-secondary px-2.5 py-1 text-[10px]"
                                disabled={safePage >= totalPages - 1}
                                onClick={() => setPage(p => Math.min(totalPages - 1, p + 1))}
                            >
                                {t('common.next')}
                            </button>
                        </div>
                    </div>
                )}
            </div>

            <ConfirmDialog
                open={!!deleteTarget}
                title={t('accounts.deleteConfirmTitle')}
                message={t('accounts.deleteConfirmMessage', { email: deleteTarget?.email || '' })}
                confirmLabel={t('common.delete')}
                cancelLabel={t('common.cancel')}
                onConfirm={() => { onDelete(deleteTarget); setDeleteTarget(null) }}
                onCancel={() => setDeleteTarget(null)}
            />
        </div>
    )
}