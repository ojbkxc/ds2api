import { useState, useRef, useEffect, useCallback, useMemo } from 'react'
import { Send, Trash2, ChevronDown, ChevronUp, Loader2, Copy, Check, Maximize2, Minimize2 } from 'lucide-react'
import clsx from 'clsx'
import { useI18n } from '../../i18n'
import SegmentedControl from '../../components/ui/SegmentedControl'

function StatusLabel({ message }) {
    return (
        <div className="px-3 py-2 text-xs font-medium" style={{ color: 'var(--ds-text-secondary)', background: 'var(--ds-surface)', borderRadius: 'var(--radius-ctrl)' }}>
            {message}
        </div>
    )
}

export default function ChatPanel({ onSend, config, onClear, onMessage }) {
    const { t } = useI18n()
    const [prompt, setPrompt] = useState('')
    const [reply, setReply] = useState('')
    const [loading, setLoading] = useState(false)
    const [collapsed, setCollapsed] = useState(false)
    const [copied, setCopied] = useState(false)
    const [expanded, setExpanded] = useState(false)
    const [status, setStatus] = useState('')
    const [advanceOpen, setAdvanceOpen] = useState(false)
    const [modelMode, setModelMode] = useState(
        config?.default_model || 'deepseek-v4'
    )
    const [stream, setStream] = useState(true)
    const [temperature, setTemperature] = useState(0.7)
    const [maxTokens, setMaxTokens] = useState(4096)
    const replyRef = useRef(null)
    const abortRef = useRef(null)

    const models = useMemo(() => {
        const list = []
        if (config?.models) {
            Object.entries(config.models).forEach(([key, val]) => {
                list.push({ key, label: val.display_name || key })
            })
        }
        if (list.length === 0) {
            list.push({ key: 'deepseek-v4', label: 'deepseek-v4' })
        }
        return list
    }, [config])

    useEffect(() => {
        if (replyRef.current) {
            replyRef.current.scrollTop = replyRef.current.scrollHeight
        }
    }, [reply])

    const handleCopy = useCallback(async () => {
        if (!reply) return
        try {
            await navigator.clipboard.writeText(reply)
            setCopied(true)
            setTimeout(() => setCopied(false), 2000)
        } catch { /* ignore */ }
    }, [reply])

    const handleSubmit = async (e) => {
        e?.preventDefault()
        if (!prompt.trim() || loading) return

        setLoading(true)
        setReply('')
        setStatus(t('chat.sending'))

        const controller = new AbortController()
        abortRef.current = controller

        try {
            const res = await fetch('/v1/chat/completions', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    model: modelMode,
                    messages: [{ role: 'user', content: prompt }],
                    stream,
                    temperature,
                    max_tokens: maxTokens,
                }),
                signal: controller.signal,
            })

            if (!res.ok) {
                const err = await res.json().catch(() => ({}))
                throw new Error(err.detail || err.error?.message || `HTTP ${res.status}`)
            }

            if (stream) {
                setStatus(t('chat.receiving'))
                const reader = res.body.getReader()
                const decoder = new TextDecoder()
                let buffer = ''

                while (true) {
                    const { done, value } = await reader.read()
                    if (done) break
                    buffer += decoder.decode(value, { stream: true })
                    const lines = buffer.split('\n')
                    buffer = lines.pop() || ''
                    for (const line of lines) {
                        if (!line.startsWith('data: ')) continue
                        const data = line.slice(6)
                        if (data === '[DONE]') continue
                        try {
                            const parsed = JSON.parse(data)
                            const content = parsed.choices?.[0]?.delta?.content || ''
                            setReply(prev => prev + content)
                        } catch { /* skip bad JSON */ }
                    }
                }
            } else {
                const data = await res.json()
                setReply(data.choices?.[0]?.message?.content || '')
            }

            setStatus('')
        } catch (err) {
            if (err.name !== 'AbortError') {
                setReply(err.message)
                onMessage?.('error', err.message)
            }
            setStatus('')
        } finally {
            setLoading(false)
            abortRef.current = null
        }
    }

    const handleCancel = () => {
        abortRef.current?.abort()
        setLoading(false)
        setStatus('')
    }

    const handleClear = () => {
        setPrompt('')
        setReply('')
        setStatus('')
        onClear?.()
    }

    return (
        <div className="space-y-4">
            <div className="border" style={{ borderColor: 'var(--ds-border)', borderRadius: 'var(--radius-card)', background: 'var(--ds-card)' }}>
                <button
                    onClick={() => setCollapsed(!collapsed)}
                    className="w-full flex items-center justify-between px-5 py-3.5 transition-colors"
                    style={{ borderBottom: collapsed ? 'none' : '1px solid var(--ds-border)' }}
                >
                    <div className="flex items-center gap-2">
                        <Send className="w-4 h-4" style={{ color: 'var(--ds-blue)' }} />
                        <span className="text-sm font-bold" style={{ color: 'var(--ds-text)' }}>{t('chat.title')}</span>
                    </div>
                    {collapsed ? <ChevronDown className="w-4 h-4" style={{ color: 'var(--ds-text-tertiary)' }} /> : <ChevronUp className="w-4 h-4" style={{ color: 'var(--ds-text-tertiary)' }} />}
                </button>

                {!collapsed && (
                    <div className="p-5 space-y-4">
                        <div className="space-y-2">
                            {status && <StatusLabel message={status} />}

                            {reply && (
                                <div className="relative">
                                    <div className="flex items-center justify-between mb-2">
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
                                    <div
                                        ref={replyRef}
                                        className={clsx(
                                            "p-4 text-sm leading-relaxed whitespace-pre-wrap overflow-auto custom-scrollbar",
                                            expanded ? "max-h-[600px]" : "max-h-[300px]"
                                        )}
                                        style={{
                                            background: 'var(--ds-bg)',
                                            border: '1px solid var(--ds-border)',
                                            borderRadius: 'var(--radius-ctrl)',
                                            color: 'var(--ds-text)',
                                        }}
                                    >
                                        {reply}
                                    </div>
                                </div>
                            )}
                        </div>

                        <form onSubmit={handleSubmit} className="space-y-3">
                            <textarea
                                className="w-full p-3.5 text-sm resize-none"
                                rows={3}
                                style={{
                                    background: 'var(--ds-bg)',
                                    border: '1px solid var(--ds-border)',
                                    borderRadius: 'var(--radius-ctrl)',
                                    color: 'var(--ds-text)',
                                }}
                                placeholder={t('chat.placeholder')}
                                value={prompt}
                                onChange={e => setPrompt(e.target.value)}
                                onKeyDown={e => {
                                    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
                                        e.preventDefault()
                                        handleSubmit()
                                    }
                                }}
                            />

                            <div className="flex items-center justify-between gap-2">
                                <div className="flex items-center gap-2">
                                    <button
                                        type="button"
                                        onClick={() => setAdvanceOpen(!advanceOpen)}
                                        className="ds-btn-secondary text-[10px] px-2 py-1"
                                    >
                                        {advanceOpen ? <ChevronUp className="w-3 h-3 mr-1" /> : <ChevronDown className="w-3 h-3 mr-1" />}
                                        {t('chat.advanced')}
                                    </button>
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
                                            {t('chat.cancel')}
                                        </button>
                                    ) : null}
                                    <button
                                        type="submit"
                                        disabled={loading || !prompt.trim()}
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

                            {advanceOpen && (
                                <div className="p-4 space-y-3" style={{ background: 'var(--ds-bg)', border: '1px solid var(--ds-border)', borderRadius: 'var(--radius-ctrl)' }}>
                                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
                                        <div className="space-y-1">
                                            <label className="text-[10px] font-semibold uppercase tracking-wider" style={{ color: 'var(--ds-text-tertiary)' }}>{t('chat.model')}</label>
                                            <SegmentedControl
                                                options={models}
                                                value={modelMode}
                                                onChange={setModelMode}
                                                size="sm"
                                                ariaLabel={t('chat.model')}
                                            />
                                        </div>
                                        <div className="space-y-1">
                                            <label className="text-[10px] font-semibold uppercase tracking-wider" style={{ color: 'var(--ds-text-tertiary)' }}>{t('chat.stream')}</label>
                                            <div className="flex items-center gap-2 h-8">
                                                <button
                                                    type="button"
                                                    onClick={() => setStream(true)}
                                                    className="text-[10px] px-2 py-1 font-medium border transition-colors"
                                                    style={{
                                                        borderRadius: 'var(--radius-ctrl)',
                                                        background: stream ? 'var(--ds-blue)' : 'transparent',
                                                        color: stream ? 'var(--ds-text-on-primary)' : 'var(--ds-text-secondary)',
                                                        borderColor: stream ? 'var(--ds-blue)' : 'var(--ds-border)',
                                                    }}
                                                >
                                                    {t('chat.streamOn')}
                                                </button>
                                                <button
                                                    type="button"
                                                    onClick={() => setStream(false)}
                                                    className="text-[10px] px-2 py-1 font-medium border transition-colors"
                                                    style={{
                                                        borderRadius: 'var(--radius-ctrl)',
                                                        background: !stream ? 'var(--ds-blue)' : 'transparent',
                                                        color: !stream ? 'var(--ds-text-on-primary)' : 'var(--ds-text-secondary)',
                                                        borderColor: !stream ? 'var(--ds-blue)' : 'var(--ds-border)',
                                                    }}
                                                >
                                                    {t('chat.streamOff')}
                                                </button>
                                            </div>
                                        </div>
                                        <div className="space-y-1">
                                            <label className="text-[10px] font-semibold uppercase tracking-wider" style={{ color: 'var(--ds-text-tertiary)' }}>{t('chat.temperature')}</label>
                                            <input
                                                type="number"
                                                className="ds-input py-2 text-xs"
                                                min={0}
                                                max={2}
                                                step={0.1}
                                                value={temperature}
                                                onChange={e => setTemperature(parseFloat(e.target.value) || 0)}
                                            />
                                        </div>
                                        <div className="space-y-1">
                                            <label className="text-[10px] font-semibold uppercase tracking-wider" style={{ color: 'var(--ds-text-tertiary)' }}>{t('chat.maxTokens')}</label>
                                            <input
                                                type="number"
                                                className="ds-input py-2 text-xs"
                                                min={1}
                                                max={32768}
                                                step={1}
                                                value={maxTokens}
                                                onChange={e => setMaxTokens(parseInt(e.target.value) || 1)}
                                            />
                                        </div>
                                    </div>
                                </div>
                            )}
                        </form>
                    </div>
                )}
            </div>
        </div>
    )
}