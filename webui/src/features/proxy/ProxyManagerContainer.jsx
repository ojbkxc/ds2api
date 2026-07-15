import { useState } from 'react'
import { Pencil, Play, Plus, Shield, Trash2 } from 'lucide-react'

import { useI18n } from '../../i18n'
import Modal from '../../components/ui/Modal'
import Button from '../../components/ui/Button'
import Input from '../../components/ui/Input'
import Badge from '../../components/ui/Badge'
import EmptyState from '../../components/ui/EmptyState'

async function readApiResponse(res, nonJsonMessage) {
    const contentType = String(res.headers.get('content-type') || '').toLowerCase()
    const raw = await res.text()
    const trimmed = raw.trim()

    if (!trimmed) {
        return {}
    }

    if (contentType.includes('application/json')) {
        try {
            return JSON.parse(trimmed)
        } catch (_err) {
            if (!res.ok) {
                return { detail: trimmed }
            }
            throw new Error(nonJsonMessage)
        }
    }

    if (!res.ok) {
        return { detail: trimmed }
    }

    throw new Error(nonJsonMessage)
}

const EMPTY_FORM = {
    name: '',
    type: 'socks5h',
    host: '',
    port: 1080,
    username: '',
    password: '',
}

function createEmptyProxyForm() {
    return { ...EMPTY_FORM }
}

function ProxyStatusBadge({ t, result, testing = false }) {
    if (testing) {
        return <Badge tone="info">{t('proxyManager.testing')}</Badge>
    }
    if (!result) {
        return <Badge tone="muted">{t('proxyManager.untested')}</Badge>
    }
    return (
        <Badge tone={result.success ? 'success' : 'danger'}>
            {result.success
                ? t('proxyManager.testSuccessShort', { time: result.response_time ?? 0 })
                : t('proxyManager.testFailedShort')}
        </Badge>
    )
}

