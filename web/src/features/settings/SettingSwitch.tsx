import { Switch } from '@heroui/react'

/** Keys whose on-disk value is the inverse of what the switch shows (switch ON = feature enabled). */
export const DISABLE_SWITCH_IDS = new Set([
  'DisableTCP',
  'DisableUTP',
  'DisableUPNP',
  'DisableDHT',
  'DisablePEX',
  'DisableUpload',
])

export interface SettingSwitchProps {
  id: string
  label: string
  helper?: string
  checked: boolean
  onChange: (id: string, checked: boolean) => void
}

/**
 * Labeled toggle row. Uses plain block text (not HeroUI Label+Description inline pairing) so the
 * title and helper never run together on one line — that was especially visible in Network protocols.
 */
export function SettingSwitch({ id, label, helper, checked, onChange }: SettingSwitchProps) {
  return (
    <div className='flex min-h-12 items-start justify-between gap-4 py-1.5'>
      <div className='min-w-0 flex-1'>
        <label htmlFor={id} className='block text-sm font-medium leading-snug text-foreground'>
          {label}
        </label>
        {helper ? <p className='mt-1.5 block text-sm leading-relaxed text-muted'>{helper}</p> : null}
      </div>
      <Switch id={id} isSelected={checked} onChange={value => onChange(id, value)} className='mt-0.5 shrink-0'>
        <Switch.Content>
          <Switch.Control>
            <Switch.Thumb />
          </Switch.Control>
        </Switch.Content>
      </Switch>
    </div>
  )
}
