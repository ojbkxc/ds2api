import { useI18n } from '../i18n'

export default function LanguageToggle({ className = '' }) {
    const { lang, setLang, t } = useI18n()
    const nextLang = lang === 'zh' ? 'en' : 'zh'
    const label = nextLang === 'zh' ? t('language.chinese') : t('language.english')

    return (
        <button
            type="button"
            onClick={() => setLang(nextLang)}
            className={`text-[10px] font-semibold px-2 py-1 transition-colors ${className}`}
            style={{
                background: 'var(--ds-surface)',
                border: '1px solid var(--ds-border)',
                borderRadius: 'var(--radius-ctrl)',
                color: 'var(--ds-text-secondary)',
            }}
            title={t('language.label')}
        >
            {label}
        </button>
    )
}