import {
    ChevronDown,
    MessageSquare,
    Cpu,
    Search as SearchIcon,
    Terminal,
    Zap,
    ToggleLeft,
    ToggleRight,
} from 'lucide-react'
import clsx from 'clsx'

import { maskSecret } from '../../utils/maskSecret'

export default function ConfigPanel({
    t,
    configExpanded,
    setConfigExpanded,
    models,
    model,
    setModel,
    modelsLoaded,
    streamingMode,
    setStreamingMode,
    selectedAccount,
    setSelectedAccount,
    accounts,
    resolveAccountIdentifier,
    apiKey,
    setApiKey,
    config,
    customKeyActive,
    customKeyManaged,
}) {
    const iconMap = {
        MessageSquare,
        Cpu,
        SearchIcon,
        Terminal,
        Zap,
        ToggleLeft,
        ToggleRight,
    }
    const selectedModel = models.find(m => m.id === model) || models[0]
    const SelectedModelIcon = selectedModel ? (iconMap[selectedModel.icon] || MessageSquare) : MessageSquare
    const defaultKeyPreview = maskSecret(config.keys?.[0])
    const hasModels = models.length > 0

    const labelStyle = {
        fontSize: 11,
        fontWeight: 600,
        color: 'var(--ds-text-tertiary)',
        textTransform: 'uppercase',
        letterSpacing: '0.05em',
        marginLeft: 2,
    }

    const selectStyle = {
        width: '100%',
        height: 40,
        paddingLeft: 12,
        paddingRight: 36,
        background: 'var(--ds-surface)',
        border: '1px solid var(--ds-border)',
        borderRadius: 'var(--radius-ctrl)',
        fontSize: 13,
        color: 'var(--ds-text)',
        appearance: 'none',
        cursor: 'pointer',
        transition: 'border-color 0.15s',
    }

    return (
        <div
            className={clsx(
                'lg:col-span-3 flex flex-col transition-all duration-300 ease-in-out z-20 min-h-0',
                configExpanded ? 'h-auto' : 'h-14 lg:h-full',
            )}
        >
            <div
                className="ds-card flex flex-col h-full min-h-0 overflow-hidden"
                style={{ boxShadow: 'var(--ds-elevate-1)' }}
            >
                {/* Mobile expand toggle */}
                <button
                    onClick={() => setConfigExpanded(!configExpanded)}
                    className="lg:hidden flex items-center justify-between p-4 w-full transition-colors"
                    style={{
                        background: 'var(--ds-surface)',
                        border: 'none',
                        cursor: 'pointer',
                    }}
                >
                    <div className="flex items-center gap-2.5 font-medium text-sm" style={{ color: 'var(--ds-text)' }}>
                        <div
                            style={{
                                padding: '6px',
                                borderRadius: 'var(--radius-ctrl)',
                                background: 'transparent',
                                color: 'var(--ds-text)',
                            }}
                        >
                            <Terminal className="w-4 h-4" />
                        </div>
                        <span>{t('apiTester.config')}</span>
                    </div>
                    <div
                        className={clsx(
                            'transition-transform duration-300',
                            configExpanded ? 'rotate-180' : '',
                        )}
                        style={{ color: 'var(--ds-text-secondary)' }}
                    >
                        <ChevronDown className="w-4 h-4" />
                    </div>
                </button>

                <div
                    className={clsx(
                        'p-4 flex flex-col gap-5',
                        !configExpanded && 'hidden lg:flex',
                    )}
                >
                    {/* Model selector */}
                    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }} className="shrink-0">
                        <label style={labelStyle}>{t('apiTester.modelLabel')}</label>
                        <div style={{ position: 'relative' }}>
                            <select
                                style={selectStyle}
                                value={model}
                                onChange={e => setModel(e.target.value)}
                                disabled={!hasModels}
                            >
                                {hasModels
                                    ? models.map(m => (
                                        <option key={m.id} value={m.id}>
                                            {m.name}
                                        </option>
                                    ))
                                    : (
                                        <option value="">
                                            {modelsLoaded ? t('apiTester.noModels') : t('apiTester.loadingModels')}
                                        </option>
                                    )}
                            </select>
                            <ChevronDown
                                className="w-4 h-4 pointer-events-none"
                                style={{
                                    position: 'absolute',
                                    right: 10,
                                    top: 12,
                                    color: 'var(--ds-text-tertiary)',
                                }}
                            />
                        </div>
                        {selectedModel ? (
                            <div
                                style={{
                                    borderRadius: 'var(--radius-ctrl)',
                                    border: '1px solid var(--ds-border)',
                                    background: 'var(--ds-bg)',
                                    padding: 12,
                                    marginTop: 4,
                                }}
                            >
                                <div style={{ display: 'flex', alignItems: 'flex-start', gap: 12 }}>
                                    <div
                                        style={{
                                            padding: 8,
                                            borderRadius: 'var(--radius-ctrl)',
                                            flexShrink: 0,
                                            border: '1px solid var(--ds-border)',
                                            background: 'var(--ds-card)',
                                            color: 'var(--ds-blue)',
                                        }}
                                    >
                                        <SelectedModelIcon className="w-4 h-4" />
                                    </div>
                                    <div style={{ minWidth: 0, flex: 1 }}>
                                        <div
                                            style={{
                                                fontWeight: 500,
                                                fontSize: 13,
                                                color: 'var(--ds-text)',
                                                overflow: 'hidden',
                                                textOverflow: 'ellipsis',
                                                whiteSpace: 'nowrap',
                                            }}
                                        >
                                            {selectedModel.name}
                                        </div>
                                        <div
                                            style={{
                                                fontSize: 11,
                                                color: 'var(--ds-text-tertiary)',
                                                marginTop: 4,
                                                lineHeight: 1.5,
                                            }}
                                        >
                                            {selectedModel.desc}
                                        </div>
                                    </div>
                                </div>
                                <p
                                    style={{
                                        fontSize: 11,
                                        color: 'var(--ds-text-tertiary)',
                                        marginTop: 8,
                                        margin: '8px 0 0 0',
                                    }}
                                >
                                    {t('apiTester.modelPickerHint')}
                                </p>
                            </div>
                        ) : (
                            <div
                                style={{
                                    borderRadius: 'var(--radius-ctrl)',
                                    border: '1px dashed var(--ds-border)',
                                    background: 'var(--ds-bg)',
                                    padding: 12,
                                    marginTop: 4,
                                    fontSize: 11,
                                    color: 'var(--ds-text-tertiary)',
                                    lineHeight: 1.5,
                                }}
                            >
                                {modelsLoaded ? t('apiTester.noModelsHint') : t('apiTester.loadingModelsHint')}
                            </div>
                        )}
                    </div>

                    {/* Stream mode toggle */}
                    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }} className="shrink-0">
                        <label style={labelStyle}>{t('apiTester.streamMode')}</label>
                        <button
                            onClick={() => setStreamingMode(!streamingMode)}
                            className={clsx(
                                'w-full flex items-center justify-between px-3 py-2 transition-all duration-200',
                            )}
                            style={{
                                borderRadius: 'var(--radius-ctrl)',
                                border: '1px solid',
                                borderColor: streamingMode ? 'var(--ds-blue)' : 'var(--ds-border)',
                                background: streamingMode ? 'var(--ds-blue-light)' : 'var(--ds-card)',
                                color: streamingMode ? 'var(--ds-text)' : 'var(--ds-text-secondary)',
                                cursor: 'pointer',
                            }}
                        >
                            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                                <div
                                    style={{
                                        padding: 6,
                                        borderRadius: 'var(--radius-ctrl)',
                                        background: streamingMode ? 'var(--ds-blue)' : 'var(--ds-surface)',
                                        color: streamingMode ? 'var(--ds-text-on-primary)' : 'var(--ds-text-tertiary)',
                                    }}
                                >
                                    <Zap className="w-4 h-4" />
                                </div>
                                <span style={{ fontSize: 13, fontWeight: 500 }}>{t('apiTester.streamMode')}</span>
                            </div>
                            {streamingMode ? (
                                <ToggleRight className="w-5 h-5" style={{ color: 'var(--ds-blue)' }} />
                            ) : (
                                <ToggleLeft className="w-5 h-5" style={{ color: 'var(--ds-text-tertiary)' }} />
                            )}
                        </button>
                    </div>

                    {/* Account selector */}
                    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }} className="shrink-0">
                        <label style={labelStyle}>{t('apiTester.accountSelector')}</label>
                        <div style={{ position: 'relative' }}>
                            <select
                                style={selectStyle}
                                value={selectedAccount}
                                onChange={e => setSelectedAccount(e.target.value)}
                            >
                                <option value="">{t('apiTester.autoRandom')}</option>
                                {accounts.map((acc, i) => {
                                    const id = resolveAccountIdentifier(acc)
                                    if (!id) return null
                                    return (
                                        <option key={i} value={id}>
                                            {id}
                                        </option>
                                    )
                                })}
                            </select>
                            <ChevronDown
                                className="w-4 h-4 pointer-events-none"
                                style={{
                                    position: 'absolute',
                                    right: 10,
                                    top: 12,
                                    color: 'var(--ds-text-tertiary)',
                                }}
                            />
                        </div>
                    </div>

                    {/* API key input */}
                    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }} className="shrink-0">
                        <label style={labelStyle}>{t('apiTester.apiKeyOptional')}</label>
                        <input
                            type="text"
                            autoComplete="off"
                            spellCheck={false}
                            className="ds-input"
                            style={{ fontFamily: 'monospace', height: 40 }}
                            placeholder={
                                defaultKeyPreview
                                    ? t('apiTester.apiKeyDefault', { preview: defaultKeyPreview })
                                    : t('apiTester.apiKeyPlaceholder')
                            }
                            value={apiKey}
                            onChange={e => setApiKey(e.target.value)}
                        />
                        {customKeyActive && (
                            <p
                                style={{
                                    fontSize: 11,
                                    marginTop: 2,
                                    color: customKeyManaged ? 'var(--ds-success)' : 'var(--ds-warning)',
                                }}
                            >
                                {customKeyManaged ? t('apiTester.modeManaged') : t('apiTester.modeDirect')}
                            </p>
                        )}
                    </div>
                </div>
            </div>
        </div>
    )
}