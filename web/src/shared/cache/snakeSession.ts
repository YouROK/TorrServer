import type { TorrentCache } from 'shared/api/types'

/**
 * Snake state that must survive a closed dialog. Details / cache-map dialogs are
 * unmounted on close (and tab panels are unmounted by react-aria), so without
 * this store every reopen paints an empty placeholder grid for one poll and then
 * re-anchors the camera somewhere else.
 */

/** Saved focus window. `budget` is the cell count it was measured for. */
export interface SnakeCamera {
  budget: number
  start: number
}

/** Long enough to cover browsing away and back, short enough to not resurrect stale maps. */
const SESSION_TTL_MS = 5 * 60_000
/** Keep a handful of recent torrents; entries are small but unbounded growth is not free. */
const SESSION_LIMIT = 8

interface Entry<T> {
  value: T
  at: number
}

const cacheStore = new Map<string, Entry<TorrentCache>>()
const cameraStore = new Map<string, Entry<SnakeCamera>>()

const readEntry = <T>(store: Map<string, Entry<T>>, key: string | undefined, now: number): T | undefined => {
  if (!key) return undefined
  const entry = store.get(key)
  if (!entry) return undefined
  if (now - entry.at > SESSION_TTL_MS) {
    store.delete(key)
    return undefined
  }
  return entry.value
}

const writeEntry = <T>(store: Map<string, Entry<T>>, key: string | undefined, value: T, now: number) => {
  if (!key) return
  // Re-insert so Map iteration order stays least-recently-written first.
  store.delete(key)
  store.set(key, { value, at: now })
  for (const [entryKey, entry] of store) {
    if (now - entry.at > SESSION_TTL_MS) store.delete(entryKey)
  }
  while (store.size > SESSION_LIMIT) {
    const oldest = store.keys().next()
    if (oldest.done) break
    store.delete(oldest.value)
  }
}

/** Camera key — mini and detailed snakes of one torrent have different budgets. */
export const snakeCameraKey = (hash: string | undefined, mode: string): string | undefined =>
  hash ? `${hash}:${mode}` : undefined

export const readSnakeCache = (hash: string | undefined, now = Date.now()): TorrentCache | undefined =>
  readEntry(cacheStore, hash, now)

export const writeSnakeCache = (hash: string | undefined, cache: TorrentCache, now = Date.now()): void => {
  writeEntry(cacheStore, hash, cache, now)
}

export const readSnakeCamera = (key: string | undefined, now = Date.now()): SnakeCamera | undefined =>
  readEntry(cameraStore, key, now)

export const writeSnakeCamera = (key: string | undefined, camera: SnakeCamera, now = Date.now()): void => {
  writeEntry(cameraStore, key, camera, now)
}

/** Test helper — module state would otherwise leak between cases. */
export const clearSnakeSessions = (): void => {
  cacheStore.clear()
  cameraStore.clear()
}
