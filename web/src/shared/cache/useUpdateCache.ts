import { useEffect, useMemo, useRef, useState } from 'react'
import axios from 'axios'
import type { TorrentCache } from 'shared/api/types'
import { cacheHost } from 'shared/api/hosts'

import { buildFocusModel, type CacheDrawModel } from './buildCacheMap'

/** Active fill cadence while pieces/readers change (near-real-time snake). */
const CACHE_POLL_ACTIVE_MS = 100
/** Idle cadence when cache snapshot is unchanged and no readers. */
const CACHE_POLL_IDLE_MS = 400
/** Switch to idle after this many ms without visual changes (and no readers). */
const CACHE_IDLE_AFTER_MS = 2000

const readersFingerprint = (readers: TorrentCache['Readers']) => {
  if (!readers?.length) return ''
  return readers.map(r => `${r.Reader ?? ''}:${r.Start ?? ''}-${r.End ?? ''}`).join('|')
}

const piecesFingerprint = (pieces: TorrentCache['Pieces']) => {
  if (!pieces) return ''
  if (Array.isArray(pieces)) {
    let acc = ''
    for (let i = 0; i < pieces.length; i++) {
      const p = pieces[i]
      if (!p) continue
      acc += `${i}:${p.Size ?? 0}:${p.Priority ?? 0}:${p.Completed ? 1 : 0};`
    }
    return acc
  }
  let acc = ''
  for (const [key, p] of Object.entries(pieces)) {
    if (!p) continue
    acc += `${key}:${p.Size ?? 0}:${p.Priority ?? 0}:${p.Completed ? 1 : 0};`
  }
  return acc
}

const cacheVisualEqual = (a: TorrentCache, b: TorrentCache) =>
  a.Filled === b.Filled &&
  a.Capacity === b.Capacity &&
  a.PiecesCount === b.PiecesCount &&
  a.PiecesLength === b.PiecesLength &&
  readersFingerprint(a.Readers) === readersFingerprint(b.Readers) &&
  piecesFingerprint(a.Pieces) === piecesFingerprint(b.Pieces)

export interface UseUpdateCacheOptions {
  /** When false, polling stops. Defaults to true when hash is set. */
  enabled?: boolean
  /** When false, never uses the 100ms active cadence (idle/slow only). Default true. */
  fast?: boolean
}

/**
 * Poll `/cache` for the snake visualization.
 * Active (~100ms) while pieces/readers change; idle (~400ms) after quiet + no readers.
 * Keeps the last good snapshot on error; pauses timers while `document.hidden`.
 */
export const useUpdateCache = (hash?: string, options?: UseUpdateCacheOptions) => {
  const enabled = options?.enabled ?? true
  const fast = options?.fast ?? true
  const [cache, setCache] = useState<TorrentCache>({})
  const componentIsMounted = useRef(true)
  const timerID = useRef<ReturnType<typeof setTimeout> | null>(null)
  const inFlight = useRef(false)
  const cacheRef = useRef<TorrentCache>({})
  const lastChangeAt = useRef(0)
  const pollMs = useRef(CACHE_POLL_ACTIVE_MS)

  useEffect(
    () => () => {
      componentIsMounted.current = false
    },
    [],
  )

  useEffect(() => {
    if (!hash || !enabled) {
      if (timerID.current) clearTimeout(timerID.current)
      return undefined
    }

    let cancelled = false
    pollMs.current = fast ? CACHE_POLL_ACTIVE_MS : CACHE_POLL_IDLE_MS

    const scheduleNext = () => {
      if (cancelled) return
      timerID.current = setTimeout(fetchCache, pollMs.current)
    }

    const fetchCache = () => {
      if (cancelled || inFlight.current) return
      if (document.hidden) return
      inFlight.current = true
      axios
        .post(cacheHost(), { action: 'get', hash })
        .then(({ data }) => {
          if (!componentIsMounted.current || cancelled) return
          const next = (data || {}) as TorrentCache
          const hasReaders = (next.Readers?.length ?? 0) > 0
          if (cacheVisualEqual(cacheRef.current, next)) {
            // Stay near-real-time while readers are active OR fill recently changed.
            // Idle 400ms only after quiet + no readers (peers can still fill without a player).
            if (fast) {
              const quiet = Date.now() - lastChangeAt.current >= CACHE_IDLE_AFTER_MS
              pollMs.current = !hasReaders && quiet ? CACHE_POLL_IDLE_MS : CACHE_POLL_ACTIVE_MS
            }
            return
          }
          lastChangeAt.current = Date.now()
          pollMs.current = fast ? CACHE_POLL_ACTIVE_MS : CACHE_POLL_IDLE_MS
          cacheRef.current = next
          setCache(next)
        })
        .catch(() => {
          // Keep last good snapshot — wiping to {} flickers the snake empty on transient errors.
          if (!componentIsMounted.current || cancelled) return
          pollMs.current = CACHE_POLL_IDLE_MS
        })
        .finally(() => {
          inFlight.current = false
          if (!document.hidden) scheduleNext()
        })
    }

    fetchCache()

    const onVisibility = () => {
      if (document.hidden) {
        if (timerID.current) clearTimeout(timerID.current)
        return
      }
      pollMs.current = fast ? CACHE_POLL_ACTIVE_MS : CACHE_POLL_IDLE_MS
      fetchCache()
    }
    document.addEventListener('visibilitychange', onVisibility)

    return () => {
      cancelled = true
      if (timerID.current) clearTimeout(timerID.current)
      document.removeEventListener('visibilitychange', onVisibility)
    }
  }, [hash, enabled, fast])

  return cache
}

export const useCreateFocusMap = (cache: TorrentCache, visibleCells: number): CacheDrawModel => {
  const lastWindowStartRef = useRef<number | undefined>(undefined)
  return useMemo(() => {
    const model = buildFocusModel(cache, visibleCells, {
      // eslint-disable-next-line react-hooks/refs -- read previous camera window for sticky focus
      lastWindowStart: lastWindowStartRef.current,
    })
    if (model.windowStart != null && model.windowEnd != null && model.windowEnd >= model.windowStart) {
      // eslint-disable-next-line react-hooks/refs -- persist camera window across cache polls
      lastWindowStartRef.current = model.windowStart
    }
    return model
  }, [cache, visibleCells])
}
