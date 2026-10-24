import { describe, expect, it } from 'vitest'

import {
  CACHE_IDLE_AFTER_MS,
  CACHE_POLL_ACTIVE_MS,
  CACHE_POLL_IDLE_MS,
  cachePollDelay,
  cachePollPaused,
  takeCacheSnapshot,
} from './cachePoll'

describe('cache poll', () => {
  it('pauses while the document is hidden', () => {
    expect(cachePollPaused(true)).toBe(true)
    expect(cachePollPaused(false)).toBe(false)
  })

  it('stays fast while readers move and idles after a quiet unchanged poll', () => {
    expect(cachePollDelay({ fast: true, changed: true, hasActiveReaders: true, quietForMs: 0, failed: false })).toBe(
      CACHE_POLL_ACTIVE_MS,
    )
    expect(
      cachePollDelay({
        fast: true,
        changed: false,
        hasActiveReaders: false,
        quietForMs: CACHE_IDLE_AFTER_MS,
        failed: false,
      }),
    ).toBe(CACHE_POLL_IDLE_MS)
    expect(
      cachePollDelay({
        fast: true,
        changed: false,
        hasActiveReaders: true,
        quietForMs: CACHE_IDLE_AFTER_MS,
        failed: false,
      }),
    ).toBe(CACHE_POLL_ACTIVE_MS)
    expect(cachePollDelay({ fast: false, changed: true, hasActiveReaders: true, quietForMs: 0, failed: false })).toBe(
      CACHE_POLL_IDLE_MS,
    )
    expect(cachePollDelay({ fast: true, changed: true, hasActiveReaders: true, quietForMs: 0, failed: true })).toBe(
      CACHE_POLL_IDLE_MS,
    )
  })

  it('drops a cancelled or unchanged snapshot so a slow response cannot overwrite newer state', () => {
    const prev = { Filled: 1, PiecesCount: 10 }
    const next = { Filled: 2, PiecesCount: 10 }
    expect(takeCacheSnapshot(true, prev, next)).toBeNull()
    expect(takeCacheSnapshot(false, prev, prev)).toBeNull()
    expect(takeCacheSnapshot(false, prev, next)).toBe(next)
  })
})
