import { Suspense, lazy, useCallback, useEffect, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import {
    LayoutDashboard,
    Upload,
    Cloud,
    Settings as SettingsIcon,
    LogOut,
    Menu,
    X,
    Server,
    Users,
    Globe,
    History,
    Loader2
} from 'lucide-react'
import clsx from 'clsx'

import LanguageToggle from '../components/LanguageToggle'
import { useI18n } from '../i18n'

const AccountManagerContainer = lazy(() => import('../features/account/AccountManagerContainer'))
const ApiTesterContainer = lazy(() => import('../features/apiTester/ApiTesterContainer'))
const ChatHistoryContainer = lazy(() => import('../features/chatHistory/ChatHistoryContainer'))
const BatchImport = lazy(() => import('../components/BatchImport'))
const VercelSyncContainer = lazy(() => import('../features/vercel/VercelSyncContainer'))
const SettingsContainer = lazy(() => import('../features/settings/SettingsContainer'))
const ProxyManagerContainer = lazy(() => import('../features/proxy/ProxyManagerContainer'))

function TabLoadingFallback({ label }) {
    return (
        <div className="min-h-[320px] border border-border flex items-center justify-center" style={{ background: 'var(--ds-card)', borderRadius: 'var(--radius-card)' }}>
            <div className="flex items-center gap-3 text-sm" style={{ color: 'var(--ds-text-secondary)' }}>
                <Loader2 className="w-4 h-4 animate-spin" />
                <span>{label}</span>
            </div>
        </div>
    )
}

export default function DashboardShell({ token, onLogout, config, fetchConfig, showMessage, message, onForceLogout, isVercel }) {
    const { t } = useI18n()
    const location = useLocation()
    const navigate = useNavigate()
    const [sidebarOpen, setSidebarOpen] = useState(false)

    const navItems = [
        { id: 'accounts', label: t('nav.accounts.label'), icon: Users, description: t('nav.accounts.desc') },
        { id: 'proxies', label: t('nav.proxies.label'), icon: Globe, description: t('nav.proxies.desc') },
        { id: 'test', label: t('nav.test.label'), icon: Server, description: t('nav.test.desc') },
        { id: 'history', label: t('nav.history.label'), icon: History, description: t('nav.history.desc') },
        { id: 'import', label: t('nav.import.label'), icon: Upload, description: t('nav.import.desc') },
        { id: 'vercel', label: t('nav.vercel.label'), icon: Cloud, description: t('nav.vercel.desc') },
        { id: 'settings', label: t('nav.settings.label'), icon: SettingsIcon, description: t('nav.settings.desc') },
    ]

    const tabIds = new Set(navItems.map(item => item.id))
    const pathSegments = location.pathname.replace(/^\/+|\/+$/g, '').split('/').filter(Boolean)
    const routeSegments = pathSegments[0] === 'admin' ? pathSegments.slice(1) : pathSegments
    const pathTab = routeSegments[0] || ''
    const activeTab = tabIds.has(pathTab) ? pathTab : 'accounts'
    const adminBasePath = pathSegments[0] === 'admin' ? '/admin' : ''
    const activeNavItem = navItems.find(n => n.id === activeTab)

    const navigateToTab = useCallback((tabID) => {
        const nextPath = tabID === 'accounts'
            ? `${adminBasePath || ''}/`
            : `${adminBasePath}/${tabID}`
        navigate(nextPath)
        setSidebarOpen(false)
    }, [adminBasePath, navigate])

    const authFetch = useCallback(async (url, options = {}) => {
        const headers = {
            ...options.headers,
            'Authorization': `Bearer ${token}`
        }
        const res = await fetch(url, { ...options, headers })

        if (res.status === 401) {
            onLogout()
            throw new Error(t('auth.expired'))
        }
        return res
    }, [onLogout, t, token])


    const [versionInfo, setVersionInfo] = useState(null)

    useEffect(() => {
        let disposed = false
        async function loadVersion() {
            try {
                const res = await authFetch('/admin/version')
                const data = await res.json()
                if (!disposed) {
                    setVersionInfo(data)
                }
            } catch (_err) {
                if (!disposed) {
                    setVersionInfo(null)
                }
            }
        }
        loadVersion()
        return () => {
            disposed = true
        }
    }, [authFetch])
    const renderTab = () => {
        switch (activeTab) {
            case 'accounts':
                return <AccountManagerContainer config={config} onRefresh={fetchConfig} onMessage={showMessage} authFetch={authFetch} />
            case 'proxies':
                return <ProxyManagerContainer config={config} onRefresh={fetchConfig} onMessage={showMessage} authFetch={authFetch} />
            case 'test':
                return <ApiTesterContainer config={config} onMessage={showMessage} authFetch={authFetch} />
            case 'history':
                return <ChatHistoryContainer onMessage={showMessage} authFetch={authFetch} />
            case 'import':
                return <BatchImport onRefresh={fetchConfig} onMessage={showMessage} authFetch={authFetch} />
            case 'vercel':
                return <VercelSyncContainer onMessage={showMessage} authFetch={authFetch} isVercel={isVercel} config={config} />
            case 'settings':
                return <SettingsContainer onRefresh={fetchConfig} onMessage={showMessage} authFetch={authFetch} onForceLogout={onForceLogout} isVercel={isVercel} />
            default:
                return null
        }
    }

    return (
        <div className="flex h-screen overflow-hidden" style={{ background: 'var(--ds-shell-bg)', color: 'var(--ds-text)' }}>
            {sidebarOpen && (
                <div
                    className="fixed inset-0 z-40 lg:hidden"
                    style={{ background: 'rgba(15,20,35,0.6)', backdropFilter: 'blur(2px)' }}
                    onClick={() => setSidebarOpen(false)}
                />
            )}

            <aside className={clsx(
                "fixed lg:static inset-y-0 left-0 z-50 w-64 border-r transition-transform duration-300 ease-in-out lg:transform-none flex flex-col",
                sidebarOpen ? "translate-x-0" : "-translate-x-full"
            )}
                style={{ background: 'var(--ds-card)', borderColor: 'var(--ds-border)' }}>
                <div className="p-6">
                    <div className="flex items-center gap-2.5 font-bold text-xl tracking-tight" style={{ color: 'var(--ds-text)' }}>
                        <div className="w-8 h-8 flex items-center justify-center" style={{ background: 'var(--ds-blue)', borderRadius: 'var(--radius-ctrl)', color: 'var(--ds-text-on-primary)' }}>
                            <LayoutDashboard className="w-5 h-5" />
                        </div>
                        <span>DS2API</span>
                    </div>
                    <div className="flex items-center justify-between mt-2">
                        <p className="text-[10px] font-semibold tracking-[0.1em] uppercase opacity-60 px-1" style={{ color: 'var(--ds-text-tertiary)' }}>{t('sidebar.onlineAdminConsole')}</p>
                        <LanguageToggle />
                    </div>
                </div>

                <nav className="flex-1 px-3 space-y-1 overflow-y-auto pt-2">
                    {navItems.map((item) => {
                        const Icon = item.icon
                        const isActive = activeTab === item.id
                        return (
                            <button
                                key={item.id}
                                onClick={() => {
                                    navigateToTab(item.id)
                                }}
                                className={clsx(
                                    "w-full flex items-center gap-3 px-3 py-2.5 text-sm font-medium transition-all duration-200 group relative",
                                )}
                                style={{
                                    borderRadius: 'var(--radius-ctrl)',
                                    background: isActive ? 'var(--ds-surface)' : 'transparent',
                                    color: isActive ? 'var(--ds-text)' : 'var(--ds-text-secondary)',
                                    border: isActive ? '1px solid var(--ds-border)' : '1px solid transparent',
                                }}
                                onMouseEnter={e => {
                                    if (!isActive) {
                                        e.currentTarget.style.color = 'var(--ds-text)'
                                        e.currentTarget.style.background = 'var(--ds-surface)'
                                    }
                                }}
                                onMouseLeave={e => {
                                    if (!isActive) {
                                        e.currentTarget.style.color = 'var(--ds-text-secondary)'
                                        e.currentTarget.style.background = 'transparent'
                                    }
                                }}
                            >
                                <Icon className={clsx("w-4 h-4 transition-colors")} style={{ color: isActive ? 'var(--ds-blue)' : 'var(--ds-text-secondary)' }} />
                                <span className="flex-1 text-left">{item.label}</span>
                                {isActive && (
                                    <div className="absolute bottom-0 left-3 right-3 h-0.5" style={{ background: 'var(--ds-blue)', borderRadius: '2px 2px 0 0' }} />
                                )}
                            </button>
                        )
                    })}
                </nav>

                <div className="p-4 border-t" style={{ borderColor: 'var(--ds-border)', background: 'var(--ds-card)' }}>
                    <div className="space-y-4">
                        <div className="flex items-center justify-between text-sm px-1">
                            <span className="font-semibold text-[10px] uppercase tracking-wider" style={{ color: 'var(--ds-text-secondary)' }}>{t('sidebar.systemStatus')}</span>
                            <span className="flex items-center gap-1.5 text-[10px] font-bold" style={{ color: 'var(--ds-success)', background: 'var(--ds-success-bg)', borderRadius: 'var(--radius-pill)', padding: '0.125rem 0.5rem', border: '1px solid var(--ds-success-border)' }}>
                                <span className="w-1.5 h-1.5 rounded-full animate-pulse" style={{ background: 'var(--ds-success)' }}></span>
                                {t('sidebar.statusOnline')}
                            </span>
                        </div>
                        <div className="grid grid-cols-2 gap-2">
                            <div className="p-3 border" style={{ background: 'var(--ds-bg)', borderRadius: 'var(--radius-card)', borderColor: 'var(--ds-border)' }}>
                                <div className="text-[9px] font-bold uppercase tracking-wider mb-0.5 opacity-70" style={{ color: 'var(--ds-text-tertiary)' }}>{t('sidebar.accounts')}</div>
                                <div className="text-lg font-bold leading-tight" style={{ color: 'var(--ds-text)' }}>{config.accounts?.length || 0}</div>
                            </div>
                            <div className="p-3 border" style={{ background: 'var(--ds-bg)', borderRadius: 'var(--radius-card)', borderColor: 'var(--ds-border)' }}>
                                <div className="text-[9px] font-bold uppercase tracking-wider mb-0.5 opacity-70" style={{ color: 'var(--ds-text-tertiary)' }}>{t('sidebar.keys')}</div>
                                <div className="text-lg font-bold" style={{ color: 'var(--ds-text)' }}>{config.keys?.length || 0}</div>
                            </div>
                        </div>
                        <div className="p-3 border" style={{ background: 'var(--ds-bg)', borderRadius: 'var(--radius-card)', borderColor: 'var(--ds-border)' }}>
                            <div className="text-[9px] font-bold uppercase tracking-wider mb-1 opacity-70" style={{ color: 'var(--ds-text-tertiary)' }}>{t('sidebar.version')}</div>
                            <div className="text-xs font-semibold" style={{ color: 'var(--ds-text)' }}>{versionInfo?.current_tag || '-'}</div>
                            {versionInfo?.has_update && (
                                <a
                                    className="inline-flex mt-1 text-[10px] hover:underline"
                                    style={{ color: 'var(--ds-warning)' }}
                                    href={versionInfo?.release_url || 'https://github.com/ojbkxc/ds2api/releases/latest'}
                                    target="_blank"
                                    rel="noreferrer"
                                >
                                    {t('sidebar.updateAvailable', { latest: versionInfo.latest_tag || '' })}
                                </a>
                            )}
                        </div>
                        <button
                            onClick={onLogout}
                            className="w-full h-10 flex items-center justify-center gap-2 text-xs font-medium transition-all"
                            style={{ borderRadius: 'var(--radius-ctrl)', border: '1px solid var(--ds-border)', color: 'var(--ds-text-secondary)', background: 'transparent' }}
                            onMouseEnter={e => {
                                e.currentTarget.style.color = 'var(--ds-danger)'
                                e.currentTarget.style.borderColor = 'var(--ds-danger-border)'
                                e.currentTarget.style.background = 'var(--ds-danger-bg)'
                            }}
                            onMouseLeave={e => {
                                e.currentTarget.style.color = 'var(--ds-text-secondary)'
                                e.currentTarget.style.borderColor = 'var(--ds-border)'
                                e.currentTarget.style.background = 'transparent'
                            }}
                        >
                            <LogOut className="w-3.5 h-3.5" />
                            {t('sidebar.signOut')}
                        </button>
                    </div>
                </div>
            </aside>

            <main className="flex-1 flex flex-col min-w-0 overflow-hidden relative">
                <header className="lg:hidden h-14 flex items-center justify-between px-4 border-b" style={{ borderColor: 'var(--ds-border)', background: 'var(--ds-card)' }}>
                    <div className="flex items-center gap-2">
                        <div className="w-6 h-6 flex items-center justify-center" style={{ background: 'var(--ds-blue)', borderRadius: 'var(--radius-ctrl)', color: 'var(--ds-text-on-primary)' }}>
                            <LayoutDashboard className="w-3.5 h-3.5" />
                        </div>
                        <span className="font-semibold text-sm" style={{ color: 'var(--ds-text)' }}>DS2API</span>
                    </div>
                    <div className="flex items-center gap-2">
                        <LanguageToggle />
                        <button
                            onClick={() => setSidebarOpen(true)}
                            className="p-2 -mr-2"
                            style={{ color: 'var(--ds-text-secondary)' }}
                        >
                            <Menu className="w-5 h-5" />
                        </button>
                    </div>
                </header>

                <div className="flex-1 overflow-auto p-4 md:p-6 lg:p-8" style={{ background: 'var(--ds-shell-bg)' }}>
                    <div className="max-w-6xl mx-auto space-y-4 lg:space-y-6">
                        <div className="hidden lg:block mb-8">
                            <h1 className="text-[17px] font-bold tracking-tight mb-2" style={{ color: 'var(--ds-text)' }}>
                                {activeNavItem?.label}
                            </h1>
                            <p style={{ color: 'var(--ds-text-secondary)', fontSize: '13px' }}>
                                {activeNavItem?.description}
                            </p>
                        </div>

                        {message && (
                            <div className={clsx(
                                "p-4 flex items-center gap-3 animate-in fade-in slide-in-from-top-2",
                            )}
                                style={{
                                    borderRadius: 'var(--radius-ctrl)',
                                    border: message.type === 'error'
                                        ? '1px solid var(--ds-danger-border)'
                                        : '1px solid var(--ds-success-border)',
                                    background: message.type === 'error'
                                        ? 'var(--ds-danger-bg)'
                                        : 'var(--ds-success-bg)',
                                    color: message.type === 'error'
                                        ? 'var(--ds-danger)'
                                        : 'var(--ds-success)',
                                }}
                            >
                                {message.type === 'error' ? <X className="w-5 h-5" /> : <div className="w-5 h-5 rounded-full border-2 flex items-center justify-center text-[10px]" style={{ borderColor: 'var(--ds-success)' }}>✓</div>}
                                {message.text}
                            </div>
                        )}

                        <div className="animate-in fade-in duration-500">
                            <Suspense fallback={<TabLoadingFallback label={activeNavItem?.label || 'DS2API'} />}>
                                {renderTab()}
                            </Suspense>
                        </div>
                    </div>
                </div>
            </main>
        </div>
    )
}