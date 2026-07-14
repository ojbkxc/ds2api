import { ArrowDown, Bot, ChevronDown, Clock3, Copy, Download, Sparkles, UserRound } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import clsx from 'clsx'

import {
    MESSAGE_COLLAPSE_AT,
    buildListModeMessages,
    copyTextWithFallback,
    downloadTextFile,
    formatElapsed,
} from './chatHistoryUtils'

function ExpandableText({ text = '', threshold = MESSAGE_COLLAPSE_AT, expandLabel, collapseLabel, buttonClassName }) {
    const shouldCollapse = text.length > threshold
    const [expanded, setExpanded] = useState(false)
    const contentRef = useRef(null)
    const [maxHeight, setMaxHeight] = useState('none')

    useEffect(() => {
        setExpanded(false)
    }, [text])

    const visibleText = shouldCollapse && !expanded ? `${text.slice(0, threshold)}...` : text

    useEffect(() => {
        if (!contentRef.current) return
        setMaxHeight(`${contentRef.current.scrollHeight}px`)
    }, [expanded, visibleText])

    return (
        <div>
            <div className="overflow-hidden" style={{ maxHeight, transition: 'max-height 0.3s ease-out' }}>
                <div ref={contentRef} className="whitespace-pre-wrap break-words">
                    {visibleText}
                </div>
            </div>
            {shouldCollapse && (
                <button
                    type="button"
                    onClick={() => setExpanded(prev => !prev)}
                    className={clsx('mt-3 inline-flex items-center gap-2 text-xs font-medium transition-colors', buttonClassName)}
                    style={{ borderRadius: 'var(--radius-ctrl)', color: 'var(--ds-text-secondary)' }}
                >
                    <ChevronDown className={clsx('w-3.5 h-3.5 transition-transform duration-300', expanded && 'rotate-180')} />
                    {expanded ? collapseLabel : expandLabel}
                </button>
            )}
        </div>
    )
}

function RequestMessages({ item, t, messages }) {
    const requestMessages = Array.isArray(messages) && messages.length > 0
        ? messages
        : [{ role: 'user', content: item?.user_input || t('chatHistory.emptyUserInput') }]

    return (
        <div className="space-y-5 max-w-4xl mx-auto">
            {requestMessages.map((message, index) => {
                const role = message.role || 'user'
                const isUser = role === 'user'
                const isAssistant = role === 'assistant'
                const isTool = role === 'tool'
                const label = isUser
                    ? t('chatHistory.role.user')
                    : (isAssistant ? t('chatHistory.role.assistant') : (isTool ? t('chatHistory.role.tool') : t('chatHistory.role.system')))
                return (
                    <div key={`${role}-${index}`} className={clsx('flex gap-4', isUser && 'flex-row-reverse justify-start')}>
                        <div
                            className="w-8 h-8 flex items-center justify-center shrink-0"
                            style={{
                                borderRadius: 'var(--radius-ctrl)',
                                background: isUser ? 'var(--ds-surface)' : (isAssistant ? 'var(--ds-surface)' : 'var(--ds-bg)'),
                                border: '1px solid var(--ds-border)',
                            }}
                        >
                            {isUser
                                ? <UserRound className="w-4 h-4" style={{ color: 'var(--ds-text-secondary)' }} />
                                : <Bot className="w-4 h-4" style={{ color: 'var(--ds-text)' }} />}
                        </div>
                        <div className="max-w-[88%] lg:max-w-[78%] text-left">
                            <div
                                className={clsx('text-[11px] uppercase tracking-[0.12em] mb-2 px-1', isUser && 'text-right')}
                                style={{ color: 'var(--ds-text-secondary)' }}
                            >
                                {label}
                            </div>
                            <div
                                className="px-5 py-3 text-sm leading-relaxed whitespace-pre-wrap break-words"
                                style={{
                                    borderRadius: isUser
                                        ? 'var(--radius-ctrl) var(--radius-ctrl) 2px var(--radius-ctrl)'
                                        : 'var(--radius-ctrl) var(--radius-ctrl) var(--radius-ctrl) 2px',
                                    background: isUser ? 'var(--ds-blue)' : (isAssistant ? 'var(--ds-surface)' : 'var(--ds-bg)'),
                                    color: isUser ? 'var(--ds-text-on-primary)' : 'var(--ds-text)',
                                    border: isUser ? '1px solid var(--ds-blue)' : '1px solid var(--ds-border)',
                                    boxShadow: isUser ? 'var(--ds-elevate-1)' : 'none',
                                }}
                            >
                                <div className="whitespace-pre-wrap break-words">
                                    {message.content || t('chatHistory.emptyUserInput')}
                                </div>
                            </div>
                        </div>
                    </div>
                )
            })}
        </div>
    )
}

