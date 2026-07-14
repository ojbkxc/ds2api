import { useState } from 'react'
import { Check, ChevronDown, Copy, Pencil, Plus, Trash2 } from 'lucide-react'
import clsx from 'clsx'

import { maskSecret } from '../../utils/maskSecret'

function fallbackCopyText(text) {
    const textArea = document.createElement('textarea')
    textArea.value = text
    textArea.setAttribute('readonly', '')
    textArea.style.position = 'fixed'
    textArea.style.top = '-9999px'
    textArea.style.left = '-9999px'

    document.body.appendChild(textArea)
    textArea.focus()
    textArea.select()

    let copied = false
    try {
        copied = document.execCommand('copy')
    } finally {
        document.body.removeChild(textArea)
    }

    if (!copied) {
        throw new Error('copy failed')
    }
}

export default function ApiKeysPanel({
    t,
    config,
    keysExpanded,
    setKeysExpanded,
    onAddKey,
    onEditKey,
    copiedKey,
    setCopiedKey,
    onDeleteKey,
}) {
    const [failedKey, setFailedKey] = useState(null)
    const apiKeys = Array.isArray(config?.api_keys) && config.api_keys.length > 0
        ? config.api_keys
        : (config?.keys || []).map(key => ({ key, name: '', remark: '' }))

    const handleCopyKey = async (key) => {
        try {
            if (navigator.clipboard?.writeText) {
                await navigator.clipboard.writeText(key)
            } else {
                fallbackCopyText(key)
            }
            setCopiedKey(key)
            setFailedKey(null)
            setTimeout(() => setCopiedKey(null), 2000)
        } catch {
            try {
                fallbackCopyText(key)
                setCopiedKey(key)
                setFailedKey(null)
                setTimeout(() => setCopiedKey(null), 2000)
            } catch {
                setFailedKey(key)
                setTimeout(() => setFailedKey(null), 2500)
            }
        }
    }

    return (
        <div className="ds-card overflow-hidden">
            <div
                className="p-6 flex flex-col md:flex-row md:items-center justify-between gap-4 cursor-pointer select-none transition-colors"
                style={{ borderRadius: 'var(--radius-card)' }}
                onClick={() => setKeysExpanded(!keysExpanded)}
            >
                <div className="flex items-center gap-3">
                    <ChevronDown className={clsx(
                        "w-5 h-5 transition-transform duration-200",
                        keysExpanded ? "rotate-0" : "-rotate-90"
                    )} style={{ color: 'var(--ds-text-secondary)' }} />
                    <div>
                        <h2 className="text-lg font-semibold" style={{ color: 'var(--ds-text)' }}>{t('accountManager.apiKeysTitle')}</h2>
                        <p className="text-sm" style={{ color: 'var(--ds-text-secondary)' }}>{t('accountManager.apiKeysDesc')} ({apiKeys.length || 0})</p>
                    </div>
                </div>
                <button
                    onClick={(e) => { e.stopPropagation(); onAddKey() }}
                    className="ds-btn-primary text-sm font-medium"
                >
                    <Plus className="w-4 h-4" />
                    {t('accountManager.addKey')}
                </button>
            </div>

            {keysExpanded && (
                <div className="divide-y border-t" style={{ borderColor: 'var(--ds-border)' }}>
                    {apiKeys.length > 0 ? (
                        apiKeys.map((item, i) => (
                            <div key={i} className="p-4 flex items-center justify-between transition-colors group" style={{ borderRadius: 'var(--radius-card)' }}>
                                <div className="grid grid-cols-1 md:grid-cols-3 gap-2 flex-1">
                                    <div className="text-sm" style={{ color: 'var(--ds-text)' }}>{item.name || '-'}</div>
                                    <button
                                        onClick={() => handleCopyKey(item.key)}
                                        className="font-mono text-sm px-3 py-1 rounded inline-block transition-colors"
                                        style={{ backgroundColor: 'var(--ds-surface)' }}
                                        title={t('accountManager.copyKeyTitle')}
                                    >
                                        {maskSecret(item.key)}
                                    </button>
                                    <div className="text-sm truncate" style={{ color: 'var(--ds-text-secondary)' }}>{item.remark || '-'}</div>
                                    {copiedKey === item.key && (
                                        <span className="text-xs animate-pulse" style={{ color: 'var(--ds-success)' }}>{t('accountManager.copied')}</span>
                                    )}
                                    {failedKey === item.key && (
                                        <span className="text-xs" style={{ color: 'var(--ds-danger)' }}>{t('accountManager.copyFailed')}</span>
                                    )}
                                </div>
                                <div className="flex items-center gap-1">
                                    <button
                                        onClick={() => onEditKey(item)}
                                        className="p-2 rounded-md transition-colors"
                                        style={{ color: 'var(--ds-text-tertiary)' }}
                                        title={t('accountManager.editKeyTitle')}
                                    >
                                        <Pencil className="w-4 h-4" />
                                    </button>
                                    <button
                                        onClick={() => handleCopyKey(item.key)}
                                        className="p-2 rounded-md transition-colors"
                                        style={{ color: 'var(--ds-text-tertiary)' }}
                                        title={t('accountManager.copyKeyTitle')}
                                    >
                                        {copiedKey === item.key ? <Check className="w-4 h-4" style={{ color: 'var(--ds-success)' }} /> : <Copy className="w-4 h-4" />}
                                    </button>
                                    <button
                                        onClick={() => onDeleteKey(item.key)}
                                        className="p-2 rounded-md transition-colors"
                                        style={{ color: 'var(--ds-text-tertiary)' }}
                                        title={t('accountManager.deleteKeyTitle')}
                                    >
                                        <Trash2 className="w-4 h-4" />
                                    </button>
                                </div>
                            </div>
                        ))
                    ) : (
                        <div className="p-8 text-center" style={{ color: 'var(--ds-text-tertiary)' }}>{t('accountManager.noApiKeys')}</div>
                    )}
                </div>
            )}
        </div>
    )
}
