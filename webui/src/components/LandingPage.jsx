import React from 'react'
import { useI18n } from '../i18n'
import LanguageToggle from './LanguageToggle'

const LandingPage = ({ onEnter }) => {
    const { t } = useI18n()
    return (
        <div className="min-h-screen relative overflow-hidden flex flex-col items-center justify-center p-6 text-center" style={{ background: 'var(--ds-shell-bg)', color: 'var(--ds-text)', fontFamily: 'inherit' }}>
            {/* Background glow */}
            <div style={{
                position: 'fixed', top: 0, left: 0, width: '100%', height: '100%', zIndex: 0,
                background: 'radial-gradient(circle at 20% 30%, var(--ds-blue-glow) 0%, transparent 40%), radial-gradient(circle at 80% 70%, var(--ds-blue-glow) 0%, transparent 40%)'
            }} />

            {/* Blobs */}
            <div style={{
                position: 'absolute', width: '400px', height: '400px',
                background: 'var(--ds-blue)',
                filter: 'blur(80px)', opacity: 0.12, borderRadius: '50%', zIndex: 0,
                top: '10%', left: '15%',
                animation: 'move 20s infinite alternate'
            }} />
            <div style={{
                position: 'absolute', width: '400px', height: '400px',
                background: 'var(--ds-blue-soft)',
                filter: 'blur(80px)', opacity: 0.10, borderRadius: '50%', zIndex: 0,
                bottom: '10%', right: '15%',
                animation: 'move 20s infinite alternate',
                animationDelay: '-5s'
            }} />

            <style>{`
                @keyframes move {
                    from { transform: translate(-10%, -10%) scale(1); }
                    to { transform: translate(10%, 10%) scale(1.1); }
                }
                @keyframes fadeInUp {
                    from { opacity: 0; transform: translateY(20px); }
                    to { opacity: 1; transform: translateY(0); }
                }
            `}</style>

            <div className="absolute top-6 right-6 z-20">
                <LanguageToggle />
            </div>

            <div style={{ position: 'relative', zIndex: 10, maxWidth: '900px', animation: 'fadeInUp 0.8s ease-out' }}>
                <header className="mb-12">
                    <h1 style={{
                        fontSize: 'clamp(3rem, 10vw, 5rem)',
                        fontWeight: 700,
                        color: 'var(--ds-blue)',
                        letterSpacing: '-2px',
                        marginBottom: '0.5rem',
                    }}>
                        DS2API
                    </h1>
                    <p className="text-xl max-w-2xl mx-auto leading-relaxed" style={{ color: 'var(--ds-text-secondary)' }}>
                        DeepSeek to OpenAI & Claude Compatible API Interface
                    </p>
                </header>

                <div className="flex flex-wrap gap-4 justify-center mb-16">
                    <button
                        onClick={onEnter}
                        className="px-8 py-3 font-bold transition-all flex items-center gap-2"
                        style={{
                            background: 'var(--ds-blue)',
                            color: 'var(--ds-text-on-primary)',
                            borderRadius: 'var(--radius-ctrl)',
                            border: '1px solid var(--ds-blue)',
                            boxShadow: 'var(--ds-elevate-1)',
                        }}
                        onMouseEnter={e => {
                            e.currentTarget.style.background = 'var(--ds-blue-hover)'
                            e.currentTarget.style.transform = 'translateY(-1px)'
                        }}
                        onMouseLeave={e => {
                            e.currentTarget.style.background = 'var(--ds-blue)'
                            e.currentTarget.style.transform = 'none'
                        }}
                    >
                        <span>{t('landing.adminConsole')}</span>
                    </button>
                    <a
                        href="/v1/models"
                        target="_blank"
                        className="px-8 py-3 font-semibold transition-all flex items-center gap-2"
                        style={{
                            background: 'transparent',
                            border: '1px solid var(--ds-border)',
                            borderRadius: 'var(--radius-ctrl)',
                            color: 'var(--ds-text-secondary)',
                        }}
                        onMouseEnter={e => {
                            e.currentTarget.style.borderColor = 'var(--ds-blue)'
                            e.currentTarget.style.color = 'var(--ds-text)'
                        }}
                        onMouseLeave={e => {
                            e.currentTarget.style.borderColor = 'var(--ds-border)'
                            e.currentTarget.style.color = 'var(--ds-text-secondary)'
                        }}
                    >
                        <span>{t('landing.apiStatus')}</span>
                    </a>
                    <a
                        href="https://github.com/ojbkxc/ds2api"
                        target="_blank"
                        className="px-8 py-3 font-semibold transition-all flex items-center gap-2"
                        style={{
                            background: 'transparent',
                            border: '1px solid var(--ds-border)',
                            borderRadius: 'var(--radius-ctrl)',
                            color: 'var(--ds-text-secondary)',
                        }}
                        onMouseEnter={e => {
                            e.currentTarget.style.borderColor = 'var(--ds-blue)'
                            e.currentTarget.style.color = 'var(--ds-text)'
                        }}
                        onMouseLeave={e => {
                            e.currentTarget.style.borderColor = 'var(--ds-border)'
                            e.currentTarget.style.color = 'var(--ds-text-secondary)'
                        }}
                    >
                        <span>GitHub</span>
                    </a>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-6 text-left">
                    {[
                        { icon: '🚀', title: t('landing.features.compatibility.title'), desc: t('landing.features.compatibility.desc') },
                        { icon: '⚖️', title: t('landing.features.loadBalancing.title'), desc: t('landing.features.loadBalancing.desc') },
                        { icon: '🧠', title: t('landing.features.reasoning.title'), desc: t('landing.features.reasoning.desc') },
                        { icon: '🔍', title: t('landing.features.search.title'), desc: t('landing.features.search.desc') },
                    ].map((feature, idx) => (
                        <div key={idx} className="p-6 transition-all"
                            style={{
                                background: 'var(--ds-card)',
                                border: '1px solid var(--ds-border)',
                                borderRadius: 'var(--radius-card)',
                            }}
                            onMouseEnter={e => {
                                e.currentTarget.style.borderColor = 'var(--ds-blue)'
                                e.currentTarget.style.background = 'var(--ds-surface)'
                                e.currentTarget.style.transform = 'translateY(-5px)'
                            }}
                            onMouseLeave={e => {
                                e.currentTarget.style.borderColor = 'var(--ds-border)'
                                e.currentTarget.style.background = 'var(--ds-card)'
                                e.currentTarget.style.transform = 'none'
                            }}
                        >
                            <span className="text-2xl mb-4 block">{feature.icon}</span>
                            <h3 className="text-lg font-bold mb-2" style={{ color: 'var(--ds-text)' }}>{feature.title}</h3>
                            <p className="text-sm leading-relaxed" style={{ color: 'var(--ds-text-secondary)' }}>{feature.desc}</p>
                        </div>
                    ))}
                </div>

                <footer className="mt-20 opacity-40 text-sm">
                    <p>&copy; 2026 DS2API Project. Designed for flexibility & performance.</p>
                </footer>
            </div>
        </div>
    )
}

export default LandingPage
