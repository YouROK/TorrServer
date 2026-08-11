import { detectApplePlatform, detectStandaloneApp } from './platform'

const APP_HEIGHT_VAR = '--app-height'

/**
 * Pick a height that matches the *visible* browser viewport (between Safari chrome).
 *
 * On iOS Safari (non-standalone), `documentElement.clientHeight` is often the
 * large layout viewport (≈ `100vh`) while `visualViewport.height` / `innerHeight`
 * match what is actually on screen. Preferring `max(inner, client)` makes
 * `--app-height` too tall so fullscreen modals extend under the toolbar.
 */
export function preferVisibleViewportHeight(inner: number, client: number, visual: number): number {
  if (visual > 0) return Math.round(visual)
  if (inner > 0) return Math.round(inner)
  return Math.round(client) || 0
}

/**
 * Measure the physical viewport height for layout.
 *
 * On iOS installed PWAs (iOS 17–26), `100dvh` and often `window.innerHeight`
 * resolve to the viewport *excluding* the home-indicator band (~34px). That
 * leaves a dead strip under an in-flow bottom nav — the exact bug we hit.
 *
 * Master avoided this with `height: 100vh` in `@media (display-mode: standalone)`.
 * Modern guidance (piclaw PWA.md, meshcore-webui, SO 79902310): prefer CSS `100vh`
 * (probe) / `screen` metrics over `100dvh` in standalone — **iOS only**.
 *
 * Desktop standalone PWAs (macOS/Windows) are windowed: `screen.height` is the
 * display, not the app window. Using it makes `--app-height` taller than the
 * window and clips the sidebar footer / bottom chrome.
 */
function readViewportHeightPx(): number {
  if (typeof window === 'undefined') return 0

  const inner = window.innerHeight || 0
  const client = document.documentElement?.clientHeight || 0
  const visual = Math.round(window.visualViewport?.height || 0)

  if (!detectStandaloneApp()) {
    return preferVisibleViewportHeight(inner, client, visual) || Math.round(inner)
  }

  // Windowed desktop PWA — track the app window, never the display.
  if (!detectApplePlatform().isIOS) {
    return Math.round(Math.max(inner, client, visual) || inner)
  }

  let vhProbe = 0
  try {
    const probe = document.createElement('div')
    probe.setAttribute('aria-hidden', 'true')
    probe.style.cssText =
      'position:fixed;left:0;top:0;width:0;height:100vh;visibility:hidden;pointer-events:none;z-index:-1'
    document.documentElement.appendChild(probe)
    vhProbe = probe.getBoundingClientRect().height
    probe.remove()
  } catch {
    /* ignore */
  }

  // screen.height is CSS px on iOS and equals the full display in portrait standalone.
  const screenH = window.screen?.height ?? 0
  const screenW = window.screen?.width ?? 0
  // In landscape, the "short" side is the visual height.
  const screenMin = Math.min(screenH, screenW) || 0
  const screenMax = Math.max(screenH, screenW) || 0
  const isLandscape = window.matchMedia?.('(orientation: landscape)')?.matches
  const screenForOrientation = isLandscape ? screenMin : screenMax

  return Math.round(Math.max(inner, client, vhProbe, screenForOrientation) || inner)
}

function applyAppHeight() {
  if (typeof document === 'undefined') return
  const px = readViewportHeightPx()
  if (px > 0) {
    document.documentElement.style.setProperty(APP_HEIGHT_VAR, `${px}px`)
  }
}

/**
 * Pin `--app-height` before first paint when possible, and keep it in sync.
 * Call {@link installAppHeight} from the entry module (not only useEffect).
 */
export function installAppHeight(): () => void {
  if (typeof window === 'undefined') return () => undefined

  applyAppHeight()

  const onResize = () => applyAppHeight()
  window.addEventListener('resize', onResize)
  window.addEventListener('orientationchange', onResize)
  window.visualViewport?.addEventListener('resize', onResize)
  window.visualViewport?.addEventListener('scroll', onResize)

  // Cold-start: iOS often reports a short first value — remeasure after paint.
  requestAnimationFrame(() => {
    applyAppHeight()
    requestAnimationFrame(applyAppHeight)
  })
  window.setTimeout(applyAppHeight, 100)
  window.setTimeout(applyAppHeight, 500)

  return () => {
    window.removeEventListener('resize', onResize)
    window.removeEventListener('orientationchange', onResize)
    window.visualViewport?.removeEventListener('resize', onResize)
    window.visualViewport?.removeEventListener('scroll', onResize)
  }
}

/** @deprecated Prefer {@link installAppHeight} */
export const startAppHeightSync = installAppHeight
