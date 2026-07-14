const STATUS_MAP = {
  active:    { color: 'var(--ds-success)', shadow: '0 0 6px var(--ds-success)' },
  failed:    { color: 'var(--ds-danger)',  shadow: '0 0 6px var(--ds-danger)' },
  banned:    { color: 'var(--ds-purple)',  shadow: '0 0 6px var(--ds-purple)' },
  unknown:   { color: 'var(--ds-info)',    shadow: '0 0 6px var(--ds-info)' },
  disabled:  { color: 'var(--ds-text-tertiary)', shadow: '0 0 6px var(--ds-text-tertiary)' },
  pending:   { color: 'var(--ds-warning)', shadow: '0 0 6px var(--ds-warning)' },
}

export default function StatusDot({ status, pulse = false }) {
  const s = STATUS_MAP[status] || STATUS_MAP.unknown
  return (
    <span
      className={`ds-status-dot ${pulse ? 'animate-pulse' : ''}`}
      style={{ backgroundColor: s.color, boxShadow: s.shadow }}
    />
  )
}