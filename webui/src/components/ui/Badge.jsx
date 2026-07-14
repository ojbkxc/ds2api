import clsx from 'clsx'

const toneMap = {
  success: 'ds-badge ds-badge-success',
  warning: 'ds-badge ds-badge-warning',
  info: 'ds-badge ds-badge-info',
  purple: 'ds-badge ds-badge-purple',
  danger: 'ds-badge ds-badge-danger',
  muted: 'ds-badge',
}

export default function Badge({ tone = 'muted', className, children }) {
  return (
    <span className={clsx(toneMap[tone], className)}>
      {children}
    </span>
  )
}