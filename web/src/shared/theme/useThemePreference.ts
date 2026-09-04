import { useCallback, useSyncExternalStore } from 'react'

export const THEME_MODES = {
  LIGHT: 'light',
  DARK: 'dark',
  AUTO: 'system',
} as const

export type ThemePreference = (typeof THEME_MODES)[keyof typeof THEME_MODES]

/** Named accent/chrome palettes layered on top of light/dark. */
export const THEME_PALETTES = {
  FOREST: 'forest',
  OCEAN: 'ocean',
  SLATE: 'slate',
  EMBER: 'ember',
  ARCTIC: 'arctic',
  ROSE: 'rose',
  MEADOW: 'meadow',
  FOG: 'fog',
  COPPER: 'copper',
  CITRUS: 'citrus',
  MIDNIGHT: 'midnight',
  SAGE: 'sage',
  CORAL: 'coral',
  HONEY: 'honey',
  MOSS: 'moss',
  SKY: 'sky',
  WINE: 'wine',
  MINT: 'mint',
  CHARCOAL: 'charcoal',
  LAVA: 'lava',
  ICE: 'ice',
  OLIVE: 'olive',
  STEEL: 'steel',
  BLOSSOM: 'blossom',
  SAND: 'sand',
  PINE: 'pine',
  AZURE: 'azure',
  GRAPHITE: 'graphite',
  NOIR: 'noir',
  JADE: 'jade',
  NAVY: 'navy',
  RUST: 'rust',
  ESPRESSO: 'espresso',
  SEAFOAM: 'seafoam',
  BRICK: 'brick',
  STORM: 'storm',
  SAFFRON: 'saffron',
  DENIM: 'denim',
  CEDAR: 'cedar',
  VOLT: 'volt',
  PEARL: 'pearl',
  INK: 'ink',
  FERN: 'fern',
  BRONZE: 'bronze',
} as const

export type ThemePalette = (typeof THEME_PALETTES)[keyof typeof THEME_PALETTES]

export const THEME_PALETTE_IDS = Object.values(THEME_PALETTES) as ThemePalette[]

const STORAGE_KEY = 'ts-color-scheme'
const STORAGE_KEY_PALETTE = 'ts-color-palette'

function readStoredPreference(): ThemePreference {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored === 'light' || stored === 'dark' || stored === 'system') return stored
  } catch {
    /* ignore */
  }
  return THEME_MODES.AUTO
}

function readStoredPalette(): ThemePalette {
  try {
    const stored = localStorage.getItem(STORAGE_KEY_PALETTE)
    if (stored && (THEME_PALETTE_IDS as string[]).includes(stored)) return stored as ThemePalette
  } catch {
    /* ignore */
  }
  return THEME_PALETTES.FOREST
}

function resolveDark(preference: ThemePreference, systemDark: boolean): boolean {
  if (preference === 'dark') return true
  if (preference === 'light') return false
  return systemDark
}

function applyThemeToDocument(isDark: boolean, palette: ThemePalette) {
  const root = document.documentElement
  root.classList.toggle('dark', isDark)
  root.dataset.palette = palette
  // Force paint into the iOS home-indicator strip (var() backgrounds can fail there).
  // Use --surface so any gap under BottomNav matches the tab bar.
  const surface = getComputedStyle(root).getPropertyValue('--surface').trim()
  if (surface) {
    root.style.backgroundColor = surface
    if (document.body) document.body.style.backgroundColor = surface
  }
}

function readSystemDark(): boolean {
  if (typeof window === 'undefined') return false
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

interface ThemeStoreSnapshot {
  preference: ThemePreference
  palette: ThemePalette
  systemDark: boolean
  isDark: boolean
}

const listeners = new Set<() => void>()

let snapshot: ThemeStoreSnapshot = (() => {
  const preference = typeof window === 'undefined' ? THEME_MODES.AUTO : readStoredPreference()
  const palette = typeof window === 'undefined' ? THEME_PALETTES.FOREST : readStoredPalette()
  const systemDark = readSystemDark()
  return {
    preference,
    palette,
    systemDark,
    isDark: resolveDark(preference, systemDark),
  }
})()

function emit() {
  for (const listener of listeners) listener()
}

function persistAndApply(next: ThemeStoreSnapshot) {
  snapshot = next
  if (typeof document !== 'undefined') {
    applyThemeToDocument(next.isDark, next.palette)
  }
  try {
    localStorage.setItem(STORAGE_KEY, next.preference)
    localStorage.setItem(STORAGE_KEY_PALETTE, next.palette)
  } catch {
    /* ignore */
  }
  emit()
}

function setPreference(mode: ThemePreference) {
  persistAndApply({
    ...snapshot,
    preference: mode,
    isDark: resolveDark(mode, snapshot.systemDark),
  })
}

function setPalette(palette: ThemePalette) {
  persistAndApply({
    ...snapshot,
    palette,
  })
}

function setSystemDark(systemDark: boolean) {
  if (snapshot.systemDark === systemDark) return
  persistAndApply({
    ...snapshot,
    systemDark,
    isDark: resolveDark(snapshot.preference, systemDark),
  })
}

function subscribe(listener: () => void) {
  listeners.add(listener)
  return () => listeners.delete(listener)
}

function getSnapshot(): ThemeStoreSnapshot {
  return snapshot
}

function getServerSnapshot(): ThemeStoreSnapshot {
  return {
    preference: THEME_MODES.AUTO,
    palette: THEME_PALETTES.FOREST,
    systemDark: false,
    isDark: false,
  }
}

let mediaQueryBound = false

function ensureSystemDarkListener() {
  if (mediaQueryBound || typeof window === 'undefined') return
  mediaQueryBound = true
  const mq = window.matchMedia('(prefers-color-scheme: dark)')
  mq.addEventListener('change', event => setSystemDark(event.matches))
}

// Apply stored theme as soon as this module loads in the browser.
if (typeof document !== 'undefined') {
  applyThemeToDocument(snapshot.isDark, snapshot.palette)
  ensureSystemDarkListener()
}

export interface ThemePreferenceState {
  isDark: boolean
  preference: ThemePreference
  setPreference: (mode: ThemePreference) => void
  palette: ThemePalette
  setPalette: (palette: ThemePalette) => void
}

/**
 * Shared brightness + palette preference (localStorage). Safe to call from many components —
 * one module store keeps document/`dark`/`data-palette` in sync.
 */
export function useThemePreference(): ThemePreferenceState {
  ensureSystemDarkListener()
  const state = useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot)

  const setPreferenceCb = useCallback((mode: ThemePreference) => {
    setPreference(mode)
  }, [])

  const setPaletteCb = useCallback((palette: ThemePalette) => {
    setPalette(palette)
  }, [])

  return {
    isDark: state.isDark,
    preference: state.preference,
    setPreference: setPreferenceCb,
    palette: state.palette,
    setPalette: setPaletteCb,
  }
}
