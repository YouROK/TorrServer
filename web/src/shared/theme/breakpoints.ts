/**
 * Canonical layout scale for TorrServer web.
 *
 * Use `queryMax('mobile')` for phone chrome (stacked actions, section Select),
 * and `queryMax('dialog')` when a sheet should go fullscreen (details / player).
 * Do not mix them casually — dialog (960) includes tablets that still want
 * two-column desktop patterns in some places.
 */
export const BP = {
  narrow: 340,
  micro: 410,
  phone: 420,
  compact: 600,
  mobile: 700,
  cardDense: 770,
  tablet: 840,
  dialog: 960,
  shortTable: 970,
  list2: 1100,
  list3: 1260,
  desktop: 1280,
} as const

export type Breakpoint = keyof typeof BP

export const bp = (name: Breakpoint): number => BP[name]

/** CSS `@media` fragment for styled layers that cannot use JS `useMediaQuery`. */
export const mediaMax = (name: Breakpoint): string => `@media (max-width: ${BP[name]}px)`

/** MatchMedia query string for HeroUI `useMediaQuery` / `window.matchMedia`. */
export const queryMax = (name: Breakpoint): string => `(max-width: ${BP[name]}px)`

/** Short landscape phones / split-view: hide tall chrome that would crush content. */
export const MEDIA_SHORT_VIEWPORT = '(max-height: 500px)'

/**
 * iPad / tablet landscape: wide enough for a "desktop" sheet, but short — prefer
 * edge-to-edge dialogs. Covers iPad 11 Pro landscape (~1194×834); skips typical 1440+ laptops.
 */
export const MEDIA_TABLET_LANDSCAPE =
  '(min-width: 701px) and (max-width: 1366px) and (orientation: landscape) and (max-height: 1024px)'
