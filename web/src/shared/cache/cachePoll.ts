import type { TorrentCache } from 'shared/api/types'

import { cacheVisualEqual } from './cacheFingerprint'

/** Active cadence while pieces or readers change. */
export const CACHE_POLL_ACTIVE_MS = 100
/** Quiet cadence after no visual changes and no readers. */
export const CACHE_POLL_IDLE_MS = 1000
/** Switch to idle after this many ms without visual changes (and no readers). */
export const CACHE_IDLE_AFTER_MS = 2000

/** Hidden tabs do not hit `/cache`. */
export const cachePollPaused = (hidden: boolean): boolean => hidden

/**
 * Next delay after a poll. A failed or unchanged idle poll stays at 1s.
 * Unchanged + active readers (or a recent change) stays at 100ms when `fast`.
 */
export function cachePollDelay(input: {
  fast: boolean
  changed: boolean
  hasActiveReaders: boolean
  quietForMs: number
  failed: boolean
}): number {
  if (input.failed) return CACHE_POLL_IDLE_MS
  if (!input.fast) return CACHE_POLL_IDLE_MS
  if (!input.changed) {
    const quiet = input.quietForMs >= CACHE_IDLE_AFTER_MS
    return !input.hasActiveReaders && quiet ? CACHE_POLL_IDLE_MS : CACHE_POLL_ACTIVE_MS
  }
  return CACHE_POLL_ACTIVE_MS
}

/**
 * Snapshot to paint. `null` means keep the previous one: the request was
 * cancelled (slow response after unmount or hash switch) or nothing visual changed.
 */
export function takeCacheSnapshot(cancelled: boolean, prev: TorrentCache, next: TorrentCache): TorrentCache | null {
  if (cancelled || cacheVisualEqual(prev, next)) return null
  return next
}
