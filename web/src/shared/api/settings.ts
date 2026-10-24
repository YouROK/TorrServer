import axios from 'axios'

import type { BTSets } from 'shared/api/types'
import { settingsHost } from 'shared/api/hosts'

/** Wire integers. A float or NaN makes Go reject the whole `BTSets` object (`encoding/json` into `int`). */
const BT_INT_FALLBACKS: Record<string, number> = {
  ReaderReadAHead: 95,
  PreloadCache: 50,
  RetrackersMode: 1,
  TorrentDisconnectTimeout: 30,
  DownloadRateLimit: 0,
  UploadRateLimit: 0,
  ConnectionsLimit: 25,
  PeersListenPort: 0,
  SslPort: 0,
}

const roundFinite = (value: unknown, fallback: number): number => {
  const n = typeof value === 'number' ? value : Number(value)
  return Number.isFinite(n) ? Math.round(n) : fallback
}

/** Bytes on the wire, integer fields rounded, non-finite numbers replaced with server defaults. */
export const sanitizeBTSetsForWire = (sets: BTSets): BTSets => {
  const next: BTSets = { ...sets }
  next.CacheSize = roundFinite(sets.CacheSize, 64 * 1024 * 1024)
  for (const [key, fallback] of Object.entries(BT_INT_FALLBACKS)) {
    next[key] = roundFinite(sets[key], fallback)
  }
  next.ReaderReadAHead = Math.min(100, Math.max(5, next.ReaderReadAHead ?? 95))
  next.PreloadCache = Math.min(100, Math.max(0, next.PreloadCache ?? 50))
  return next
}

export const SETTINGS_QUERY_KEY = ['settings'] as const

export const getSettings = async (signal?: AbortSignal): Promise<BTSets> => {
  const { data } = await axios.post<BTSets>(settingsHost(), { action: 'get' }, { signal })
  return data
}

/** Persist BTSets (`action: set`). CacheSize is expected in bytes on the wire. */
export const setSettings = async (sets: BTSets): Promise<void> => {
  await axios.post(settingsHost(), { action: 'set', sets: sanitizeBTSetsForWire(sets) })
}

/** Restore BTSets to server defaults (`action: def`). */
export const resetSettings = async (): Promise<void> => {
  await axios.post(settingsHost(), { action: 'def' })
}
