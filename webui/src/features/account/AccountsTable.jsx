import { useState } from 'react'
import { Trash2, Edit3, Copy, Check, Plus, Key, Search, ToggleLeft, ToggleRight, Loader2, RefreshCw, Globe, Zap } from 'lucide-react'
import clsx from 'clsx'
import { useI18n } from '../../i18n'
import StatusDot from '../../components/ui/StatusDot'
import Badge from '../../components/ui/Badge'
import ConfirmDialog from '../../components/ui/ConfirmDialog'
import { copyToClipboard } from '../../utils/copyToClipboard'

export default function AccountsTable({
    t,
    accounts,
    loadingAccounts,
    testing,
    testingAll,
    batchProgress,
    sessionCounts,
    deletingSessions,
    updatingProxy,
    togglingDisabled,
    totalAccounts,
    page,
    pageSize,
    totalPages,
    resolveAccountIdentifier,
    proxies,
    onTestAll,
    onShowAddAccount,
    onEditAccount,
    onTestAccount,
    onDeleteAccount,
    onDeleteAllSessions,
    onUpdateAccountProxy,
    onToggleDisabled,
    onPrevPage,
    onNextPage,
    onPageSizeChange,
    searchQuery,
    onSearchChange,
    envBacked,
}) {
    const [deleteTarget, setDeleteTarget] = useState(null)
    const [sessionDeleteTarget, setSessionDeleteTarget] = useState(null)

    const statusTone = (status) => {
        switch (status) {
            case 'active': return 'success'
            case 'banned': return 'purple'
            case 'failed': return 'danger'
            case 'disabled': return 'muted'
            default: return 'info'
        }
    }

    const [copiedAccountId, setCopiedAccountId] = useState(null)

    const handleCopyAccount = async (accountId) => {
        try {
            await copyToClipboard(accountId)
            setCopiedAccountId(accountId)
            setTimeout(() => {
                setCopiedAccountId(prev => prev === accountId ? null : prev)
            }, 2000)
        } catch {
            // Best-effort copy; ignore failures in non-secure contexts.
        }
    }

    const safePage = Math.min(page || 1, Math.max(1, totalPages || 1))

    return (
        <div className="space-y-4">
            {/* Header: search + actions */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                <div className="relative flex-1 max-w-sm">
                    <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none" style={{ color: 'var(--ds-text-tertiary)' }}>
                        <Search className="w-3.5 h-3.5" />
                    </div>
                    <input
                        className="ds-input pl-9 text-xs"
                        placeholder={t('accountManager.searchPlaceholder')}
                        value={searchQuery || ''}
                        onChange={e => onSearchChange(e.target.value)}
                    />
                </div>
                <div className="flex items-center gap-2">
                    <button
                        onClick={onTestAll}
                        disabled={testingAll}
                        className="ds-btn-secondary text-xs"
                    >
                        {testingAll ? (
                            <Loader2 className="w-3.5 h-3.5 mr-1.5 animate-spin" />
                        ) : (
                            <RefreshCw className="w-3.5 h-3.5 mr-1.5" />
                        )}
                        {testingAll ? t('accountManager.testingAllAccounts') : t('accountManager.testAll')}
                    </button>
                    <button
                        onClick={onShowAddAccount}
                        className="ds-btn-primary text-xs"
                    >
                        <Plus className="w-3.5 h-3.5 mr-1.5" />
                        {t('accountManager.addAccount')}
                    </button>
                </div>
            </div>

            {/* Batch progress bar */}
            {testingAll && batchProgress?.total > 0 && (
                <div className="p-3 border" style={{ borderColor: 'var(--ds-border)', borderRadius: 'var(--radius-ctrl)', background: 'var(--ds-bg)' }}>
                    <div className="flex items-center justify-between mb-2">
                        <span className="text-xs font-medium" style={{ color: 'var(--ds-text-secondary)' }}>
                            {batchProgress.current} / {batchProgress.total}
                        </span>
                        <span className="text-xs" style={{ color: 'var(--ds-text-tertiary)' }}>
                            {batchProgress.results?.filter(r => r.success).length || 0} {t('accountManager.available')}
                        </span>
                    </div>
                    <div className="h-1.5 overflow-hidden" style={{ background: 'var(--ds-surface)', borderRadius: 'var(--radius-pill)' }}>
                        <div
                            className="h-full transition-all duration-300"
                            style={{
                                width: `${(batchProgress.current / batchProgress.total) * 100}%`,
                                background: 'var(--ds-blue)',
                                borderRadius: 'var(--radius-pill)',
                            }}
                        />
                    </div>
                </div>
            )}

            {/* Table */}
            <div className="border overflow-hidden" style={{ borderColor: 'var(--ds-border)', borderRadius: 'var(--radius-card)', background: 'var(--ds-card)' }}>
                <div className="overflow-x-auto">
                    <table className="w-full text-xs">
                        <thead>
                            <tr style={{ borderBottom: '1px solid var(--ds-border)', background: 'var(--ds-bg)' }}>
                                <th className="text-left px-4 py-3 font-semibold uppercase tracking-wider" style={{ color: 'var(--ds-text-secondary)' }}>{t('accountManager.accountsTitle')}</th>
                                <th className="text-left px-4 py-3 font-semibold uppercase tracking-wider" style={{ color: 'var(--ds-text-secondary)' }}>{t('accountManager.accountProxyLabel')}</th>
                                <th className="text-left px-4 py-3 font-semibold uppercase tracking-wider" style={{ color: 'var(--ds-text-secondary)' }}>{t('common.sessions')}</th>
                                <th className="text-right px-4 py-3 font-semibold uppercase tracking-wider" style={{ color: 'var(--ds-text-secondary)' }}>{t('common.actions')}</th>
                            </tr>
                        </thead>
                        <tbody>
                            {loadingAccounts && accounts.length === 0 ? (
                                <tr>
                                    <td colSpan={4} className="px-4 py-16 text-center">
                                        <Loader2 className="w-5 h-5 animate-spin mx-auto mb-2" style={{ color: 'var(--ds-text-tertiary)' }} />
                                        <span className="text-xs" style={{ color: 'var(--ds-text-tertiary)' }}>{t('common.loading')}</span>
                                    </td>
                                </tr>
                            ) : accounts.length === 0 ? (
                                <tr>
                                    <td colSpan={4} className="px-4 py-16 text-center" style={{ color: 'var(--ds-text-tertiary)' }}>
                                        {searchQuery ? t('accountManager.searchNoResults') : t('accountManager.noAccounts')}
                                    </td>
                                </tr>
                            ) : (
                                accounts.map((account, i) => {
                                    const identifier = resolveAccountIdentifier(account)
                                    const isTesting = testing && testing[identifier]
                                    const isDeletingSessions = deletingSessions && deletingSessions[identifier]
                                    const isUpdatingProxy = updatingProxy && updatingProxy[identifier]
                                    const isToggling = togglingDisabled && togglingDisabled[identifier]
                                    const sessionCount = sessionCounts?.[identifier]

                                    return (
                                        <tr
                                            key={identifier || i}
                                            className="transition-colors"
                                            style={{
                                                borderBottom: '1px solid var(--ds-border)',
                                                opacity: account.disabled ? 0.45 : 1,
                                            }}
                                            onMouseEnter={e => { e.currentTarget.style.background = 'var(--ds-surface)' }}
                                            onMouseLeave={e => { e.currentTarget.style.background = 'transparent' }}
                                        >
                                            {/* Account info */}
                                            <td className="px-4 py-3">
                                                <div className="flex items-center gap-3">
                                                    <StatusDot status={account.status} pulse={account.status === 'active'} />
                                                    <div className="min-w-0 flex-1">
                                                        <div className="flex items-center gap-2">
                                                            <span className="font-medium truncate" style={{ color: 'var(--ds-text)' }}>
                                                                {account.name || account.email || identifier}
                                                            </span>
                                                            {account.disabled && (
                                                                <Badge tone="muted">{t('accountManager.accountDisabled')}</Badge>
                                                            )}
                                                        </div>
                                                        <div className="flex items-center gap-2 mt-0.5">
                                                            <span className="text-[10px]" style={{ color: 'var(--ds-text-tertiary)' }}>
                                                                {account.email || account.mobile || '-'}
                                                            </span>
                                                            {account.plan_type && (
                                                                <span className="text-[10px] px-1.5 py-0.5" style={{ color: 'var(--ds-text-tertiary)', background: 'var(--ds-bg)', borderRadius: 'var(--radius-ctrl)' }}>
                                                                    {account.plan_type}
                                                                </span>
                                                            )}
                                                        </div>
                                                    </div>
                                                </div>
                                            </td>

                                            {/* Proxy selector */}
                                            <td className="px-4 py-3">
                                                <div className="flex items-center gap-1.5">
                                                    <Globe className="w-3 h-3 shrink-0" style={{ color: 'var(--ds-text-tertiary)' }} />
                                                    <select
                                                        value={account.proxy_id || ''}
                                                        onChange={e => onUpdateAccountProxy(identifier, e.target.value)}
                                                        disabled={isUpdatingProxy || envBacked}
                                                        className="text-[10px] py-1 px-2"
                                                        style={{
                                                            background: 'var(--ds-bg)',
                                                            border: '1px solid var(--ds-border)',
                                                            borderRadius: 'var(--radius-ctrl)',
                                                            color: 'var(--ds-text-secondary)',
                                                            cursor: 'pointer',
                                                            maxWidth: '120px',
                        }}
                                                    >
                                                        <option value="">{t('accountManager.proxyNone')}</option>
                                                        {(proxies || []).map(p => (
                                                            <option key={p.id} value={p.id}>{p.name}</option>
                                                        ))}
                                                    </select>
                                                    {isUpdatingProxy && <Loader2 className="w-3 h-3 animate-spin" style={{ color: 'var(--ds-text-tertiary)' }} />}
                                                </div>
                                            </td>

                                            {/* Session count */}
                                            <td className="px-4 py-3">
                                                <div className="flex items-center gap-1.5">
                                                    {sessionCount !== undefined ? (
                                                        <>
                                                            <span className="text-xs font-mono" style={{ color: 'var(--ds-text-secondary)' }}>
                                                                {sessionCount}
                                                            </span>
                                                            <button
                                                                onClick={() => setSessionDeleteTarget({ identifier, email: account.email })}
                                                                disabled={isDeletingSessions}
                                                                className="ds-action-btn p-1"
                                                                title={t('accountManager.deleteAllSessions')}
                                                                style={{ borderRadius: 'var(--radius-ctrl)' }}
                                                            >
                                                                {isDeletingSessions ? (
                                                                    <Loader2 className="w-3 h-3 animate-spin" />
                                                                ) : (
                                                                    <Trash2 className="w-3 h-3" />
                                                                )}
                                                            </button>
                                                        </>
                                                    ) : (
                                                        <span className="text-[10px]" style={{ color: 'var(--ds-text-tertiary)' }}>-</span>
                                                    )}
                                                </div>
                                            </td>

                                            {/* Actions */}
                                            <td className="px-4 py-3">
                                                <div className="flex items-center justify-end gap-1">
                                                    <button
                                                        onClick={() => onTestAccount(identifier)}
                                                        disabled={isTesting}
                                                        className="ds-action-btn p-1.5"
                                                        title={t('actions.test')}
                                                        style={{ borderRadius: 'var(--radius-ctrl)' }}
                                                    >
                                                        {isTesting ? (
                                                            <Loader2 className="w-3.5 h-3.5 animate-spin" />
                                                        ) : (
                                                            <Zap className="w-3.5 h-3.5" />
                                                        )}
                                                    </button>
                                                    <button
                                                        onClick={() => handleCopyAccount(identifier)}
                                                        className="ds-action-btn p-1.5"
                                                        title={t('accountManager.copyAccountTitle')}
                                                        style={{ borderRadius: 'var(--radius-ctrl)' }}
                                                    >
                                                        {copiedAccountId === identifier ? (
                                                            <Check className="w-3.5 h-3.5" style={{ color: 'var(--ds-success)' }} />
                                                        ) : (
                                                            <Copy className="w-3.5 h-3.5" />
                                                        )}
                                                    </button>
                                                    <button
                                                        onClick={() => onEditAccount(account)}
                                                        className="ds-action-btn p-1.5"
                                                        title={t('accountManager.editAccountTitle')}
                                                        style={{ borderRadius: 'var(--radius-ctrl)' }}
                                                    >
                                                        <Edit3 className="w-3.5 h-3.5" />
                                                    </button>
                                                    {onToggleDisabled && (
                                                        <button
                                                            onClick={() => onToggleDisabled(account)}
                                                            disabled={isToggling}
                                                            className="ds-action-btn p-1.5"
                                                            title={account.disabled ? t('accountManager.enableAccount') : t('accountManager.disableAccount')}
                                                            style={{
                                                                borderRadius: 'var(--radius-ctrl)',
                                                                color: account.disabled ? 'var(--ds-text-tertiary)' : 'var(--ds-warning)',
                                                            }}
                                                        >
                                                            {isToggling ? (
                                                                <Loader2 className="w-3.5 h-3.5 animate-spin" />
                                                            ) : account.disabled ? (
                                                                <ToggleLeft className="w-3.5 h-3.5" />
                                                            ) : (
                                                                <ToggleRight className="w-3.5 h-3.5" />
                                                            )}
                                                        </button>
                                                    )}
                                                    <button
                                                        onClick={() => setDeleteTarget(account)}
                                                        className="ds-action-btn p-1.5"
                                                        title={t('common.delete')}
                                                        style={{ borderRadius: 'var(--radius-ctrl)', color: 'var(--ds-danger)' }}
                                                    >
                                                        <Trash2 className="w-3.5 h-3.5" />
                                                    </button>
                                                </div>
                                            </td>
                                        </tr>
                                    )
                                })
                            )}
                        </tbody>
                    </table>
                </div>

                {/* Pagination */}
                {(totalPages || 1) > 1 && (
                    <div className="flex items-center justify-between px-4 py-3 border-t" style={{ borderColor: 'var(--ds-border)' }}>
                        <span className="text-xs" style={{ color: 'var(--ds-text-tertiary)' }}>
                            {totalAccounts || 0} {t('accountManager.accountsUnit')} · {safePage}/{totalPages || 1}
                        </span>
                        <div className="flex items-center gap-2">
                            <select
                                value={pageSize || 10}
                                onChange={e => onPageSizeChange(Number(e.target.value))}
                                className="text-[10px] py-1 px-2"
                                style={{
                                    background: 'var(--ds-bg)',
                                    border: '1px solid var(--ds-border)',
                                    borderRadius: 'var(--radius-ctrl)',
                                    color: 'var(--ds-text-secondary)',
                                    cursor: 'pointer',
                                }}
                            >
                                <option value={10}>10</option>
                                <option value={20}>20</option>
                                <option value={50}>50</option>
                            </select>
                            <div className="flex gap-1">
                                <button
                                    className="ds-btn-secondary px-2.5 py-1 text-[10px]"
                                    disabled={safePage <= 1}
                                    onClick={onPrevPage}
                                >
                                    {t('common.prev')}
                                </button>
                                <button
                                    className="ds-btn-secondary px-2.5 py-1 text-[10px]"
                                    disabled={safePage >= (totalPages || 1)}
                                    onClick={onNextPage}
                                >
                                    {t('common.next')}
                                </button>
                            </div>
                        </div>
                    </div>
                )}
            </div>

            {/* Delete account confirm */}
            <ConfirmDialog
                open={!!deleteTarget}
                title={t('accountManager.deleteAccountConfirm')}
                message={t('accountManager.deleteAccountConfirm')}
                confirmLabel={t('common.delete')}
                cancelLabel={t('common.cancel')}
                onConfirm={() => { onDeleteAccount(resolveAccountIdentifier(deleteTarget)); setDeleteTarget(null) }}
                onCancel={() => setDeleteTarget(null)}
            />

            {/* Delete all sessions confirm */}
            <ConfirmDialog
                open={!!sessionDeleteTarget}
                title={t('accountManager.deleteAllSessions')}
                message={t('accountManager.deleteAllSessionsConfirm')}
                confirmLabel={t('common.delete')}
                cancelLabel={t('common.cancel')}
                onConfirm={() => {
                    onDeleteAllSessions(sessionDeleteTarget?.identifier)
                    setSessionDeleteTarget(null)
                }}
                onCancel={() => setSessionDeleteTarget(null)}
            />
        </div>
    )
}
