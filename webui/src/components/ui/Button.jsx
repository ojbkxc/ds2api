import clsx from 'clsx'

const variants = {
  primary: 'ds-btn-primary',
  secondary: 'ds-btn-secondary',
  danger: 'ds-btn-danger',
}

const sizes = {
  sm: 'px-2.5 py-1.5 text-[11px]',
  md: 'px-4 py-2 text-[13px]',
  lg: 'px-5 py-2.5 text-[14px]',
}

export default function Button({ variant = 'primary', size = 'md', disabled, className, children, ...props }) {
  return (
    <button
      className={clsx(variants[variant], sizes[size], className)}
      disabled={disabled}
      {...props}
    >
      {children}
    </button>
  )
}