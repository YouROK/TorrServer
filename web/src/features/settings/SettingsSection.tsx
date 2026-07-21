import type { ReactNode } from 'react'

export interface SettingsSectionProps {
  icon?: ReactNode
  title: string
  description?: string
  children: ReactNode
  className?: string
}

/**
 * Groups related settings behind a heading + card, so a tab reads as a handful of labeled
 * sections instead of one long undifferentiated column of switches/fields.
 */
export default function SettingsSection({ icon, title, description, children, className }: SettingsSectionProps) {
  return (
    <section className={className}>
      <div className='mb-2 flex items-start gap-2'>
        {icon ? <span className='mt-0.5 shrink-0 text-muted [&>svg]:size-4'>{icon}</span> : null}
        <h3 className='text-xs font-semibold tracking-wide text-muted uppercase'>{title}</h3>
      </div>
      {description ? <p className='mb-3 max-w-prose text-sm leading-relaxed text-muted'>{description}</p> : null}
      <div className='space-y-4 rounded-xl border border-border bg-surface-secondary p-3 sm:p-4'>{children}</div>
    </section>
  )
}
