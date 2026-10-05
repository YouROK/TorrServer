import { detectStandaloneApp } from './platform'

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
 * Standalone: `100dvh` / `clientHeight` can be short (iOS excluded the home-indicator
 * or status-bar band), so take the largest of the window metrics. Never use `100vh`
 * or `screen.height`: with status-bar-style `default` (iOS 27) the system reserves the
 * status-bar band above the web view, and those report the full display — the bottom
 * nav would sit ~60px off-screen. Desktop PWAs are windowed, same reason.
 */
function readViewportHeightPx(): number {
  if (typeof window === 'undefined') return 0

  const inner = window.innerHeight || 0
  const client = document.documentElement?.clientHeight || 0
  const visual = Math.round(window.visualViewport?.height || 0)

  if (!detectStandaloneApp()) {
    return preferVisibleViewportHeight(inner, client, visual) || Math.round(inner)
  }

  return Math.round(Math.max(inner, client, visual) || inner)
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
