import { useState } from 'react'
import { FileCode, Download, Upload, Copy, Check, AlertTriangle } from 'lucide-react'

import { useI18n } from '../i18n'
import { getBatchImportTemplates } from '../utils/batchImportTemplates'
import Button from '../components/ui/Button'

export default function BatchImport({ onRefresh, onMessage, authFetch }) {
    const { t } = useI18n()
    const [jsonInput, setJsonInput] = useState('')
    const [loading, setLoading] = useState(false)
    const [result, setResult] = useState(null)
    const [copied, setCopied] = useState(false)

    const apiFetch = authFetch || fetch
    const templates = getBatchImportTemplates(t)

    const handleImport = async () => {
        if (!jsonInput.trim()) {
            onMessage('error', t('batchImport.enterJson'))
            return
        }

        let config
        try {
            config = JSON.parse(jsonInput)
        } catch (e) {
            onMessage('error', t('messages.invalidJson'))
            return
        }

        setLoading(true)
        setResult(null)
        try {
            const res = await apiFetch('/admin/import', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(config),
            })
            const data = await res.json()
            if (res.ok) {
                setResult(data)
                onMessage('success', t('batchImport.importSuccess', { keys: data.imported_keys, accounts: data.imported_accounts }))
                onRefresh()
            } else {
                onMessage('error', data.detail || t('messages.importFailed'))
            }
        } catch (e) {
            onMessage('error', t('messages.networkError'))
        } finally {
            setLoading(false)
        }
    }

    const loadTemplate = (key) => {
        const tpl = templates[key]
        if (tpl) {
            setJsonInput(JSON.stringify(tpl.config, null, 2))
            onMessage('info', t('batchImport.templateLoaded', { name: tpl.name }))
        }
    }

    const handleExport = async () => {
        try {
            const res = await apiFetch('/admin/export')
            if (res.ok) {
                const data = await res.json()
                setJsonInput(JSON.stringify(JSON.parse(data.json), null, 2))
                onMessage('success', t('batchImport.currentConfigLoaded'))
            }
        } catch (e) {
            onMessage('error', t('batchImport.fetchConfigFailed'))
        }
    }

    const copyBase64 = async () => {
        try {
            const res = await apiFetch('/admin/export')
            if (res.ok) {
                const data = await res.json()
                await navigator.clipboard.writeText(data.base64)
                setCopied(true)
                setTimeout(() => setCopied(false), 2000)
                onMessage('success', t('batchImport.copySuccess'))
            }
        } catch (e) {
            onMessage('error', t('messages.copyFailed'))
        }
    }

    return (
        <div className="flex flex-col lg:grid lg:grid-cols-3 gap-6 lg:h-[calc(100vh-140px)]">
            {/* Templates Panel */}
            <div className="md:col-span-1 space-y-4">
                <div className="ds-card p-5">
                    <h3 className="font-semibold flex items-center gap-2 mb-4" style={{ color: 'var(--ds-text)' }}>
                        <FileCode className="w-4 h-4" style={{ color: 'var(--ds-blue)' }} />
                        {t('batchImport.quickTemplates')}
                    </h3>
                    <div className="space-y-3">
                        {Object.entries(templates).map(([key, tpl]) => (
                            <button
                                key={key}
                                onClick={() => loadTemplate(key)}
                                style={{
                                    display: 'block',
                                    width: '100%',
                                    textAlign: 'left',
                                    padding: '0.75rem',
                                    borderRadius: 'var(--radius-ctrl)',
                                    border: '1px solid var(--ds-border)',
                                    background: 'var(--ds-surface)',
                                    color: 'var(--ds-text)',
                                    cursor: 'pointer',
                                    transition: 'all 0.15s',
                                }}
                                onMouseEnter={(e) => {
                                    e.currentTarget.style.borderColor = 'var(--ds-blue)'
                                    e.currentTarget.style.background = 'var(--ds-surface-hover)'
                                }}
                                onMouseLeave={(e) => {
                                    e.currentTarget.style.borderColor = 'var(--ds-border)'
                                    e.currentTarget.style.background = 'var(--ds-surface)'
                                }}
                            >
                                <div className="font-medium text-sm" style={{ color: 'var(--ds-text)' }}>
                                    {tpl.name}
                                </div>
                                <div className="text-xs mt-0.5" style={{ color: 'var(--ds-text-secondary)' }}>
                                    {tpl.desc}
                                </div>
                            </button>
                        ))}
                    </div>
                </div>

                <div
                    className="p-5"
                    style={{
                        background: 'var(--ds-blue-light)',
                        border: '1px solid var(--ds-selected-border)',
                        borderRadius: 'var(--radius-card)',
                    }}
                >
                    <h3 className="font-semibold flex items-center gap-2 mb-2" style={{ color: 'var(--ds-blue)' }}>
                        <Download className="w-4 h-4" />
                        {t('batchImport.dataExport')}
                    </h3>
                    <p className="text-sm mb-4" style={{ color: 'var(--ds-text-secondary)' }}>
                        {t('batchImport.dataExportDesc')}
                    </p>
                    <Button variant="primary" size="md" onClick={copyBase64} className="w-full">
                        {copied ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
                        <span className="ml-1.5">{copied ? t('batchImport.copied') : t('batchImport.copyBase64')}</span>
                    </Button>
                    <p className="text-[10px] mt-2 text-center" style={{ color: 'var(--ds-text-tertiary)' }}>
                        {t('batchImport.variableName')}:{' '}
                        <code
                            style={{
                                background: 'var(--ds-bg)',
                                padding: '0.125rem 0.375rem',
                                borderRadius: 'var(--radius-ctrl)',
                                border: '1px solid var(--ds-border)',
                                color: 'var(--ds-text)',
                            }}
                        >
                            DS2API_CONFIG_JSON
                        </code>
                    </p>
                </div>
            </div>

            {/* Editor Panel */}
            <div
                className="lg:col-span-2 flex flex-col overflow-hidden min-h-[400px] lg:h-full"
                style={{
                    background: 'var(--ds-card)',
                    border: '1px solid var(--ds-border)',
                    borderRadius: 'var(--radius-card)',
                }}
            >
                <div
                    className="p-4 flex items-center justify-between"
                    style={{
                        borderBottom: '1px solid var(--ds-border)',
                        background: 'var(--ds-surface)',
                    }}
                >
                    <h3 className="font-semibold flex items-center gap-2" style={{ color: 'var(--ds-text)' }}>
                        <Upload className="w-4 h-4" style={{ color: 'var(--ds-blue)' }} />
                        {t('batchImport.jsonEditor')}
                    </h3>
                    <div className="flex gap-2">
                        <Button variant="secondary" size="sm" onClick={handleExport}>
                            {t('batchImport.loadCurrentConfig')}
                        </Button>
                        <Button variant="primary" size="sm" onClick={handleImport} disabled={loading}>
                            {loading ? t('batchImport.importing') : t('batchImport.applyConfig')}
                        </Button>
                    </div>
                </div>

                <div className="flex-1 relative min-h-[400px]">
                    <textarea
                        className="absolute inset-0 w-full h-full p-4 font-mono text-sm resize-none focus:outline-none custom-scrollbar"
                        style={{
                            background: 'var(--ds-bg)',
                            color: 'var(--ds-text)',
                            border: 'none',
                        }}
                        value={jsonInput}
                        onChange={e => setJsonInput(e.target.value)}
                        placeholder={'{\n  "keys": ["your-api-key"],\n  "accounts": [\n    {"email": "...", "password": "...", "token": ""}\n  ]\n}'}
                        spellCheck={false}
                    />
                </div>

                {result && (
                    <div
                        className="p-4"
                        style={{
                            borderTop: '1px solid',
                            background: (result.imported_keys || result.imported_accounts)
                                ? 'var(--ds-success-bg)'
                                : 'var(--ds-danger-bg)',
                            borderColor: (result.imported_keys || result.imported_accounts)
                                ? 'var(--ds-success-border)'
                                : 'var(--ds-danger-border)',
                        }}
                    >
                        <div className="flex items-start gap-3">
                            {result.imported_keys || result.imported_accounts ? (
                                <Check className="w-5 h-5 mt-0.5" style={{ color: 'var(--ds-success)' }} />
                            ) : (
                                <AlertTriangle className="w-5 h-5 mt-0.5" style={{ color: 'var(--ds-danger)' }} />
                            )}
                            <div>
                                <h4 className="font-medium" style={{
                                    color: (result.imported_keys || result.imported_accounts)
                                        ? 'var(--ds-success)'
                                        : 'var(--ds-danger)'
                                }}>
                                    {t('batchImport.importComplete')}
                                </h4>
                                <p className="text-sm mt-1" style={{ color: 'var(--ds-text-secondary)' }}>
                                    {t('batchImport.importSummary', { keys: result.imported_keys, accounts: result.imported_accounts })}
                                </p>
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </div>
    )
}