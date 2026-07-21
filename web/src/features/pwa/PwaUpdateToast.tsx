import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { registerSW } from 'virtual:pwa-register'

/** How often to ask the browser to check for a newer service worker. */
const SW_UPDATE_INTERVAL_MS = 60 * 60 * 1000

/**
 * Registers the PWA service worker with auto-update.
 * Periodically re-checks `sw.js`; surfaces registration failures instead of an unhandled promise.
 */
export default function PwaUpdateToast() {
  const { t } = useTranslation()

  useEffect(() => {
    let cancelled = false
    let intervalId = 0

    registerSW({
      immediate: true,
      onRegisteredSW(swUrl, registration) {
        if (!registration || cancelled) return
        const id = window.setInterval(() => {
          if (registration.installing || !navigator.onLine) return
          void (async () => {
            try {
              const resp = await fetch(swUrl, {
                cache: 'no-store',
                headers: { 'cache-control': 'no-cache' },
              })
              if (resp.ok) await registration.update()
            } catch {
              // offline / transient — ignore
            }
          })()
        }, SW_UPDATE_INTERVAL_MS)
        if (cancelled) {
          window.clearInterval(id)
          return
        }
        intervalId = id
      },
      onRegisterError() {
        if (!cancelled) toast.error(t('PwaUpdateFailed'), { duration: 8000 })
      },
    })

    return () => {
      cancelled = true
      if (intervalId) window.clearInterval(intervalId)
    }
  }, [t])

  return null
}