function ProxiesTable({
    t,
    proxies,
    testing,
    testResults,
    onCreate,
    onTest,
    onEdit,
    onDelete,
}) {
    return (
        <div className="ds-card" style={{ borderRadius: 'var(--radius-card)', overflow: 'hidden' }}>
            <div
                className="p-6 flex flex-col md:flex-row md:items-center justify-between gap-4"
                style={{ borderBottom: '1px solid var(--ds-border)' }}
            >
                <div>
                    <h2 className="text-lg font-semibold" style={{ color: 'var(--ds-text)' }}>{t('proxyManager.title')}</h2>
                    <p className="text-sm mt-0.5" style={{ color: 'var(--ds-text-secondary)' }}>{t('proxyManager.desc')}</p>
                </div>
                <Button variant="primary" size="md" onClick={onCreate}>
                    <Plus className="w-4 h-4" />
                    <span className="ml-1.5">{t('proxyManager.addProxy')}</span>
                </Button>
            </div>

            {proxies.length === 0 ? (
                <EmptyState
                    icon={<Shield className="w-8 h-8" />}
                    title={t('proxyManager.noProxies')}
                />
            ) : (
                <div style={{ display: 'flex', flexDirection: 'column' }}>
                    {proxies.map((proxy) => {
                        const result = testResults[proxy.id]
                        return (
                            <div
                                key={proxy.id}
                                className="p-4 md:p-5 flex flex-col lg:flex-row lg:items-center justify-between gap-4"
                                style={{
                                    borderTop: '1px solid var(--ds-border)',
                                    transition: 'background 0.15s',
                                }}
                                onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--ds-surface-hover)' }}
                                onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent' }}
                            >
                                <div className="min-w-0">
                                    <div className="flex flex-wrap items-center gap-2">
                                        <div className="font-medium" style={{ color: 'var(--ds-text)' }}>
                                            {proxy.name || `${proxy.host}:${proxy.port}`}
                                        </div>
                                        <Badge tone="info">{proxy.type}</Badge>
                                        {proxy.username && (
                                            <Badge tone="muted">
                                                <Shield className="w-3 h-3" style={{ display: 'inline', marginRight: 2 }} />
                                                {proxy.username}
                                            </Badge>
                                        )}
                                        <ProxyStatusBadge t={t} result={result} testing={testing[proxy.id]} />
                                    </div>
                                    <div className="mt-2 flex flex-wrap items-center gap-2 text-xs" style={{ color: 'var(--ds-text-secondary)' }}>
                                        <span
                                            className="font-mono px-2 py-1"
                                            style={{
                                                background: 'var(--ds-bg)',
                                                border: '1px solid var(--ds-border)',
                                                borderRadius: 'var(--radius-ctrl)',
                                                color: 'var(--ds-text)',
                                                fontSize: '0.75rem',
                                            }}
                                        >
                                            {proxy.host}:{proxy.port}
                                        </span>
                                        {proxy.has_password && (
                                            <Badge tone="warning">{t('proxyManager.authEnabled')}</Badge>
                                        )}
                                        {result?.message && (
                                            <span className="truncate max-w-full">{result.message}</span>
                                        )}
                                    </div>
                                </div>

                                <div className="flex items-center gap-2 self-start lg:self-auto">
                                    <Button
                                        variant="secondary"
                                        size="sm"
                                        onClick={() => onTest(proxy)}
                                        disabled={testing[proxy.id]}
                                    >
                                        <Play className="w-3.5 h-3.5" />
                                        <span className="ml-1">{t('proxyManager.testAction')}</span>
                                    </Button>
                                    <button
                                        onClick={() => onEdit(proxy)}
                                        className="ds-action-btn p-2"
                                        style={{ borderRadius: 'var(--radius-ctrl)' }}
                                        title={t('proxyManager.editProxy')}
                                    >
                                        <Pencil className="w-4 h-4" />
                                    </button>
                                    <button
                                        onClick={() => onDelete(proxy)}
                                        className="ds-action-btn p-2"
                                        style={{
                                            borderRadius: 'var(--radius-ctrl)',
                                            color: 'var(--ds-text-tertiary)',
                                        }}
                                        onMouseEnter={(e) => { e.currentTarget.style.color = 'var(--ds-danger)' }}
                                        onMouseLeave={(e) => { e.currentTarget.style.color = 'var(--ds-text-tertiary)' }}
                                        title={t('proxyManager.deleteProxy')}
                                    >
                                        <Trash2 className="w-4 h-4" />
                                    </button>
                                </div>
                            </div>
                        )
                    })}
                </div>
            )}
        </div>
    )
}

