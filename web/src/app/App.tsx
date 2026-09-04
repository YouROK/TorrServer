import { type ReactNode, useEffect } from 'react'
import { useMediaQuery } from '@heroui/react'
import { Toaster } from 'sonner'

import { useSettingsQuery } from 'shared/hooks/useSettingsQuery'
import { queryMax } from 'shared/theme/breakpoints'
import { useThemePreference } from 'shared/theme/useThemePreference'
import { ModalOpenProvider } from 'shared/ui/ModalOpenContext'
import { AppSnackbarProvider } from 'shared/ui/Toast'
import AuthGate from 'features/auth/AuthGate'
import PwaUpdateToast from 'features/pwa/PwaUpdateToast'

import Shell from './Shell'
import { BOTTOM_NAV_TOAST_OFFSET } from './bottomNavLayout'

/** Keep SETTINGS_QUERY_KEY warm so play/resume can read TrackTimecode without opening Settings. */
function SettingsQueryBootstrap() {
  useSettingsQuery()
  return null
}

/** Apply `dark` / `data-palette` on `<html>` for the whole session (prefs live in localStorage). */
function ThemeBootstrap() {
  const { isDark, palette } = useThemePreference()
  useEffect(() => {
    // Module-load apply can run before <body>; re-sync surface paint for iOS home-indicator.
    const surface = getComputedStyle(document.documentElement).getPropertyValue('--surface').trim()
    if (surface) {
      document.documentElement.style.backgroundColor = surface
      document.body.style.backgroundColor = surface
    }
  }, [isDark, palette])
  return null
}

function AuthedApp({ children }: { children: ReactNode }) {
  return (
    <div className='flex min-h-0 flex-1 flex-col'>
      <SettingsQueryBootstrap />
      {children}
    </div>
  )
}

/**
 * Root providers: modal-open chrome, snackbars, auth gate, settings cache bootstrap, Shell.
 * Toaster offset clears the mobile BottomNav when present.
 * `--app-height` is installed from `index.tsx` before first paint.
 */
export default function App() {
  const hasBottomNav = useMediaQuery(queryMax('mobile'))

  return (
    <div className='flex min-h-0 flex-1 flex-col'>
      <ModalOpenProvider>
        <AppSnackbarProvider>
          <ThemeBootstrap />
          <AuthGate>
            <AuthedApp>
              <Shell />
            </AuthedApp>
          </AuthGate>
          <PwaUpdateToast />
          <Toaster
            richColors
            closeButton
            position='bottom-center'
            offset={hasBottomNav ? BOTTOM_NAV_TOAST_OFFSET : 'calc(24px + env(safe-area-inset-bottom, 0px))'}
          />
        </AppSnackbarProvider>
      </ModalOpenProvider>
    </div>
  )
}
