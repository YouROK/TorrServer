import { beforeEach, describe, expect, it } from 'vitest'
import type { TorrentCache } from 'shared/api/types'

import {
  clearSnakeSessions,
  readSnakeCache,
  readSnakeCamera,
  snakeCameraKey,
  writeSnakeCache,
  writeSnakeCamera,
} from './snakeSession'

const cacheOf = (filled: number): TorrentCache => ({ Filled: filled, PiecesCount: 100 })

describe('snakeSession', () => {
  beforeEach(() => clearSnakeSessions())

  it('returns the snapshot a reopened dialog needs to paint immediately', () => {
    writeSnakeCache('abc', cacheOf(42))
    expect(readSnakeCache('abc')?.Filled).toBe(42)
    expect(readSnakeCache('other')).toBeUndefined()
  })

  it('ignores a missing hash on both read and write', () => {
    writeSnakeCache(undefined, cacheOf(1))
    expect(readSnakeCache(undefined)).toBeUndefined()
    writeSnakeCamera(undefined, { budget: 10, start: 1 })
    expect(readSnakeCamera(undefined)).toBeUndefined()
  })

  it('drops entries older than the TTL', () => {
    const now = 1_000_000
    writeSnakeCache('abc', cacheOf(7), now)
    writeSnakeCamera('abc:detailed', { budget: 200, start: 5 }, now)

    expect(readSnakeCache('abc', now + 4 * 60_000)?.Filled).toBe(7)
    expect(readSnakeCache('abc', now + 6 * 60_000)).toBeUndefined()
    expect(readSnakeCamera('abc:detailed', now + 6 * 60_000)).toBeUndefined()
  })

  it('evicts the least recently written torrent past the limit', () => {
    for (let i = 0; i < 9; i++) writeSnakeCache(`h${i}`, cacheOf(i))
    expect(readSnakeCache('h0')).toBeUndefined()
    expect(readSnakeCache('h8')?.Filled).toBe(8)
  })

  it('keeps a torrent alive when it is written again', () => {
    for (let i = 0; i < 8; i++) writeSnakeCache(`h${i}`, cacheOf(i))
    writeSnakeCache('h0', cacheOf(100))
    writeSnakeCache('h8', cacheOf(8))
    expect(readSnakeCache('h0')?.Filled).toBe(100)
    expect(readSnakeCache('h1')).toBeUndefined()
  })

  it('separates camera windows per snake mode', () => {
    writeSnakeCamera(snakeCameraKey('abc', 'detailed'), { budget: 400, start: 120 })
    writeSnakeCamera(snakeCameraKey('abc', 'mini'), { budget: 60, start: 130 })
    expect(readSnakeCamera(snakeCameraKey('abc', 'detailed'))).toEqual({ budget: 400, start: 120 })
    expect(readSnakeCamera(snakeCameraKey('abc', 'mini'))).toEqual({ budget: 60, start: 130 })
    expect(snakeCameraKey(undefined, 'detailed')).toBeUndefined()
  })
})
