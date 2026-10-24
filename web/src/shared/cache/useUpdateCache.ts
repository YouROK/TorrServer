import { useEffect, useMemo, useRef, useState } from 'react'
import axios from 'axios'
import type { TorrentCache } from 'shared/api/types'
import { getCache } from 'shared/api/cache'

import { buildFocusModel, isReaderActive, type CacheDrawModel } from './buildCacheMap'
import {
  cachePollDelay,
  cachePollPaused,
  CACHE_POLL_ACTIVE_MS,
  CACHE_POLL_IDLE_MS,
  takeCacheSnapshot,
} from './cachePoll'
import { cheapPiecesFingerprint, cheapReadersFingerprint } from './cacheFingerprint'
import { readSnakeCache, readSnakeCamera, writeSnakeCache, writeSnakeCamera } from './snakeSession'

export interface UseUpdateCacheOptions {
  /** When false, polling stops. Defaults to true when hash is set. */
  enabled?: boolean
  /** When false, never uses the 100ms active cadence (idle/slow only). Default true. */
  fast?: boolean
}

/**
 * Poll `/cache` for the snake visualization.
 * Active (~100ms) while pieces/readers change; idle (~1s) after quiet + no readers.
 * Keeps the last good snapshot on error; pauses timers while `document.hidden`.
 */
export const useUpdateCache = (hash?: string, options?: UseUpdateCacheOptions) => {
  const enabled = options?.enabled ?? true
  const fast = options?.fast ?? true
  // Seeded from the session store so a reopened dialog paints the previous snake
  // immediately instead of an empty grid until the first poll lands.
  const [cache, setCache] = useState<TorrentCache>(() => readSnakeCache(hash) ?? {})
  const timerID = useRef<ReturnType<typeof setTimeout> | null>(null)
  const cacheRef = useRef<TorrentCache>(cache)
  const seededHash = useRef(hash)
  const lastChangeAt = useRef(0)
  const pollMs = useRef(CACHE_POLL_ACTIVE_MS)

  useEffect(() => {
    if (!hash || !enabled) {
      if (timerID.current) clearTimeout(timerID.current)
      return undefined
    }

    if (seededHash.current !== hash) {
      seededHash.current = hash
      const seeded = readSnakeCache(hash) ?? {}
      cacheRef.current = seeded
      setCache(seeded)
    }

    let cancelled = false
    let controller: AbortController | null = null
    // Per-run flag on purpose: a ref shared across runs left the chain dead when
    // the effect restarted (fast/hash change, StrictMode) while a request was in
    // flight — the new run skipped its fetch and the old run refused to reschedule.
    let inFlight = false
    pollMs.current = fast ? CACHE_POLL_ACTIVE_MS : CACHE_POLL_IDLE_MS

    const scheduleNext = () => {
      if (cancelled) return
      timerID.current = setTimeout(fetchCache, pollMs.current)
    }

    const fetchCache = () => {
      if (cancelled || inFlight) return
      if (cachePollPaused(document.hidden)) return
      inFlight = true
      controller = new AbortController()
      getCache(hash, controller.signal)
        .then(next => {
          const taken = takeCacheSnapshot(cancelled, cacheRef.current, next)
          if (!taken) {
            if (!cancelled && fast) {
              pollMs.current = cachePollDelay({
                fast: true,
                changed: false,
                hasActiveReaders: (next.Readers ?? []).some(isReaderActive),
                quietForMs: Date.now() - lastChangeAt.current,
                failed: false,
              })
            }
            return
          }
          lastChangeAt.current = Date.now()
          pollMs.current = cachePollDelay({
            fast,
            changed: true,
            hasActiveReaders: (taken.Readers ?? []).some(isReaderActive),
            quietForMs: 0,
            failed: false,
          })
          cacheRef.current = taken
          writeSnakeCache(hash, taken)
          setCache(taken)
        })
        .catch((err: unknown) => {
          if (cancelled || axios.isCancel(err)) return
          pollMs.current = cachePollDelay({
            fast,
            changed: false,
            hasActiveReaders: false,
            quietForMs: 0,
            failed: true,
          })
        })
        .finally(() => {
          inFlight = false
          if (!cachePollPaused(document.hidden)) scheduleNext()
        })
    }

    fetchCache()

    const onVisibility = () => {
      if (cachePollPaused(document.hidden)) {
        if (timerID.current) clearTimeout(timerID.current)
        return
      }
      pollMs.current = fast ? CACHE_POLL_ACTIVE_MS : CACHE_POLL_IDLE_MS
      fetchCache()
    }
    document.addEventListener('visibilitychange', onVisibility)

    return () => {
      cancelled = true
      controller?.abort()
      if (timerID.current) clearTimeout(timerID.current)
      document.removeEventListener('visibilitychange', onVisibility)
    }
  }, [hash, enabled, fast])

  return cache
}

interface CameraState {
  /** Snake this window belongs to — see {@link snakeCameraKey}. */
  key?: string
  budget: number
  start?: number
}

const restoreCamera = (key: string | undefined, fallbackBudget: number): CameraState => {
  const saved = readSnakeCamera(key)
  return saved ? { key, budget: saved.budget, start: saved.start } : { key, budget: fallbackBudget }
}

/**
 * Sticky 1:1 focus window. Camera is React state so budget changes clear sticky
 * start immediately and dead-zone re-centers without ref-during-render.
 * `cameraKey` additionally restores the window across remounts (dialog reopen,
 * tab switch) so the snake comes back where it was instead of re-anchoring.
 */
export const useCreateFocusMap = (cache: TorrentCache, visibleCells: number, cameraKey?: string): CacheDrawModel => {
  const [camera, setCamera] = useState<CameraState>(() => restoreCamera(cameraKey, visibleCells))
  // A key change means a different snake — fall back to its saved window, not this one's.
  // Budget may change when the pane resizes; keep lastStart so the head keeps walking
  // instead of re-anchoring to the reader range on every layout tick.
  const active = camera.key === cameraKey ? camera : restoreCamera(cameraKey, visibleCells)
  const lastStart = active.start

  const model = useMemo(
    () =>
      buildFocusModel(cache, visibleCells, {
        lastWindowStart: lastStart,
      }),
    [cache, visibleCells, lastStart],
  )

  useEffect(() => {
    if (!cameraKey) return
    if (model.windowStart == null || model.windowEnd == null || model.windowEnd < model.windowStart) return
    writeSnakeCamera(cameraKey, { budget: visibleCells, start: model.windowStart })
    // Sticky camera across poll ticks; skip update when unchanged to avoid loops.
    // eslint-disable-next-line react-hooks/set-state-in-effect -- persist focus window between cache polls
    setCamera(prev => {
      if (prev.key === cameraKey && prev.budget === visibleCells && prev.start === model.windowStart) return prev
      return { key: cameraKey, budget: visibleCells, start: model.windowStart }
    })
  }, [cameraKey, model.windowStart, model.windowEnd, visibleCells])

  return model
}

// Re-export fingerprints for memo consumers that previously inlined them.
export { cheapPiecesFingerprint, cheapReadersFingerprint }
