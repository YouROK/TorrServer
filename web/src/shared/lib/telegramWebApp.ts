/** Minimal Telegram Mini App bridge — reuses the existing SPA inside Telegram. */

export interface TelegramThemeParams {
  bg_color?: string
  text_color?: string
  hint_color?: string
  button_color?: string
  button_text_color?: string
  secondary_bg_color?: string
}

export interface TelegramWebApp {
  initData: string
  ready: () => void
  expand: () => void
  themeParams?: TelegramThemeParams
  setHeaderColor?: (color: string) => void
  setBackgroundColor?: (color: string) => void
  colorScheme?: 'light' | 'dark'
}

declare global {
  interface Window {
    Telegram?: { WebApp?: TelegramWebApp }
  }
}

const TELEGRAM_SDK_URL = 'https://telegram.org/js/telegram-web-app.js'
const TELEGRAM_SDK_SCRIPT_ID = 'telegram-web-app-sdk'
const TELEGRAM_SDK_TIMEOUT_MS = 5000

let loadPromise: Promise<void> | null = null

export function getTelegramWebApp(): TelegramWebApp | null {
  if (typeof window === 'undefined') return null
  return window.Telegram?.WebApp ?? null
}

export function getTelegramInitData(): string {
  return getTelegramWebApp()?.initData?.trim() || ''
}

/** True when opened via bot Mini App link (`?tg=1`) or Telegram already injected WebApp. */
export function hasTelegramMiniAppQuery(search = typeof window !== 'undefined' ? window.location.search : ''): boolean {
  return new URLSearchParams(search).get('tg') === '1'
}

/** Pure gate used by needsTelegramSdk (and unit tests). */
export function needsTelegramSdkState(search: string, webAppPresent: boolean): boolean {
  return webAppPresent || hasTelegramMiniAppQuery(search)
}

export function needsTelegramSdk(): boolean {
  if (typeof window === 'undefined') return false
  return needsTelegramSdkState(window.location.search, Boolean(getTelegramWebApp()))
}

export function isTelegramMiniApp(): boolean {
  return Boolean(getTelegramInitData()) || hasTelegramMiniAppQuery()
}

function applyTelegramTheme(wa: TelegramWebApp): void {
  try {
    wa.ready()
    wa.expand?.()
    const bg = wa.themeParams?.bg_color
    if (bg) {
      document.documentElement.style.setProperty('--background', bg)
      wa.setBackgroundColor?.(bg)
      wa.setHeaderColor?.(bg)
    }
  } catch {
    // Older clients / incomplete SDK — ignore
  }
}

/** Inject Telegram SDK once; times out so a blocked telegram.org cannot hang the SPA. */
export function loadTelegramSdk(timeoutMs = TELEGRAM_SDK_TIMEOUT_MS): Promise<void> {
  if (typeof document === 'undefined') return Promise.resolve()
  if (getTelegramWebApp()) return Promise.resolve()

  if (loadPromise) return loadPromise

  loadPromise = new Promise<void>((resolve, reject) => {
    const existing = document.getElementById(TELEGRAM_SDK_SCRIPT_ID) as HTMLScriptElement | null
    if (existing) {
      if (getTelegramWebApp()) {
        resolve()
        return
      }
      existing.addEventListener('load', () => resolve(), { once: true })
      existing.addEventListener('error', () => reject(new Error('Telegram SDK failed to load')), { once: true })
      return
    }

    const script = document.createElement('script')
    script.id = TELEGRAM_SDK_SCRIPT_ID
    script.src = TELEGRAM_SDK_URL
    script.async = true

    const timer = window.setTimeout(() => {
      script.remove()
      loadPromise = null
      reject(new Error('Telegram SDK load timeout'))
    }, timeoutMs)

    script.onload = () => {
      window.clearTimeout(timer)
      resolve()
    }
    script.onerror = () => {
      window.clearTimeout(timer)
      script.remove()
      loadPromise = null
      reject(new Error('Telegram SDK failed to load'))
    }

    document.head.appendChild(script)
  })

  return loadPromise
}

/**
 * Call once at boot. Loads SDK only for Mini App entry (`?tg=1`); never blocks React mount.
 * Returns a promise for tests / optional await — callers should `void` it at boot.
 */
export async function bootstrapTelegramWebApp(): Promise<void> {
  if (!needsTelegramSdk()) return
  try {
    await loadTelegramSdk()
  } catch {
    return
  }
  const wa = getTelegramWebApp()
  if (!wa) return
  applyTelegramTheme(wa)
}
