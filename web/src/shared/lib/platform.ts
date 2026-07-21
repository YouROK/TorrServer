/** True when running as an installed PWA / iOS "Add to Home Screen" (display-mode standalone). */
export const detectStandaloneApp = (): boolean => {
  if (typeof window === 'undefined') return false

  const matchMedia = window.matchMedia?.bind(window)
  const isDisplayModeStandalone = (mode: string) => {
    try {
      return !!matchMedia && matchMedia(mode).matches
    } catch {
      return false
    }
  }

  const byDisplayMode =
    isDisplayModeStandalone('(display-mode: standalone)') ||
    isDisplayModeStandalone('screen and (display-mode: standalone)')

  return byDisplayMode || (window.navigator as Navigator & { standalone?: boolean }).standalone === true
}

/**
 * Eager snapshot at module load — use for install/protocol guides only.
 * Do not gate layout or features on this; prefer {@link detectStandaloneApp} when
 * the value must re-evaluate after install.
 */
export const isStandaloneApp = detectStandaloneApp()

/**
 * Apple platform detection. iPadOS 13+ often reports Macintosh in UA —
 * `maxTouchPoints > 1` on a Mac UA means iPad, not desktop Mac.
 */
export const detectApplePlatform = (): { isMac: boolean; isIOS: boolean } => {
  if (typeof window === 'undefined' || typeof navigator === 'undefined') {
    return { isMac: false, isIOS: false }
  }

  const userAgent = navigator.userAgent || ''
  const uaDataPlatform =
    (navigator as Navigator & { userAgentData?: { platform?: string } }).userAgentData?.platform || ''
  /** Deprecated but still populated by every Chromium/WebKit/Gecko build ("MacIntel", "iPhone", …) — a reliable
   *  fallback for browsers that report a reduced/generic `userAgent` string. */
  const navigatorPlatform = navigator.platform || ''
  const isMacUA = /Macintosh|Mac OS X/i.test(userAgent)
  const isMacPlatform = uaDataPlatform.toLowerCase().includes('mac') || /Mac/i.test(navigatorPlatform)

  const isMac = isMacUA || isMacPlatform
  const isIOS =
    /iPad|iPhone|iPod/i.test(userAgent) ||
    /iPad|iPhone|iPod/i.test(navigatorPlatform) ||
    (isMacUA && navigator.maxTouchPoints > 1)

  return { isMac: Boolean(isMac), isIOS: Boolean(isIOS) }
}

export const isAppleDevice = (): boolean => {
  const { isMac, isIOS } = detectApplePlatform()
  return isMac || isIOS
}

/** Desktop Mac only (excludes iPad masquerading as Macintosh). */
export const isMacOS = (): boolean => {
  const { isMac, isIOS } = detectApplePlatform()
  return isMac && !isIOS
}

/** Coarse desktop vs phone/tablet (treats touch-Mac as tablet). */
export const isDesktop = (): boolean => {
  if (typeof window === 'undefined' || typeof navigator === 'undefined') {
    return false
  }

  const userAgent = navigator.userAgent || ''
  const isMobile = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(userAgent)
  const isTabletWithTouch = /Macintosh/i.test(userAgent) && navigator.maxTouchPoints > 1

  return !isMobile && !isTabletWithTouch
}
