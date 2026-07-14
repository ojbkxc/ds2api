import { Loader2, RefreshCcw, Trash2 } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'

import { useI18n } from '../../i18n'
import Button from '../../components/ui/Button'
import SegmentedControl from '../../components/ui/SegmentedControl'
import { ChatHistoryListPane, ConfirmClearDialog, DesktopDetailPane, MobileDetailModal } from './ChatHistoryPanels'
import {
    DISABLED_LIMIT,
    LIMIT_OPTIONS,
    VIEW_MODE_KEY,
} from './chatHistoryUtils'

const LIST_REFRESH_MS = 1500
const STREAMING_DETAIL_REFRESH_MS = 750

export default function ChatHistoryContainer({ authFetch, onMessage }) {
    const { t, lang } = useI18n()
    const apiFetch = authFetch || fetch
    const [items, setItems] = useState([])
    const [limit, setLimit] = useState(20)
    const [loading, setLoading] = useState(true)
    const [refreshing, setRefreshing] = useState(false)
    const [selectedId, setSelectedId] = useState('')
    const [selectedDetail, setSelectedDetail] = useState(null)
    const [savingLimit, setSavingLimit] = useState(false)
    const [clearing, setClearing] = useState(false)
    const [deletingId, setDeletingId] = useState('')
    const [detail, setDetail] = useState('')
    const [confirmClearOpen, setConfirmClearOpen] = useState(false)
    const [autoRefreshReady, setAutoRefreshReady] = useState(false)
    const [viewMode, setViewMode] = useState(() => {
        if (typeof localStorage === 'undefined') return 'list'
        const stored = localStorage.getItem(VIEW_MODE_KEY)
        return stored === 'merged' ? 'merged' : 'list'
    })
    const [isMobileView, setIsMobileView] = useState(() => typeof window !== 'undefined' ? window.innerWidth < 1024 : false)
    const [mobileDetailOpen, setMobileDetailOpen] = useState(false)
    const [mobileDetailVisible, setMobileDetailVisible] = useState(false)
    const [mobileOrigin, setMobileOrigin] = useState({ x: 50, y: 50 })
    const [pendingJumpToAssistant, setPendingJumpToAssistant] = useState(false)

    const inFlightRef = useRef(false)
    const detailInFlightRef = useRef(false)
    const listETagRef = useRef('')
    const detailETagRef = useRef('')
    const assistantStartRef = useRef(null)
    const detailScrollRef = useRef(null)
    const mobileCloseTimerRef = useRef(null)

    const selectedSummary = items.find(item => item.id === selectedId) || items[0] || null
    const selectedItem = selectedDetail && selectedDetail.id === selectedId ? selectedDetail : null

    const syncItems = (nextItems) => {
        setItems(nextItems)
        setSelectedId(prev => {
            if (!nextItems.length) return ''
            if (prev && nextItems.some(item => item.id === prev)) return prev
            return nextItems[0].id
        })
    }

    const loadList = async ({ mode = 'silent', announceError = false } = {}) => {
        if (inFlightRef.current) return
        inFlightRef.current = true
        if (mode === 'manual') {
            setRefreshing(true)
        } else if (mode === 'initial') {
            setLoading(true)
        }
        if (announceError) {
            setDetail('')
        }
        try {
            const headers = {}
            if (listETagRef.current) {
                headers['If-None-Match'] = listETagRef.current
            }
            const res = await apiFetch('/admin/chat-history', { headers })
            if (res.status === 304) {
                return
            }
            const data = await res.json()
            if (!res.ok) {
                throw new Error(data?.detail || t('chatHistory.loadFailed'))
            }
            listETagRef.current = res.headers.get('ETag') || ''
            setLimit(typeof data.limit === 'number' ? data.limit : 20)
            syncItems(Array.isArray(data.items) ? data.items : [])
        } catch (error) {
            setDetail(error.message || t('chatHistory.loadFailed'))
            if (announceError) {
                onMessage?.('error', error.message || t('chatHistory.loadFailed'))
            }
        } finally {
            if (mode === 'initial') {
                setLoading(false)
            }
            if (mode === 'manual') {
                setRefreshing(false)
            }
            inFlightRef.current = false
        }
    }

    const loadDetail = async (id, { announceError = false } = {}) => {
        if (!id || detailInFlightRef.current) return
        detailInFlightRef.current = true
        try {
            const headers = {}
            if (detailETagRef.current) {
                headers['If-None-Match'] = detailETagRef.current
            }
            const res = await apiFetch(`/admin/chat-history/${encodeURIComponent(id)}`, { headers })
            if (res.status === 304) {
                return
            }
            const data = await res.json()
            if (!res.ok) {
                throw new Error(data?.detail || t('chatHistory.loadFailed'))
            }
            detailETagRef.current = res.headers.get('ETag') || ''
            setSelectedDetail(data.item || null)
        } catch (error) {
            if (announceError) {
                onMessage?.('error', error.message || t('chatHistory.loadFailed'))
            }
        } finally {
            detailInFlightRef.current = false
        }
    }

    useEffect(() => {
        loadList({ mode: 'initial', announceError: true }).finally(() => {
            setAutoRefreshReady(true)
        })
    }, [])

    useEffect(() => {
        if (!autoRefreshReady || limit === DISABLED_LIMIT) return undefined
        const timer = window.setInterval(() => {
            loadList({ mode: 'silent', announceError: false })
        }, LIST_REFRESH_MS)
        return () => window.clearInterval(timer)
    }, [autoRefreshReady, limit])

    useEffect(() => {
        if (!autoRefreshReady || !selectedId || selectedSummary?.status !== 'streaming') return undefined
        const timer = window.setInterval(() => {
            loadDetail(selectedId, { announceError: false })
        }, STREAMING_DETAIL_REFRESH_MS)
        return () => window.clearInterval(timer)
    }, [autoRefreshReady, selectedId, selectedSummary?.status])

    useEffect(() => {
        if (!selectedId) return undefined
        detailETagRef.current = ''
        setSelectedDetail(null)
        loadDetail(selectedId, { announceError: false })
    }, [selectedId, mobileDetailOpen])

    useEffect(() => {
        if (!pendingJumpToAssistant || !selectedItem || selectedItem.id !== selectedId) return undefined
        const frame = window.requestAnimationFrame(() => {
            assistantStartRef.current?.scrollIntoView({ behavior: 'auto', block: 'start' })
            setPendingJumpToAssistant(false)
        })
        return () => window.cancelAnimationFrame(frame)
    }, [pendingJumpToAssistant, selectedId, selectedItem?.id, selectedItem?.revision, mobileDetailOpen, viewMode])

    useEffect(() => {
        if (typeof localStorage === 'undefined') return
        localStorage.setItem(VIEW_MODE_KEY, viewMode)
    }, [viewMode])

    useEffect(() => {
        if (typeof window === 'undefined') return undefined
        const handleResize = () => setIsMobileView(window.innerWidth < 1024)
        handleResize()
        window.addEventListener('resize', handleResize)
        return () => window.removeEventListener('resize', handleResize)
    }, [])

    useEffect(() => {
        if (!isMobileView) {
            setMobileDetailOpen(false)
            setMobileDetailVisible(false)
        }
    }, [isMobileView])

    useEffect(() => {
        return () => {
            if (mobileCloseTimerRef.current) {
                window.clearTimeout(mobileCloseTimerRef.current)
            }
        }
    }, [])

    const handleRefresh = async ({ manual = true } = {}) => {
        await loadList({ mode: manual ? 'manual' : 'silent', announceError: manual })
        if (selectedId) {
            detailETagRef.current = ''
            await loadDetail(selectedId, { announceError: manual })
        }
    }

    const handleLimitChange = async (nextLimit) => {
        if (nextLimit === limit || savingLimit) return
        setSavingLimit(true)
        try {
            const res = await apiFetch('/admin/chat-history/settings', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ limit: nextLimit }),
            })
            const data = await res.json()
            if (!res.ok) {
                throw new Error(data?.detail || t('chatHistory.updateLimitFailed'))
            }
            const resolvedLimit = typeof data.limit === 'number' ? data.limit : nextLimit
            setLimit(resolvedLimit)
            listETagRef.current = ''
            syncItems(Array.isArray(data.items) ? data.items : [])
            onMessage?.(
                'success',
                resolvedLimit === DISABLED_LIMIT
                    ? t('chatHistory.disabledSuccess')
                    : t('chatHistory.limitUpdated', { limit: resolvedLimit })
            )
        } catch (error) {
            onMessage?.('error', error.message || t('chatHistory.updateLimitFailed'))
        } finally {
            setSavingLimit(false)
        }
    }

    const handleDeleteItem = async (id) => {
        if (!id || deletingId) return
        setDeletingId(id)
        try {
            const res = await apiFetch(`/admin/chat-history/${encodeURIComponent(id)}`, { method: 'DELETE' })
            const data = await res.json()
            if (!res.ok) {
                throw new Error(data?.detail || t('chatHistory.deleteFailed'))
            }
            if (selectedId === id) {
                detailETagRef.current = ''
                setSelectedDetail(null)
            }
            syncItems(items.filter(item => item.id !== id))
            onMessage?.('success', t('chatHistory.deleteSuccess'))
        } catch (error) {
            onMessage?.('error', error.message || t('chatHistory.deleteFailed'))
        } finally {
            setDeletingId('')
        }
    }

    const handleClear = async () => {
        if (clearing || !items.length) return
        setClearing(true)
        try {
            const res = await apiFetch('/admin/chat-history', { method: 'DELETE' })
            const data = await res.json()
            if (!res.ok) {
                throw new Error(data?.detail || t('chatHistory.clearFailed'))
            }
            listETagRef.current = ''
            detailETagRef.current = ''
            setSelectedDetail(null)
            syncItems([])
            onMessage?.('success', t('chatHistory.clearSuccess'))
        } catch (error) {
            onMessage?.('error', error.message || t('chatHistory.clearFailed'))
        } finally {
            setClearing(false)
        }
    }

    const openMobileDetail = (itemId, event) => {
        const x = typeof window !== 'undefined' && event?.clientX ? (event.clientX / window.innerWidth) * 100 : 50
        const y = typeof window !== 'undefined' && event?.clientY ? (event.clientY / window.innerHeight) * 100 : 50
        setMobileOrigin({ x, y })
        setPendingJumpToAssistant(true)
        setSelectedId(itemId)
        setMobileDetailOpen(true)
        setMobileDetailVisible(false)
        window.requestAnimationFrame(() => {
            window.requestAnimationFrame(() => setMobileDetailVisible(true))
        })
    }

    const closeMobileDetail = () => {
        setMobileDetailVisible(false)
        if (mobileCloseTimerRef.current) {
            window.clearTimeout(mobileCloseTimerRef.current)
        }
        mobileCloseTimerRef.current = window.setTimeout(() => {
            setMobileDetailOpen(false)
        }, 180)
    }

    const handleSelectItem = (itemId, event) => {
        if (isMobileView) {
            openMobileDetail(itemId, event)
            return
        }
        if (itemId === selectedId) {
            detailETagRef.current = ''
            setSelectedDetail(null)
            loadDetail(itemId, { announceError: false })
            return
        }
        setPendingJumpToAssistant(true)
        setSelectedId(itemId)
    }

    if (loading) {
        return (
            <div
                className="h-[calc(100vh-140px)] flex items-center justify-center"
                style={{
                    background: 'var(--ds-card)',
                    border: '1px solid var(--ds-border)',
                    borderRadius: 'var(--radius-card)',
                }}
            >
                <div className="flex items-center gap-3 text-sm" style={{ color: 'var(--ds-text-secondary)' }}>
                    <Loader2 className="w-4 h-4 animate-spin" />
                    {t('chatHistory.loading')}
                </div>
            </div>
        )
    }

    return (
        <div className="space-y-6">
            <div
                className="p-4 lg:p-5 flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between"
                style={{
                    background: 'var(--ds-card)',
                    border: '1px solid var(--ds-border)',
                    borderRadius: 'var(--radius-card)',
                }}
            >
                <div>
                    <div className="text-sm font-semibold" style={{ color: 'var(--ds-text)' }}>
                        {t('chatHistory.retentionTitle')}
                    </div>
                    <div className="text-xs mt-1" style={{ color: 'var(--ds-text-secondary)' }}>
                        {t('chatHistory.retentionDesc')}
                    </div>
                </div>
                <div className="flex flex-wrap gap-2 items-center">
                    {LIMIT_OPTIONS.map(option => {
                        const active = option === limit
                        const isDestructive = option === DISABLED_LIMIT
                        return (
                            <button
                                key={option}
                                type="button"
                                disabled={savingLimit}
                                onClick={() => handleLimitChange(option)}
                                className="h-9 px-3 text-sm font-medium transition-colors"
                                style={{
                                    borderRadius: 'var(--radius-ctrl)',
                                    border: active
                                        ? (isDestructive ? '1px solid var(--ds-danger)' : '1px solid var(--ds-blue)')
                                        : '1px solid var(--ds-border)',
                                    background: active
                                        ? (isDestructive ? 'var(--ds-danger-bg)' : 'var(--ds-blue)')
                                        : 'transparent',
                                    color: active
                                        ? (isDestructive ? 'var(--ds-danger)' : 'var(--ds-text-on-primary)')
                                        : 'var(--ds-text-secondary)',
                                    cursor: savingLimit ? 'not-allowed' : 'pointer',
                                    opacity: savingLimit ? 0.5 : 1,
                                }}
                            >
                                {option === DISABLED_LIMIT ? t('chatHistory.off') : option}
                            </button>
                        )
                    })}
                    <button
                        type="button"
                        onClick={() => handleRefresh({ manual: true })}
                        disabled={refreshing}
                        className="h-9 flex items-center justify-center transition-colors"
                        style={{
                            borderRadius: 'var(--radius-ctrl)',
                            border: '1px solid var(--ds-border)',
                            background: 'transparent',
                            color: 'var(--ds-text-secondary)',
                            cursor: refreshing ? 'not-allowed' : 'pointer',
                            padding: isMobileView ? '0' : '0 0.75rem',
                            width: isMobileView ? '36px' : 'auto',
                            gap: isMobileView ? '0' : '0.5rem',
                        }}
                        onMouseEnter={(e) => {
                            if (!refreshing) {
                                e.currentTarget.style.color = 'var(--ds-text)'
                                e.currentTarget.style.background = 'var(--ds-surface-hover)'
                            }
                        }}
                        onMouseLeave={(e) => {
                            e.currentTarget.style.color = 'var(--ds-text-secondary)'
                            e.currentTarget.style.background = 'transparent'
                        }}
                    >
                        {refreshing ? <Loader2 className="w-4 h-4 animate-spin" /> : <RefreshCcw className="w-4 h-4" />}
                        {!isMobileView && t('chatHistory.refresh')}
                    </button>
                    <button
                        type="button"
                        onClick={() => setConfirmClearOpen(true)}
                        disabled={clearing || !items.length}
                        className="h-10 w-10 flex items-center justify-center transition-colors"
                        style={{
                            borderRadius: 'var(--radius-ctrl)',
                            border: '1px solid var(--ds-border)',
                            background: 'var(--ds-bg)',
                            color: 'var(--ds-text-tertiary)',
                            cursor: (clearing || !items.length) ? 'not-allowed' : 'pointer',
                            opacity: (clearing || !items.length) ? 0.5 : 1,
                        }}
                        onMouseEnter={(e) => {
                            if (!clearing && items.length) {
                                e.currentTarget.style.color = 'var(--ds-danger)'
                                e.currentTarget.style.background = 'var(--ds-surface-hover)'
                            }
                        }}
                        onMouseLeave={(e) => {
                            e.currentTarget.style.color = 'var(--ds-text-tertiary)'
                            e.currentTarget.style.background = 'var(--ds-bg)'
                        }}
                        title={t('chatHistory.clearAll')}
                    >
                        {clearing ? <Loader2 className="w-4 h-4 animate-spin" /> : <Trash2 className="w-4 h-4" />}
                    </button>
                </div>
            </div>

            {detail && (
                <div
                    className="px-4 py-3 text-sm"
                    style={{
                        borderRadius: 'var(--radius-ctrl)',
                        border: '1px solid var(--ds-danger-border)',
                        background: 'var(--ds-danger-bg)',
                        color: 'var(--ds-danger)',
                    }}
                >
                    {detail}
                </div>
            )}

            <div className="grid grid-cols-1 lg:grid-cols-[340px,minmax(0,1fr)] gap-6 h-[calc(100vh-240px)] min-h-[520px]">
                <ChatHistoryListPane
                    items={items}
                    selectedItem={selectedItem}
                    deletingId={deletingId}
                    t={t}
                    lang={lang}
                    onSelectItem={handleSelectItem}
                    onDeleteItem={handleDeleteItem}
                />

                <DesktopDetailPane
                    selectedSummary={selectedSummary}
                    selectedItem={selectedItem}
                    t={t}
                    lang={lang}
                    viewMode={viewMode}
                    setViewMode={setViewMode}
                    detailScrollRef={detailScrollRef}
                    assistantStartRef={assistantStartRef}
                    onMessage={onMessage}
                />
            </div>

            <MobileDetailModal
                open={isMobileView && mobileDetailOpen}
                visible={mobileDetailVisible}
                origin={mobileOrigin}
                selectedItem={selectedItem}
                t={t}
                lang={lang}
                viewMode={viewMode}
                setViewMode={setViewMode}
                detailScrollRef={detailScrollRef}
                assistantStartRef={assistantStartRef}
                onClose={closeMobileDetail}
            />

            <ConfirmClearDialog
                open={confirmClearOpen}
                t={t}
                onCancel={() => setConfirmClearOpen(false)}
                onConfirm={async () => {
                    setConfirmClearOpen(false)
                    await handleClear()
                }}
            />
        </div>
    )
}