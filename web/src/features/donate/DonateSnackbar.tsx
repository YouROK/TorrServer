import { Button, useMediaQuery } from '@heroui/react'
import { CreditCard, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { useLocalBoolPref } from 'shared/hooks/useLocalPref'
import { queryMax } from 'shared/theme/breakpoints'
import { iconMenu } from 'shared/ui/iconProps'
import { BOTTOM_NAV_TOAST_OFFSET } from 'app/bottomNavLayout'

/** Persist dismiss across sessions — same key as the legacy MUI snackbar. */
const SNACKBAR_CLOSED_KEY = 'snackbarIsClosed'

export interface DonateSnackbarProps {
  onSupport: () => void
}

/**
 * One-shot bottom prompt for project support.
 * Hidden after dismiss or Support (localStorage), and offset above the mobile bottom nav.
 */
export default function DonateSnackbar({ onSupport }: DonateSnackbarProps) {
  const { t } = useTranslation()
  const [closed, setClosed] = useLocalBoolPref(SNACKBAR_CLOSED_KEY, false)
  const hasBottomNav = useMediaQuery(queryMax('mobile'))

  if (closed) return null

  const dismiss = () => setClosed(true)

  return (
    <div
      className='fixed inset-x-0 z-40 flex justify-center px-3'
      style={{
        bottom: hasBottomNav ? BOTTOM_NAV_TOAST_OFFSET : 'calc(12px + env(safe-area-inset-bottom, 0px))',
      }}
      role='status'
    >
      <div className='flex w-full max-w-lg items-center gap-2 rounded-xl border border-border bg-surface/95 px-3 py-2 shadow-lg backdrop-blur'>
        <p className='min-w-0 flex-1 text-sm text-foreground'>{t('Donate?')}</p>
        <Button
          size='sm'
          variant='secondary'
          className='min-h-11 shrink-0 gap-1.5'
          onPress={() => {
            dismiss()
            onSupport()
          }}
        >
          <CreditCard {...iconMenu} aria-hidden />
          {t('Support')}
        </Button>
        <Button
          isIconOnly
          size='sm'
          variant='ghost'
          aria-label={t('Close')}
          onPress={dismiss}
          className='min-h-11 min-w-11 shrink-0'
        >
          <X {...iconMenu} />
        </Button>
      </div>
    </div>
  )
}
