import { Button, useMediaQuery } from '@heroui/react'
import { Download, X } from 'lucide-react'
import { iconMenu } from 'shared/ui/iconProps'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { publicUrl } from 'shared/lib/publicUrl'
import { queryMax } from 'shared/theme/breakpoints'
import { BOTTOM_NAV_OFFSET } from 'app/bottomNavLayout'

import { useAndroidInstallPrompt } from './useAndroidInstallPrompt'

/**
 * Android/Chromium install banner driven by `beforeinstallprompt`.
 * Mount only when not iOS and not already standalone — hook no-ops until the browser fires the event.
 */
export default function AndroidInstallBanner() {
  const { t } = useTranslation()
  const { canInstall, promptInstall, dismiss } = useAndroidInstallPrompt()
  const [visible, setVisible] = useState(true)
  const hasBottomNav = useMediaQuery(queryMax('mobile'))

  if (!canInstall || !visible) return null

  const handleDismiss = () => {
    setVisible(false)
    window.setTimeout(() => dismiss(), 300)
  }

  return (
    <div
      className={`fixed inset-x-0 z-40 border-t border-border bg-surface px-4 pt-3 text-foreground shadow-2xl transition-transform duration-300 ${
        visible ? 'translate-y-0' : 'translate-y-[110%]'
      }`}
      style={{
        bottom: hasBottomNav ? BOTTOM_NAV_OFFSET : 0,
        paddingBottom: hasBottomNav ? '12px' : 'calc(12px + env(safe-area-inset-bottom, 0px))',
      }}
    >
      <div className='mb-2 grid grid-cols-[50px_1fr_auto] items-center gap-3'>
        <img src={publicUrl('icon.png')} width={50} height={50} alt='' className='rounded-lg' />
        <p className='text-base font-semibold'>{t('PWAGuide.AndroidHeader')}</p>
        <Button
          isIconOnly
          size='sm'
          variant='ghost'
          aria-label={t('Close')}
          onPress={handleDismiss}
          className='min-h-11 min-w-11'
        >
          <X {...iconMenu} />
        </Button>
      </div>

      <p className='mb-3 text-sm text-muted'>{t('PWAGuide.AndroidDescription')}</p>

      <Button variant='primary' className='mb-1 min-h-11 w-full gap-2' onPress={() => void promptInstall()}>
        <Download {...iconMenu} aria-hidden />
        {t('PWAGuide.Install')}
      </Button>
    </div>
  )
}
