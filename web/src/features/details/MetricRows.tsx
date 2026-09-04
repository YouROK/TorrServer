import type { ReactNode } from 'react'

export interface MetricRowItem {
  label: string
  value: string
  /** Optional tooltip clarifying metric semantics (e.g. Cache vs Loaded). */
  hint?: string
}

export interface MetricRowsProps {
  items: MetricRowItem[]
  /** Optional uppercase section heading. */
  title?: string
  /** Dense multi-column definition list (mobile secondary/swarm). */
  columns?: 1 | 2
  /** Tighter row padding for full Swarm tab. */
  dense?: boolean
  /** Wrap in a single bordered panel (default true). */
  framed?: boolean
  className?: string
  children?: ReactNode
}

/**
 * Dense Stats-tab metrics: label left, value right — no nested chip cards.
 */
export default function MetricRows({
  items,
  title,
  columns = 1,
  dense = false,
  framed = true,
  className = '',
  children,
}: MetricRowsProps) {
  const list = (
    <div className={columns === 2 ? 'grid grid-cols-2 gap-x-3 gap-y-0.5' : 'flex flex-col'}>
      {items.map(item => (
        <div
          key={item.label}
          className={`flex min-w-0 items-baseline justify-between gap-2 text-sm ${dense ? 'py-0.5' : 'py-1'}`}
          title={item.hint}
        >
          <span className='min-w-0 truncate text-muted' title={item.hint || item.label}>
            {item.label}
          </span>
          <span
            className='max-w-[55%] shrink-0 truncate text-right font-bold tabular-nums text-foreground'
            title={item.value}
          >
            {item.value}
          </span>
        </div>
      ))}
    </div>
  )

  const body = (
    <>
      {title ? <p className='mb-1.5 text-xs font-semibold tracking-wide text-muted uppercase'>{title}</p> : null}
      {list}
      {children}
    </>
  )

  if (!framed) {
    return <div className={className || undefined}>{body}</div>
  }

  return <div className={`rounded-xl border border-border bg-surface-secondary p-2.5 ${className}`.trim()}>{body}</div>
}
