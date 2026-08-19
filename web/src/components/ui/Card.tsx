import type { ReactNode } from 'react'

export default function Card({
  title,
  actions,
  children,
  className = '',
}: {
  title?: string
  actions?: ReactNode
  children: ReactNode
  className?: string
}) {
  return (
    <section
      aria-label={title}
      className={`card-glow rounded-2xl border border-slate-200/80 bg-white/80 p-4 backdrop-blur dark:border-slate-800 dark:bg-slate-900/70 ${className}`}
    >
      {(title || actions) && (
        <header className="mb-3 flex flex-wrap items-center justify-between gap-2">
          {title && <h3 className="text-sm font-bold uppercase tracking-wide text-slate-500 dark:text-slate-400">{title}</h3>}
          {actions}
        </header>
      )}
      {children}
    </section>
  )
}