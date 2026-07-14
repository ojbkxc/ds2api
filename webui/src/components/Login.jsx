import { useState } from 'react'
import { Key, ArrowRight, ShieldCheck, Lock, Check } from 'lucide-react'
import clsx from 'clsx'
import { useI18n } from '../i18n'
import LanguageToggle from './LanguageToggle'

export default function Login({ onLogin, onMessage }) {
    const { t } = useI18n()
    const [adminKey, setAdminKey] = useState('')
    const [loading, setLoading] = useState(false)
    const [remember, setRemember] = useState(true)

    const handleLogin = async (e) => {
        e.preventDefault()
        if (!adminKey.trim()) return

        setLoading(true)

        try {
            const res = await fetch('/admin/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ admin_key: adminKey }),
            })

            const data = await res.json()

            if (res.ok && data.success) {
                const storage = remember ? localStorage : sessionStorage
                storage.setItem('ds2api_token', data.token)
                storage.setItem('ds2api_token_expires', Date.now() + data.expires_in * 1000)

                onLogin(data.token)
                if (data.message) {
                    onMessage('warning', data.message)
                }
            } else {
                onMessage('error', data.detail || t('login.signInFailed'))
            }
        } catch (e) {
            onMessage('error', t('login.networkError', { error: e.message }))
        } finally {
            setLoading(false)
        }
    }

    return (
        <div className="min-h-screen w-full flex flex-col items-center justify-center p-4" style={{ background: 'var(--ds-shell-bg)', color: 'var(--ds-text)' }}>
            <div className="absolute top-6 right-6">
                <LanguageToggle />
            </div>

            <div className="w-full max-w-[400px] relative z-10 animate-in fade-in zoom-in-95 duration-200">
                <div className="w-full p-8 animate-in fade-in" style={{ background: 'var(--ds-card)', border: '1px solid var(--ds-border)', borderRadius: 'var(--radius-card)' }}>
                    <div className="text-center space-y-2 mb-8 animate-in fade-in slide-in-from-top-4 duration-500">
                        <div className="inline-flex items-center justify-center w-12 h-12 mb-2" style={{ background: 'var(--ds-blue-light)', borderRadius: 'var(--radius-ctrl)', color: 'var(--ds-blue)' }}>
                            <Lock className="w-6 h-6" />
                        </div>
                        <h1 className="text-[17px] font-bold tracking-tight" style={{ color: 'var(--ds-text)' }}>{t('login.welcome')}</h1>
                        <p className="text-sm" style={{ color: 'var(--ds-text-tertiary)' }}>{t('login.subtitle')}</p>
                    </div>

                    <form onSubmit={handleLogin} className="space-y-5 animate-in fade-in slide-in-from-bottom-4 duration-700 delay-150">
                        <div className="space-y-2">
                            <label className="text-xs font-semibold uppercase tracking-widest ml-1" style={{ color: 'var(--ds-text-secondary)' }}>{t('login.adminKeyLabel')}</label>
                            <div className="relative group">
                                <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none transition-colors" style={{ color: 'var(--ds-text-tertiary)' }}>
                                    <Key className="w-4 h-4" />
                                </div>
                                <input
                                    type="password"
                                    className="w-full pl-10 pr-4 py-3 text-sm transition-all"
                                    style={{
                                        background: 'var(--ds-bg)',
                                        border: '1px solid var(--ds-border)',
                                        borderRadius: 'var(--radius-ctrl)',
                                        color: 'var(--ds-text)',
                                    }}
                                    placeholder={t('login.adminKeyPlaceholder')}
                                    value={adminKey}
                                    onChange={e => setAdminKey(e.target.value)}
                                    autoFocus
                                />
                            </div>
                        </div>

                        <div className="flex items-center justify-between px-1">
                            <label className="flex items-center gap-2.5 cursor-pointer group">
                                <div className="relative flex items-center">
                                    <input
                                        type="checkbox"
                                        className="peer sr-only"
                                        checked={remember}
                                        onChange={e => setRemember(e.target.checked)}
                                    />
                                    <div className="w-[18px] h-[18px] border rounded-md peer-checked:border transition-all shadow-sm"
                                        style={{
                                            background: 'var(--ds-surface)',
                                            borderColor: 'var(--ds-border)',
                                        }}
                                        onMouseEnter={e => {
                                            if (!remember) e.currentTarget.style.borderColor = 'var(--ds-blue)'
                                        }}
                                    >
                                    </div>
                                    <div className="absolute inset-0 flex items-center justify-center" style={{ display: remember ? 'flex' : 'none' }}>
                                        <div className="w-[18px] h-[18px] rounded-md flex items-center justify-center" style={{ background: 'var(--ds-blue)' }}>
                                            <Check className="w-3 h-3 stroke-[3]" style={{ color: 'var(--ds-text-on-primary)' }} />
                                        </div>
                                    </div>
                                    <Check className="absolute inset-0 m-auto w-3 h-3 opacity-0 peer-checked:opacity-100 transition-opacity stroke-[3]" style={{ color: 'var(--ds-text-on-primary)' }} />
                                </div>
                                <span className="text-xs font-medium transition-colors" style={{ color: 'var(--ds-text-secondary)' }}>{t('login.rememberSession')}</span>
                            </label>
                        </div>

                        <button
                            type="submit"
                            disabled={loading}
                            className="w-full h-12 flex items-center justify-center gap-2 transition-all font-semibold text-sm disabled:opacity-50"
                            style={{
                                background: 'var(--ds-blue)',
                                color: 'var(--ds-text-on-primary)',
                                borderRadius: 'var(--radius-ctrl)',
                                boxShadow: 'var(--ds-elevate-1)',
                            }}
                        >
                            {loading ? (
                                <div className="w-5 h-5 border-2 border-t-transparent rounded-full animate-spin" style={{ borderColor: 'var(--ds-text-on-primary)', borderTopColor: 'transparent', opacity: 0.3 }} />
                            ) : (
                                <div className="flex items-center gap-2">
                                    <span>{t('login.signIn')}</span>
                                    <ArrowRight className="w-4 h-4" />
                                </div>
                            )}
                        </button>
                    </form>

                    <div className="mt-6 pt-6 border-t flex justify-center" style={{ borderColor: 'var(--ds-border)' }}>
                        <div className="flex items-center gap-1.5 text-[10px] font-medium tracking-wide uppercase" style={{ color: 'var(--ds-text-tertiary)', opacity: 0.6 }}>
                            <ShieldCheck className="w-3 h-3" />
                            <span>{t('login.secureConnection')}</span>
                        </div>
                    </div>
                </div>

                <div className="mt-8 text-center">
                    <p className="text-[10px] font-mono text-center" style={{ color: 'var(--ds-text-tertiary)', opacity: 0.3 }}>{t('login.adminPortal')}</p>
                </div>
            </div>
        </div>
    )
}