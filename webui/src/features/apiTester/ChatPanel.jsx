import { useState, useRef, useEffect, useCallback } from 'react'
import { Send, Trash2, ChevronDown, ChevronUp, Loader2, Copy, Check, Maximize2, Minimize2, Square } from 'lucide-react'
import clsx from 'clsx'

function StatusLabel({ message }) {
    return (
        <div className="px-3 py-2 text-xs font-medium" style={{ color: 'var(--ds-text-secondary)', background: 'var(--ds-surface)', borderRadius: 'var(--radius-ctrl)' }}>
            {message}
        </div>
    )
}

export default function ChatPanel({
    t,
    message,
    setMessage,
    model,
    response,
    isStreaming,
    loading,
    streamingThinking,
    streamingContent,
    onRunTest,
    onStopGeneration,
    hasAvailableModel,
}) {
    const [collapsed, setCollapsed] = useState(false)
    const [copied, setCopied] = useState(false)
    const [expanded, setExpanded] = useState(false)
    const replyRef = useRef(null)

    const displayContent = isStreaming ? streamingContent : (response?.choices?.[0]?.message?.content || '')
    const displayThinking = isStreaming ? streamingThinking : (response?.choices?.[0]?.message?.reasoning_content || '')
    const hasResponse = !!displayContent || !!displayThinking
    const isError = response && !response.success
    const errorMessage = response?.error || ''

    useEffect(() => {
        if (replyRef.current) {
            replyRef.current.scrollTop = replyRef.current.scrollHeight
        }
    }, [displayContent, displayThinking])

    const handleCopy = useCallback(async () => {
        const text = displayContent || displayThinking
        if (!text) return
        try {
            await navigator.clipboard.writeText(text)
            setCopied(true)
            setTimeout(() => setCopied(false), 2000)
        } catch { /* ignore */ }
    }, [displayContent, displayThinking])

    const handleSubmit = (e) => {
        e?.preventDefault()
        if (!message.trim() || loading || !hasAvailableModel) return
        onRunTest?.()
    }

    const handleCancel = () => {
        onStopGeneration?.()
    }

    const handleClear = () => {
        setMessage('')
    }

    const handleKeyDown = (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault()
            handleSubmit(e)
        }
    }

    return (
        <div className="lg:col-span-9 flex flex-col min-h-0">
            <div className="border flex flex-col h-full min-h-0" style={{ borderColor: 'var(--ds-border)', borderRadius: 'var(--radius-card)', background: 'var(--ds-card)' }}>
                <button
                    onClick={() => setCollapsed(!collapsed)}
                    className="w-full flex items-center justify-between px-5 py-3.5 transition-colors shrink-0"
                    style={{ borderBottom: collapsed ? 'none' : '1px solid var(--ds-border)' }}
                >
                    <div className="flex items-center gap-2">
                        <Send className="w-4 h-4" style={{ color: 'var(--ds-blue)' }} />
                        <span className="text-sm font-bold" style={{ color: 'var(--ds-text)' }}>{t('chat.title')}</span>
                        {model && (
                            <span className="text-[10px] px-2 py-0.5 font-mono" style={{ color: 'var(--ds-text-tertiary)', background: 'var(--ds-bg)', borderRadius: 'var(--radius-ctrl)' }}>
                                {model}
                            </span>
                        )}
                    </div>
                    {collapsed ? <ChevronDown className="w-4 h-4" style={{ color: 'var(--ds-text-tertiary)' }} /> : <ChevronUp className="w-4 h-4" style={{ color: 'var(--ds-text-tertiary)' }} />}
                </button>

                {!collapsed && (
                    <div className="flex flex-col flex-1 min-h-0 p-5 gap-4">
                        {/* Response area */}
                        <div className="flex-1 min-h-0 flex flex-col gap-2">
                            {loading && !isStreaming && (
                                <StatusLabel message={t('chat.sending')} />
                            )}
                            {isStreaming && (
                                <StatusLabel message={t('chat.receiving')} />
                            )}

                            {isError && (
                                <div className="p-3 text-sm" style={{ color: 'var(--ds-danger)', background: 'var(--ds-danger-bg)', border: '1px solid var(--ds-danger-border)', borderRadius: 'var(--radius-ctrl)' }}>
                                    {errorMessage}
                                </div>
                            )}

                            {hasResponse && (
                                <div className="flex-1 min-h-0 flex flex-col">
                                    <div className="flex items-center justify-between mb-2 shrink-0">
                                        <span className="text-[10px] font-semibold uppercase tracking-wider" style={{ color: 'var(--ds-text-tertiary)' }}>{t('chat.response')}</span>
                                        <div className="flex items-center gap-1">
                                            <button
                                                onClick={handleCopy}
                                                className="ds-action-btn p-1 rounded"
                                                style={{ borderRadius: 'var(--radius-ctrl)' }}
                                                title={copied ? t('chat.copied') : t('chat.copy')}
                                            >
                                                {copied ? <Check className="w-3.5 h-3.5" /> : <Copy className="w-3.5 h-3.5" />}
                                            </button>
                                            <button
                                                onClick={() => setExpanded(!expanded)}
                                                className="ds-action-btn p-1 rounded"
                                                style={{ borderRadius: 'var(--radius-ctrl)' }}
                                                title={expanded ? t('chat.collapse') : t('chat.expand')}
                                            >
                                                {expanded ? <Minimize2 className="w-3.5 h-3.5" /> : <Maximize2 className="w-3.5 h-3.5" />}
                                            </button>
                                        </div>
                                    </div>

                                    {displayThinking && (
                                        <details className="mb-3" open>
                                            <summary className="text-[10px] font-semibold uppercase tracking-wider cursor-pointer" style={{ color: 'var(--ds-text-tertiary)' }}>
                                                {t('apiTester.reasoningTrace')}
                                            </summary>
                                            <div
                                                className="mt-2 p-3 text-xs leading-relaxed whitespace-pre-wrap overflow-auto custom-scrollbar max-h-[200px]"
                                                style={{
                                                    background: 'var(--ds-bg)',
                                                    border: '1px solid var(--ds-border)',
                                                    borderRadius: 'var(--radius-ctrl)',
                                                    color: 'var(--ds-text-secondary)',
                                                    fontStyle: 'italic',
                                                }}
                                            >
                                                {displayThinking}
                                            </div>
                                        </details>
                                    )}

                                    <div
                                        ref={replyRef}
                                        className={clsx(
                                            "p-4 text-sm leading-relaxed whitespace-pre-wrap overflow-auto custom-scrollbar flex-1 min-h-0",
                                            expanded ? "max-h-[600px]" : "max-h-[300px]"
                                        )}
                                        style={{
                                            background: 'var(--ds-bg)',
                                            border: '1px solid var(--ds-border)',
                                            borderRadius: 'var(--radius-ctrl)',
                                            color: 'var(--ds-text)',
                                        }}
                                    >
                                        {displayContent || (
                                            <span style={{ color: 'var(--ds-text-tertiary)' }}>
                                                {isStreaming ? t('apiTester.generating') : ''}
                                            </span>
                                        )}
                                    </div>
                                </div>
                            )}

                            {!hasResponse && !loading && (
                                <div className="flex-1 flex items-center justify-center">
                                    <span className="text-sm" style={{ color: 'var(--ds-text-tertiary)' }}>
                                        {hasAvailableModel ? t('chat.placeholder') : t('apiTester.noModelsMessagePlaceholder')}
                                    </span>
                                </div>
                            )}
                        </div>

                        {/* Input area */}
                        <form onSubmit={handleSubmit} className="shrink-0 space-y-3">
                            <textarea
                                className="w-full p-3.5 text-sm resize-none"
                                rows={3}
                                style={{
                                    background: 'var(--ds-bg)',
                                    border: '1px solid var(--ds-border)',
                                    borderRadius: 'var(--radius-ctrl)',
                                    color: 'var(--ds-text)',
                                }}
                                placeholder={t('apiTester.enterMessage')}
                                value={message}
                                onChange={e => setMessage(e.target.value)}
                                onKeyDown={handleKeyDown}
                                disabled={!hasAvailableModel}
                            />

                            <div className="flex items-center justify-between gap-2">
                                <div className="flex items-center gap-2">
                                    <button
                                        type="button"
                                        onClick={handleClear}
                                        className="ds-btn-secondary text-[10px] px-2 py-1"
                                    >
                                        <Trash2 className="w-3 h-3 mr-1" />
                                        {t('chat.clear')}
                                    </button>
                                </div>
                                <div className="flex items-center gap-2">
                                    {loading ? (
                                        <button
                                            type="button"
                                            onClick={handleCancel}
                                            className="ds-btn-danger text-[10px] px-3 py-1.5"
                                        >
                                            <Square className="w-3 h-3 mr-1" />
                                            {t('chat.cancel')}
                                        </button>
                                    ) : null}
                                    <button
                                        type="submit"
                                        disabled={loading || !message.trim() || !hasAvailableModel}
                                        className="ds-btn-primary text-[11px] px-3 py-1.5"
                                    >
                                        {loading ? (
                                            <Loader2 className="w-3.5 h-3.5 animate-spin" />
                                        ) : (
                                            <>
                                                <Send className="w-3.5 h-3.5 mr-1.5" />
                                                {t('chat.send')}
                                            </>
                                        )}
                                    </button>
                                </div>
                            </div>
                        </form>
                    </div>
                )}
            </div>
        </div>
    )
}