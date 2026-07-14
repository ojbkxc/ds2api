import { ArrowUp, Loader2, MessageSquareText, Trash2, X } from 'lucide-react'
import clsx from 'clsx'

import Badge from '../../components/ui/Badge'
import ConfirmDialog from '../../components/ui/ConfirmDialog'
import EmptyState from '../../components/ui/EmptyState'

import DetailConversation from './ChatHistoryDetail'
import { ListModeIcon, MergeModeIcon } from './HistoryModeIcons'
import { formatDateTime, previewText } from './chatHistoryUtils'

function statusBadgeTone(status) {
    switch (status) {
        case 'success':
            return 'success'
        case 'error':
            return 'danger'
        case 'stopped':
            return 'warning'
        default:
            return 'muted'
    }
}

function ViewModeToggle({ t, viewMode, setViewMode, mobile = false }) {
    const size = mobile ? 'h-9 w-10' : 'h-9 w-12'
    return (
        <div
            className="inline-flex items-center p-1"
            style={{
                background: 'var(--ds-bg)',
                border: '1px solid var(--ds-border)',
                borderRadius: 'var(--radius-ctrl)',
            }}
        >
            <button
                type="button"
                onClick={() => setViewMode('list')}
                className={clsx(size, 'flex items-center justify-center transition-colors')}
                style={{
                    borderRadius: 'var(--radius-ctrl)',
                    background: viewMode === 'list' ? 'var(--ds-surface)' : 'transparent',
                    color: viewMode === 'list' ? 'var(--ds-text)' : 'var(--ds-text-secondary)',
                }}
                title={t('chatHistory.viewModeList')}
            >
                <ListModeIcon />
            </button>
            <button
                type="button"
                onClick={() => setViewMode('merged')}
                className={clsx(size, 'flex items-center justify-center transition-colors')}
                style={{
                    borderRadius: 'var(--radius-ctrl)',
                    background: viewMode === 'merged' ? 'var(--ds-surface)' : 'transparent',
                    color: viewMode === 'merged' ? 'var(--ds-text)' : 'var(--ds-text-secondary)',
                }}
                title={t('chatHistory.viewModeMerged')}
            >
                <MergeModeIcon />
            </button>
        </div>
    )
}

export function ChatHistoryListPane({ items, selectedItem, deletingId, t, lang, onSelectItem, onDeleteItem }) {
    return (
        <div className="ds-card min-h-0 overflow-hidden flex flex-col">
            <div
                className="px-4 py-3 flex items-center justify-between"
                style={{ borderBottom: '1px solid var(--ds-border)' }}
            >
                <div className="text-sm font-semibold" style={{ color: 'var(--ds-text)' }}>{t('chatHistory.listTitle')}</div>
                <div className="text-xs" style={{ color: 'var(--ds-text-secondary)' }}>{items.length}</div>
            </div>
            <div className="flex-1 overflow-y-auto p-3 space-y-3">
                {!items.length && (
                    <EmptyState
                        icon={<MessageSquareText className="w-8 h-8" />}
                        title={t('chatHistory.emptyTitle')}
                        description={t('chatHistory.emptyDesc')}
                    />
                )}

                {items.map(item => (
                    <button
                        key={item.id}
                        type="button"
                        onClick={(event) => onSelectItem(item.id, event)}
                        className="w-full text-left px-4 py-3 transition-colors"
                        style={{
                            borderRadius: 'var(--radius-ctrl)',
                            background: selectedItem?.id === item.id ? 'var(--ds-blue-light)' : 'transparent',
                            border: selectedItem?.id === item.id ? '1px solid var(--ds-selected-border)' : '1px solid var(--ds-border)',
                        }}
                    >
                        <div className="flex items-start justify-between gap-3">
                            <div className="min-w-0">
                                <div className="text-sm font-semibold truncate" style={{ color: 'var(--ds-text)' }}>
                                    {item.user_input || t('chatHistory.untitled')}
                                </div>
                                <div className="text-[11px] mt-1 truncate" style={{ color: 'var(--ds-text-secondary)' }}>
                                    {[item.surface, item.model].filter(Boolean).join(' \u00b7 ') || '-'}
                                </div>
                            </div>
                            <div className="flex items-center gap-2 shrink-0">
                                <Badge tone={statusBadgeTone(item.status)}>
                                    {t(`chatHistory.status.${item.status || 'streaming'}`)}
                                </Badge>
                                <button
                                    type="button"
                                    onClick={(event) => {
                                        event.stopPropagation()
                                        onDeleteItem(item.id)
                                    }}
                                    disabled={deletingId === item.id}
                                    className="ds-action-btn p-1.5"
                                    style={{ borderRadius: 'var(--radius-ctrl)' }}
                                >
                                    {deletingId === item.id ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Trash2 className="w-3.5 h-3.5" />}
                                </button>
                            </div>
                        </div>
                        <div
                            className="text-xs mt-3 line-clamp-2 whitespace-pre-wrap break-words"
                            style={{ color: 'var(--ds-text-secondary)' }}
                        >
                            {previewText(item) || t('chatHistory.noPreview')}
                        </div>
                        <div className="text-[11px] mt-3" style={{ color: 'var(--ds-text-tertiary)' }}>
                            {formatDateTime(item.completed_at || item.updated_at || item.created_at, lang)}
                        </div>
                    </button>
                ))}
            </div>
        </div>
    )
}

