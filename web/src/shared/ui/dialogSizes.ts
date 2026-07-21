import type { CSSProperties } from 'react'

/**
 * Canonical desktop dialog widths. Applied via inline `style` on Modal.Dialog /
 * AppDialog `dialogStyle` — HeroUI size ceilings live in CSS layers, so utilities
 * alone cannot widen past `lg`.
 */
export const DIALOG_SHEET_L: CSSProperties = {
  minWidth: 'min(92vw, 72rem)',
  maxWidth: 'min(92vw, 72rem)',
  width: 'min(92vw, 72rem)',
}

/** Mid-width sheet (add / search style dialogs). */
export const DIALOG_SHEET_M: CSSProperties = {
  minWidth: 'min(92vw, 48rem)',
  maxWidth: 'min(92vw, 48rem)',
  width: 'min(92vw, 48rem)',
}

/** Default player window — wide cinema frame (height comes from PlayerChrome aspect-ratio). */
export const PLAYER_DIALOG_NORMAL: CSSProperties = {
  minWidth: 'min(94vw, 64rem)',
  maxWidth: 'min(94vw, 64rem)',
  width: 'min(94vw, 64rem)',
}

/** Expanded player sheet — near-viewport cinema width. */
export const PLAYER_DIALOG_EXPANDED: CSSProperties = {
  minWidth: 'min(96vw, 80rem)',
  maxWidth: 'min(96vw, 80rem)',
  width: 'min(96vw, 80rem)',
}

/**
 * Phone / narrow tablet fullscreen surface.
 * Prefer `100dvh` so iOS PWA fills the real screen; HeroUI's `--visual-viewport-height`
 * is often short on installed PWAs and leaves a dead strip above the home indicator.
 */
export const DIALOG_FULLSCREEN: CSSProperties = {
  width: '100%',
  maxWidth: '100%',
  minWidth: '100%',
  height: '100dvh',
  maxHeight: '100dvh',
  minHeight: '100dvh',
  borderRadius: 0,
}

/** @deprecated Prefer DIALOG_FULLSCREEN — kept as alias for player call sites. */
export const PLAYER_DIALOG_MOBILE: CSSProperties = DIALOG_FULLSCREEN

/** Settings — fixed height so tab switches (Storage ↔ Network ↔ Torznab) don't resize the window. */
export const DIALOG_SETTINGS: CSSProperties = {
  ...DIALOG_SHEET_L,
  height: 'min(72dvh, 40rem)',
  maxHeight: 'min(72dvh, 40rem)',
  minHeight: 'min(72dvh, 40rem)',
}

/** Search results sheet — slightly narrower than Settings/Details so sparse tables don't float. */
export const DIALOG_SEARCH: CSSProperties = {
  minWidth: 'min(92vw, 56rem)',
  maxWidth: 'min(92vw, 56rem)',
  width: 'min(92vw, 56rem)',
}

/** Details sheet — fixed height so Overview / Files / Cache tabs don't resize the window. */
export const DIALOG_DETAILS: CSSProperties = {
  ...DIALOG_SHEET_L,
  height: 'min(80dvh, 46rem)',
  maxHeight: 'min(80dvh, 46rem)',
  minHeight: 'min(80dvh, 46rem)',
}

/** Edit torrent — fixed height so poster search results don't resize the sheet. */
export const DIALOG_EDIT: CSSProperties = {
  ...DIALOG_SHEET_L,
  height: 'min(80dvh, 40rem)',
  maxHeight: 'min(80dvh, 40rem)',
  minHeight: 'min(80dvh, 40rem)',
}

/** Full cache-map workspace stacked above Details. */
export const DIALOG_CACHE: CSSProperties = {
  minWidth: 'min(96vw, 80rem)',
  maxWidth: 'min(96vw, 80rem)',
  width: 'min(96vw, 80rem)',
  height: 'min(92dvh, 56rem)',
  maxHeight: 'min(92dvh, 56rem)',
  minHeight: 'min(92dvh, 56rem)',
}
