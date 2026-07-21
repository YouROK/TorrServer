import { Button, useMediaQuery } from '@heroui/react'
import { X } from 'lucide-react'
import { iconMenu } from 'shared/ui/iconProps'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { readLocalBool, writeLocalJson } from 'shared/lib/localPrefs'
import { publicUrl } from 'shared/lib/publicUrl'
import { queryMax } from 'shared/theme/breakpoints'
import { BOTTOM_NAV_OFFSET } from 'app/bottomNavLayout'

const CLOSED_PREF_KEY = 'pwaNotificationIsClosed'

function IOSShareIcon() {
  return (
    <svg width={20} height={20} viewBox='0 0 1000 1000' fill='#005FF2' aria-hidden className='inline align-middle'>
      <path d='M780,290H640v35h140c19.3,0,35,15.7,35,35v560c0,19.3-15.7,35-35,35H220c-19.2,0-35-15.7-35-35V360c0-19.2,15.7-35,35-35h140v-35H220c-38.7,0-70,31.3-70,70v560c0,38.7,31.3,70,70,70h560c38.7,0,70-31.3,70-70V360C850,321.3,818.7,290,780,290z M372.5,180l110-110.2v552.7c0,9.6,7.9,17.5,17.5,17.5c9.6,0,17.5-7.9,17.5-17.5V69.8l110,110c3.5,3.5,7.9,5,12.5,5s9-1.7,12.5-5c6.8-6.8,6.8-17.9,0-24.7l-140-140c-6.8-6.8-17.9-6.8-24.7,0l-140,140c-6.8,6.8-6.8,17.9,0,24.7C354.5,186.8,365.5,186.8,372.5,180z' />
    </svg>
  )
}

/**
 * iOS "Add to Home Screen" install nudge. The caller (Shell) only mounts this when
 * `detectApplePlatform().isIOS && !isStandaloneApp` — this component just owns the
 * dismiss-and-remember-forever state on top of that gate.
 */
export default function PWAInstallationGuide() {
  const { t } = useTranslation()
  const [isOpen, setIsOpen] = useState(() => !readLocalBool(CLOSED_PREF_KEY))
  const [visible, setVisible] = useState(() => !readLocalBool(CLOSED_PREF_KEY))
  /** Shell only renders BottomNav below this breakpoint — sit above it there instead of covering it. */
  const hasBottomNav = useMediaQuery(queryMax('mobile'))

  if (!isOpen) return null

  const dismiss = () => {
    setVisible(false)
    window.setTimeout(() => {
      setIsOpen(false)
      writeLocalJson(CLOSED_PREF_KEY, true)
    }, 300)
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
        <p className='text-base font-semibold'>{t('PWAGuide.Header')}</p>
        <Button
          isIconOnly
          size='sm'
          variant='ghost'
          aria-label={t('Close')}
          onPress={dismiss}
          className='min-h-11 min-w-11'
        >
          <X {...iconMenu} />
        </Button>
      </div>

      <p className='mb-2 text-sm text-muted'>{t('PWAGuide.Description')}</p>
      <p className='mb-2 text-sm text-muted'>{t('PWAGuide.PlayerButtons')}</p>
      <p className='mb-1 text-sm'>
        1. {t('PWAGuide.FirstStep')} <IOSShareIcon />
      </p>
      <p className='pb-1 text-sm'>
        2. {t('PWAGuide.SecondStep.Select')}
        <span className='font-bold text-accent'> {t('PWAGuide.SecondStep.AddToHomeScreen')}</span>
      </p>
    </div>
  )
}