export function DesktopDetailPane({ selectedSummary, selectedItem, t, lang, viewMode, setViewMode, detailScrollRef, assistantStartRef, onMessage }) {
    return (
        <div className="hidden lg:flex ds-card min-h-0 overflow-hidden flex-col relative">
            <div
                className="px-5 py-4 flex items-center justify-between gap-3"
                style={{ borderBottom: '1px solid var(--ds-border)' }}
            >
                <div>
                    <div className="text-sm font-semibold" style={{ color: 'var(--ds-text)' }}>{t('chatHistory.detailTitle')}</div>
                    <div className="text-xs mt-1" style={{ color: 'var(--ds-text-secondary)' }}>
                        {selectedSummary
                            ? formatDateTime(selectedSummary.completed_at || selectedSummary.updated_at || selectedSummary.created_at, lang)
                            : t('chatHistory.selectPrompt')}
                    </div>
                </div>
                <div className="flex items-center gap-2">
                    <ViewModeToggle t={t} viewMode={viewMode} setViewMode={setViewMode} />
                    <button
                        type="button"
                        onClick={() => detailScrollRef.current?.scrollTo({ top: 0, behavior: 'smooth' })}
                        className="ds-btn-secondary h-8 w-8 p-0"
                        title={t('chatHistory.backToTop')}
                    >
                        <ArrowUp className="w-4 h-4" />
                    </button>
                    {selectedSummary && (
                        <Badge tone={statusBadgeTone(selectedSummary.status)}>
                            {t(`chatHistory.status.${selectedSummary.status || 'streaming'}`)}
                        </Badge>
                    )}
                </div>
            </div>

            <div ref={detailScrollRef} className="flex-1 overflow-y-auto p-5 lg:p-6 space-y-6">
                {!selectedItem && (
                    <div
                        className="h-full flex items-center justify-center text-sm"
                        style={{
                            borderRadius: 'var(--radius-ctrl)',
                            border: '1px dashed var(--ds-border)',
                            background: 'var(--ds-bg)',
                            color: 'var(--ds-text-secondary)',
                        }}
                    >
                        {t('chatHistory.selectPrompt')}
                    </div>
                )}

                {selectedItem && (
                    <DetailConversation
                        selectedItem={selectedItem}
                        t={t}
                        viewMode={viewMode}
                        detailScrollRef={detailScrollRef}
                        assistantStartRef={assistantStartRef}
                        bottomButtonClassName="absolute right-5 bottom-5"
                        onMessage={onMessage}
                    />
                )}
            </div>
        </div>
    )
}

export function MobileDetailModal({ open, visible, origin, selectedItem, t, lang, viewMode, setViewMode, detailScrollRef, assistantStartRef, onClose }) {
    if (!open || !selectedItem) return null

    return (
        <div
            className={clsx(
                'fixed inset-0 z-50 flex items-center justify-center px-3 py-4 transition-opacity duration-200',
                visible ? 'opacity-100' : 'opacity-0'
            )}
            style={{ background: 'rgba(15, 20, 35, 0.42)', backdropFilter: 'blur(2px)' }}
            onClick={onClose}
        >
            <div
                onClick={(event) => event.stopPropagation()}
                className={clsx(
                    'w-full h-full overflow-hidden flex flex-col transition-transform duration-200 ease-out',
                    visible ? 'scale-100' : 'scale-90'
                )}
                style={{
                    transformOrigin: `${origin.x}% ${origin.y}%`,
                    borderRadius: 'var(--radius-ctrl)',
                    background: 'var(--ds-card)',
                    border: '1px solid var(--ds-border)',
                    boxShadow: 'var(--ds-shadow-lg)',
                }}
            >
                <div
                    className="px-5 py-4 flex items-start justify-between gap-3"
                    style={{ borderBottom: '1px solid var(--ds-border)' }}
                >
                    <div>
                        <div className="text-sm font-semibold" style={{ color: 'var(--ds-text)' }}>{t('chatHistory.detailTitle')}</div>
                        <div className="text-xs mt-1" style={{ color: 'var(--ds-text-secondary)' }}>
                            {formatDateTime(selectedItem.completed_at || selectedItem.updated_at || selectedItem.created_at, lang)}
                        </div>
                    </div>
                    <div className="flex items-center gap-2">
                        <ViewModeToggle t={t} viewMode={viewMode} setViewMode={setViewMode} mobile />
                        <button
                            type="button"
                            onClick={() => detailScrollRef.current?.scrollTo({ top: 0, behavior: 'smooth' })}
                            className="ds-btn-secondary h-9 w-9 p-0"
                            title={t('chatHistory.backToTop')}
                        >
                            <ArrowUp className="w-4 h-4" />
                        </button>
                        <button
                            type="button"
                            onClick={onClose}
                            className="ds-btn-secondary h-9 w-9 p-0"
                            title={t('actions.cancel')}
                        >
                            <X className="w-4 h-4" />
                        </button>
                    </div>
                </div>

                <div ref={detailScrollRef} className="flex-1 overflow-y-auto p-5 space-y-6">
                    <DetailConversation
                        selectedItem={selectedItem}
                        t={t}
                        viewMode={viewMode}
                        detailScrollRef={detailScrollRef}
                        assistantStartRef={assistantStartRef}
                        bottomButtonClassName="fixed right-5 bottom-5"
                    />
                </div>
            </div>
        </div>
    )
}

export function ConfirmClearDialog({ open, t, onCancel, onConfirm }) {
    return (
        <ConfirmDialog
            open={open}
            title={t('chatHistory.confirmClearTitle')}
            message={t('chatHistory.confirmClearDesc')}
            confirmLabel={t('chatHistory.confirmClearAction')}
            cancelLabel={t('actions.cancel')}
            onConfirm={onConfirm}
            onCancel={onCancel}
        />
    )
}