function PromptTextActions({ text, filename, copyTitle, downloadTitle, t, onMessage, buttonClassName }) {
    const handleCopy = async () => {
        try {
            await copyTextWithFallback(text)
            onMessage?.('success', t('chatHistory.copySuccess'))
        } catch {
            onMessage?.('error', t('chatHistory.copyFailed'))
        }
    }

    const handleDownload = () => {
        try {
            downloadTextFile(filename, text)
            onMessage?.('success', t('chatHistory.downloadSuccess'))
        } catch {
            onMessage?.('error', t('chatHistory.downloadFailed'))
        }
    }

    return (
        <div className="flex items-center gap-2">
            <button
                type="button"
                onClick={handleCopy}
                className={clsx('ds-action-btn h-8 w-8', buttonClassName)}
                style={{ borderRadius: 'var(--radius-ctrl)' }}
                title={copyTitle}
            >
                <Copy className="w-4 h-4" />
            </button>
            <button
                type="button"
                onClick={handleDownload}
                className={clsx('ds-action-btn h-8 w-8', buttonClassName)}
                style={{ borderRadius: 'var(--radius-ctrl)' }}
                title={downloadTitle}
            >
                <Download className="w-4 h-4" />
            </button>
        </div>
    )
}

function MergedPromptView({ item, t, onMessage }) {
    const merged = item?.final_prompt || ''

    return (
        <div
            className="max-w-4xl mx-auto px-5 py-4"
            style={{
                background: 'var(--ds-warning-bg)',
                border: '1px solid var(--ds-warning-border)',
                borderRadius: 'var(--radius-ctrl)',
            }}
        >
            <div className="mb-3 flex items-center justify-between gap-3">
                <div className="text-[11px] uppercase tracking-[0.12em]" style={{ color: 'var(--ds-warning)' }}>
                    {t('chatHistory.mergedInput')}
                </div>
                <PromptTextActions
                    text={merged}
                    filename={`Merged_${item?.id || 'prompt'}.txt`}
                    copyTitle={t('chatHistory.copyMerged')}
                    downloadTitle={t('chatHistory.downloadMerged')}
                    t={t}
                    onMessage={onMessage}
                    buttonClassName=""
                />
            </div>
            <div className="text-sm leading-7 whitespace-pre-wrap break-words font-mono" style={{ color: 'var(--ds-text)' }}>
                <ExpandableText
                    text={merged || t('chatHistory.emptyMergedPrompt')}
                    expandLabel={t('chatHistory.expand')}
                    collapseLabel={t('chatHistory.collapse')}
                    buttonClassName=""
                />
            </div>
        </div>
    )
}

function HistoryTextView({ item, t, onMessage }) {
    const historyText = (item?.history_text || '').trim()
    if (!historyText) return null

    return (
        <div className="max-w-4xl mx-auto ds-card px-5 py-4">
            <div className="mb-3 flex items-center justify-between gap-3">
                <div className="text-[11px] uppercase tracking-[0.12em] text-left" style={{ color: 'var(--ds-text-secondary)' }}>
                    HISTORY
                </div>
                <PromptTextActions
                    text={historyText}
                    filename={`History_${item?.id || 'history'}.txt`}
                    copyTitle={t('chatHistory.copyHistory')}
                    downloadTitle={t('chatHistory.downloadHistory')}
                    t={t}
                    onMessage={onMessage}
                    buttonClassName=""
                />
            </div>
            <div className="text-sm leading-7 whitespace-pre-wrap break-words font-mono" style={{ color: 'var(--ds-text)' }}>
                <ExpandableText
                    text={historyText}
                    threshold={Math.floor(MESSAGE_COLLAPSE_AT / 4)}
                    expandLabel={t('chatHistory.expand')}
                    collapseLabel={t('chatHistory.collapse')}
                    buttonClassName=""
                />
            </div>
        </div>
    )
}

function MetaGrid({ selectedItem, t }) {
    return (
        <div className="max-w-4xl mx-auto ds-card p-4 space-y-3">
            <div className="text-xs font-semibold uppercase tracking-[0.12em]" style={{ color: 'var(--ds-text-secondary)' }}>
                {t('chatHistory.metaTitle')}
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
                <div className="ds-surface-panel px-3 py-2">
                    <div className="text-[11px]" style={{ color: 'var(--ds-text-secondary)' }}>{t('chatHistory.metaAccount')}</div>
                    <div className="text-sm font-medium" style={{ color: 'var(--ds-text)' }}>{selectedItem.account_id || t('chatHistory.metaUnknown')}</div>
                </div>
                <div className="ds-surface-panel px-3 py-2">
                    <div className="text-[11px]" style={{ color: 'var(--ds-text-secondary)' }}>{t('chatHistory.metaElapsed')}</div>
                    <div className="text-sm font-medium flex items-center gap-2" style={{ color: 'var(--ds-text)' }}>
                        <Clock3 className="w-3.5 h-3.5" style={{ color: 'var(--ds-text-secondary)' }} />
                        {formatElapsed(selectedItem.elapsed_ms, t)}
                    </div>
                </div>
                <div className="ds-surface-panel px-3 py-2">
                    <div className="text-[11px]" style={{ color: 'var(--ds-text-secondary)' }}>{t('chatHistory.metaSurface')}</div>
                    <div className="text-sm font-medium break-all" style={{ color: 'var(--ds-text)' }}>{selectedItem.surface || t('chatHistory.metaUnknown')}</div>
                </div>
                <div className="ds-surface-panel px-3 py-2">
                    <div className="text-[11px]" style={{ color: 'var(--ds-text-secondary)' }}>{t('chatHistory.metaModel')}</div>
                    <div className="text-sm font-medium break-all" style={{ color: 'var(--ds-text)' }}>{selectedItem.model || t('chatHistory.metaUnknown')}</div>
                </div>
                <div className="ds-surface-panel px-3 py-2">
                    <div className="text-[11px]" style={{ color: 'var(--ds-text-secondary)' }}>{t('chatHistory.metaStatusCode')}</div>
                    <div className="text-sm font-medium" style={{ color: 'var(--ds-text)' }}>{selectedItem.status_code || '-'}</div>
                </div>
                <div className="ds-surface-panel px-3 py-2">
                    <div className="text-[11px]" style={{ color: 'var(--ds-text-secondary)' }}>{t('chatHistory.metaStream')}</div>
                    <div className="text-sm font-medium" style={{ color: 'var(--ds-text)' }}>{selectedItem.stream ? t('chatHistory.streamMode') : t('chatHistory.nonStreamMode')}</div>
                </div>
                <div className="ds-surface-panel px-3 py-2">
                    <div className="text-[11px]" style={{ color: 'var(--ds-text-secondary)' }}>{t('chatHistory.metaCaller')}</div>
                    <div className="text-sm font-medium break-all" style={{ color: 'var(--ds-text)' }}>{selectedItem.caller_id || t('chatHistory.metaUnknown')}</div>
                </div>
            </div>
        </div>
    )
}