function ProxyFormModal({
    show,
    t,
    form,
    setForm,
    editingProxy,
    loading,
    onClose,
    onSubmit,
}) {
    const isEditing = Boolean(editingProxy?.id)

    return (
        <Modal
            open={show}
            onClose={onClose}
            title={isEditing ? t('proxyManager.modalEditTitle') : t('proxyManager.modalAddTitle')}
            maxWidth="max-w-lg"
        >
            <p className="text-xs mb-4" style={{ color: 'var(--ds-text-tertiary)', marginTop: -8 }}>
                {t('proxyManager.modalDesc')}
            </p>

            <div className="space-y-4">
                <div className="grid md:grid-cols-2 gap-4">
                    <div>
                        <label className="block text-sm font-medium mb-1.5" style={{ color: 'var(--ds-text-secondary)' }}>
                            {t('proxyManager.nameLabel')}
                        </label>
                        <Input
                            type="text"
                            placeholder={t('proxyManager.namePlaceholder')}
                            value={form.name}
                            onChange={e => setForm({ ...form, name: e.target.value })}
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium mb-1.5" style={{ color: 'var(--ds-text-secondary)' }}>
                            {t('proxyManager.typeLabel')}
                        </label>
                        <select
                            className="ds-input"
                            value={form.type}
                            onChange={e => setForm({ ...form, type: e.target.value })}
                        >
                            <option value="socks5">socks5</option>
                            <option value="socks5h">socks5h</option>
                        </select>
                    </div>
                </div>

                <div className="grid md:grid-cols-[1fr_128px] gap-4">
                    <div>
                        <label className="block text-sm font-medium mb-1.5" style={{ color: 'var(--ds-text-secondary)' }}>
                            {t('proxyManager.hostLabel')}
                        </label>
                        <Input
                            type="text"
                            placeholder={t('proxyManager.hostPlaceholder')}
                            value={form.host}
                            onChange={e => setForm({ ...form, host: e.target.value })}
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium mb-1.5" style={{ color: 'var(--ds-text-secondary)' }}>
                            {t('proxyManager.portLabel')}
                        </label>
                        <Input
                            type="number"
                            min="1"
                            max="65535"
                            value={form.port}
                            onChange={e => setForm({ ...form, port: Number(e.target.value) || '' })}
                        />
                    </div>
                </div>

                <div className="grid md:grid-cols-2 gap-4">
                    <div>
                        <label className="block text-sm font-medium mb-1.5" style={{ color: 'var(--ds-text-secondary)' }}>
                            {t('proxyManager.usernameLabel')}
                        </label>
                        <Input
                            type="text"
                            placeholder={t('proxyManager.usernamePlaceholder')}
                            value={form.username}
                            onChange={e => setForm({ ...form, username: e.target.value })}
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium mb-1.5" style={{ color: 'var(--ds-text-secondary)' }}>
                            {t('proxyManager.passwordLabel')}
                        </label>
                        <Input
                            type="password"
                            placeholder={t('proxyManager.passwordPlaceholder')}
                            value={form.password}
                            onChange={e => setForm({ ...form, password: e.target.value })}
                        />
                        {isEditing && (
                            <p className="mt-1 text-[11px]" style={{ color: 'var(--ds-text-tertiary)' }}>
                                {t('proxyManager.passwordKeepHint')}
                            </p>
                        )}
                    </div>
                </div>

                <div
                    className="px-3 py-2 text-xs"
                    style={{
                        background: 'var(--ds-surface)',
                        border: '1px solid var(--ds-border)',
                        borderRadius: 'var(--radius-ctrl)',
                        color: 'var(--ds-text-secondary)',
                    }}
                >
                    {t('proxyManager.typeHelp')}
                </div>

                <div className="ds-modal-actions pt-2" style={{ marginTop: 0 }}>
                    <Button variant="secondary" size="md" onClick={onClose}>
                        {t('actions.cancel')}
                    </Button>
                    <Button variant="primary" size="md" onClick={onSubmit} disabled={loading}>
                        {loading
                            ? t('proxyManager.saving')
                            : (isEditing ? t('proxyManager.saveEdit') : t('proxyManager.saveAdd'))}
                    </Button>
                </div>
            </div>
        </Modal>
    )
}

