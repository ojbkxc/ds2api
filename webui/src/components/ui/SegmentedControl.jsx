export default function SegmentedControl({ options, value, onChange, ariaLabel, size = 'md' }) {
  const padding = size === 'sm' ? 'px-2 py-1 text-[11px]' : 'px-2.5 py-1.5 text-[11px]'
  return (
    <div className="ds-segmented" role="radiogroup" aria-label={ariaLabel}>
      {options.map((option) => {
        const active = option.key === value
        return (
          <button
            key={option.key}
            type="button"
            role="radio"
            aria-checked={active}
            onClick={() => onChange(option.key)}
            className={`${padding} font-medium border transition-colors duration-150`}
            style={{
              borderRadius: 'var(--radius-ctrl)',
              background: active ? 'var(--ds-blue)' : 'transparent',
              color: active ? 'var(--ds-text-on-primary)' : 'var(--ds-text-secondary)',
              borderColor: active ? 'var(--ds-blue)' : 'var(--ds-border)',
            }}
          >
            {option.label}
          </button>
        )
      })}
    </div>
  )
}