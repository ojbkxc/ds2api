export default function Toggle({ enabled, disabled, onToggle }) {
  return (
    <button
      type="button"
      onClick={() => !disabled && onToggle(!enabled)}
      disabled={disabled}
      aria-pressed={enabled}
      className="ds-switch relative shrink-0 w-10 h-[22px] rounded-full transition-colors duration-200 disabled:opacity-40"
      style={{ background: enabled ? 'var(--ds-blue)' : 'var(--ds-border)' }}
    >
      <span
        className="ds-switch-thumb absolute top-[3px] left-[3px] w-4 h-4 rounded-full transition-transform duration-200"
        style={{ transform: enabled ? 'translateX(18px)' : 'translateX(0)' }}
      />
    </button>
  )
}