export default function ProxyManagerContainer({ config, onRefresh, onMessage, authFetch }) {
    const { t } = useI18n()
    const apiFetch = authFetch || fetch

    const [showModal, setShowModal] = useState(false)
    const [editingProxy, setEditingProxy] = useState(null)
    const [form, setForm] = useState(createEmptyProxyForm())
    const [saving, setSaving] = useState(false)
    const [testing, setTesting] = useState({})
    const [testResults, setTestResults] = useState({})

    const proxies = config?.proxies || []

    const openCreate = () => {
        setEditingProxy(null)
        setForm(createEmptyProxyForm())
        setShowModal(true)
    }

    const openEdit = (proxy) => {
        setEditingProxy(proxy)
        setForm({
            name: proxy.name || '',
            type: proxy.type || 'socks5h',
            host: proxy.host || '',
            port: proxy.port || 1080,
            username: proxy.username || '',
            password: '',
        })
        setShowModal(true)
    }

    const closeModal = () => {
        setShowModal(false)
        setEditingProxy(null)
        setForm(createEmptyProxyForm())
    }

    const saveProxy = async () => {
        if (!form.host || !form.port) {
            onMessage('error', t('proxyManager.requiredFields'))
            return
        }
        setSaving(true)
        try {
            const url = editingProxy?.id
                ? `/admin/proxies/${encodeURIComponent(editingProxy.id)}`
                : '/admin/proxies'
            const method = editingProxy?.id ? 'PUT' : 'POST'
            const res = await apiFetch(url, {
                method,
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    name: form.name,
                    type: form.type,
                    host: form.host,
                    port: Number(form.port),
                    username: form.username,
                    password: form.password,
                }),
            })
            const data = await readApiResponse(res, t('settings.nonJsonResponse', { status: res.status }))
            if (!res.ok) {
                onMessage('error', data.detail || t('messages.requestFailed'))
                return
            }
            await onRefresh?.()
            onMessage('success', editingProxy?.id ? t('proxyManager.updateSuccess') : t('proxyManager.addSuccess'))
            closeModal()
        } catch (err) {
            onMessage('error', err?.message || t('messages.networkError'))
        } finally {
            setSaving(false)
        }
    }

    const deleteProxy = async (proxy) => {
        if (!confirm(t('proxyManager.deleteConfirm', { name: proxy.name || `${proxy.host}:${proxy.port}` }))) return
        try {
            const res = await apiFetch(`/admin/proxies/${encodeURIComponent(proxy.id)}`, { method: 'DELETE' })
            const data = await readApiResponse(res, t('settings.nonJsonResponse', { status: res.status }))
            if (!res.ok) {
                onMessage('error', data.detail || t('messages.deleteFailed'))
                return
            }
            await onRefresh?.()
            onMessage('success', t('messages.deleted'))
            setTestResults(prev => {
                const next = { ...prev }
                delete next[proxy.id]
                return next
            })
        } catch (err) {
            onMessage('error', err?.message || t('messages.networkError'))
        }
    }

    const testProxy = async (proxy) => {
        setTesting(prev => ({ ...prev, [proxy.id]: true }))
        try {
            const res = await apiFetch('/admin/proxies/test', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ proxy_id: proxy.id }),
            })
            const data = await readApiResponse(res, t('settings.nonJsonResponse', { status: res.status }))
            setTestResults(prev => ({ ...prev, [proxy.id]: data }))
            onMessage(data.success ? 'success' : 'error', data.message || t('messages.requestFailed'))
        } catch (err) {
            onMessage('error', err?.message || t('messages.networkError'))
        } finally {
            setTesting(prev => ({ ...prev, [proxy.id]: false }))
        }
    }

    return (
        <div className="space-y-6">
            <div className="grid gap-4 md:grid-cols-3">
                <div className="ds-card p-5">
                    <div className="text-[10px] font-bold uppercase tracking-wider" style={{ color: 'var(--ds-text-tertiary)' }}>
                        {t('proxyManager.totalProxies')}
                    </div>
                    <div className="mt-2 text-2xl font-bold" style={{ color: 'var(--ds-text)' }}>
                        {proxies.length}
                    </div>
                </div>
                <div className="ds-card p-5">
                    <div className="text-[10px] font-bold uppercase tracking-wider" style={{ color: 'var(--ds-text-tertiary)' }}>
                        {t('proxyManager.socks5hCount')}
                    </div>
                    <div className="mt-2 text-2xl font-bold" style={{ color: 'var(--ds-text)' }}>
                        {proxies.filter(proxy => proxy.type === 'socks5h').length}
                    </div>
                </div>
                <div className="ds-card p-5">
                    <div className="text-[10px] font-bold uppercase tracking-wider" style={{ color: 'var(--ds-text-tertiary)' }}>
                        {t('proxyManager.authProxyCount')}
                    </div>
                    <div className="mt-2 text-2xl font-bold" style={{ color: 'var(--ds-text)' }}>
                        {proxies.filter(proxy => proxy.username || proxy.has_password).length}
                    </div>
                </div>
            </div>

            <ProxiesTable
                t={t}
                proxies={proxies}
                testing={testing}
                testResults={testResults}
                onCreate={openCreate}
                onTest={testProxy}
                onEdit={openEdit}
                onDelete={deleteProxy}
            />

            <ProxyFormModal
                show={showModal}
                t={t}
                form={form}
                setForm={setForm}
                editingProxy={editingProxy}
                loading={saving}
                onClose={closeModal}
                onSubmit={saveProxy}
            />
        </div>
    )
}