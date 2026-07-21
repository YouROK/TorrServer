import { useCallback, useEffect, useState } from 'react'

import { readLocalBool, writeLocalJson } from 'shared/lib/localPrefs'

const DISMISS_PREF_KEY = 'androidPwaInstallDismissed'

/** Chromium fires this before showing the native install UI — we capture it for a custom banner. */
interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>
}

/**
 * Android / Chromium PWA install via `beforeinstallprompt`.
 * Returns nullish `canInstall` until the event fires and the user hasn't dismissed forever.
 */
export function useAndroidInstallPrompt() {
  const [deferred, setDeferred] = useState<BeforeInstallPromptEvent | null>(null)
  const [dismissed, setDismissed] = useState(() => readLocalBool(DISMISS_PREF_KEY))

  useEffect(() => {
    const onPrompt = (event: Event) => {
      event.preventDefault()
      setDeferred(event as BeforeInstallPromptEvent)
    }
    window.addEventListener('beforeinstallprompt', onPrompt)
    return () => window.removeEventListener('beforeinstallprompt', onPrompt)
  }, [])

  const canInstall = Boolean(deferred) && !dismissed

  const promptInstall = useCallback(async () => {
    if (!deferred) return
    await deferred.prompt()
    try {
      await deferred.userChoice
    } catch {
      // Ignore choice errors
    }
    setDeferred(null)
  }, [deferred])

  const dismiss = useCallback(() => {
    setDismissed(true)
    writeLocalJson(DISMISS_PREF_KEY, true)
    setDeferred(null)
  }, [])

  return { canInstall, promptInstall, dismiss }
}
