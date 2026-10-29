import type { TorrentCache, TorrentStat } from 'shared/api/types'
import { PRELOAD, WORKING } from 'shared/torrent/states'

/**
 * Cache filled/capacity for the details hero tile on every tab. `/cache` polls only on
 * the Cache and Stats tabs, and its snapshot is seeded from the session store, so on
 * other tabs it is stale. The details poll's `preloaded_bytes` is the server's cache
 * Filled, refreshed every second.
 */
export function heroCacheFill(
  cache: TorrentCache,
  torrent: Pick<TorrentStat, 'stat' | 'preloaded_bytes'>,
  cachePolling: boolean,
  cacheSizeBytes?: number,
): { filled?: number; capacity?: number } {
  const capacity = cache.Capacity ?? cacheSizeBytes
  if (cachePolling && cache.Filled != null) return { filled: cache.Filled, capacity }
  const active = torrent.stat === WORKING || torrent.stat === PRELOAD
  // preloaded_bytes is omitempty: an active torrent with an empty cache sends no field.
  return { filled: active ? (torrent.preloaded_bytes ?? 0) : undefined, capacity }
}