export default function DetailConversation({ selectedItem, t, viewMode, detailScrollRef, assistantStartRef, bottomButtonClassName, onMessage }) {
    if (!selectedItem) return null
    const listModeState = viewMode === 'list' ? buildListModeMessages(selectedItem, t) : null
    const showHistoryAtTop = viewMode !== 'list' || !listModeState?.historyMerged

    return (
        <>
            {showHistoryAtTop && <HistoryTextView item={selectedItem} t={t} onMessage={onMessage} />}

            {viewMode === 'list'
                ? <RequestMessages item={selectedItem} t={t} messages={listModeState?.messages} />
                : <MergedPromptView item={selectedItem} t={t} onMessage={onMessage} />}

            <div ref={assistantStartRef} className="flex gap-4 max-w-4xl mx-auto">
                <div
                    className="w-8 h-8 flex items-center justify-center shrink-0"
                    style={{
                        borderRadius: 'var(--radius-ctrl)',
                        background: selectedItem.status === 'error' ? 'var(--ds-danger-bg)' : 'var(--ds-surface)',
                        border: selectedItem.status === 'error' ? '1px solid var(--ds-danger-border)' : '1px solid var(--ds-border)',
                    }}
                >
                    <Bot
                        className="w-4 h-4"
                        style={{ color: selectedItem.status === 'error' ? 'var(--ds-danger)' : 'var(--ds-text)' }}
                    />
                </div>
                <div className="space-y-4 flex-1 min-w-0">
                    {(selectedItem.reasoning_content || '').trim() && (
                        <div
                            className="p-3 space-y-1.5"
                            style={{
                                background: 'var(--ds-surface)',
                                border: '1px solid var(--ds-border)',
                                borderRadius: 'var(--radius-ctrl)',
                            }}
                        >
                            <div className="flex items-center gap-1.5" style={{ color: 'var(--ds-text-secondary)' }}>
                                <Sparkles className="w-3.5 h-3.5" />
                                <span className="font-medium text-xs">{t('chatHistory.reasoningTrace')}</span>
                            </div>
                            <div
                                className="whitespace-pre-wrap leading-relaxed font-mono text-[12px] md:text-[13px] max-h-64 overflow-y-auto custom-scrollbar pl-5 break-words"
                                style={{ color: 'var(--ds-text-secondary)', borderLeft: '2px solid var(--ds-border)' }}
                            >
                                {selectedItem.reasoning_content}
                            </div>
                        </div>
                    )}

                    <div className="text-sm leading-7 whitespace-pre-wrap break-words" style={{ color: 'var(--ds-text)' }}>
                        {selectedItem.status === 'error'
                            ? <span className="font-medium" style={{ color: 'var(--ds-danger)' }}>{selectedItem.error || t('chatHistory.failedOutput')}</span>
                            : (selectedItem.content || t('chatHistory.emptyAssistantOutput'))}
                    </div>
                </div>
            </div>

            <MetaGrid selectedItem={selectedItem} t={t} />

            <button
                type="button"
                onClick={() => detailScrollRef.current?.scrollTo({ top: detailScrollRef.current?.scrollHeight || 0, behavior: 'smooth' })}
                className={clsx('ds-action-btn h-12 w-12', bottomButtonClassName)}
                style={{
                    borderRadius: 'var(--radius-pill)',
                    background: 'var(--ds-card)',
                    border: '1px solid var(--ds-border)',
                    boxShadow: 'var(--ds-shadow-lg)',
                }}
                title={t('chatHistory.backToBottom')}
            >
                <ArrowDown className="w-5 h-5" />
            </button>
        </>
    )
}