import { describe, expect, it } from 'vitest'

import type { TorrentStat } from 'shared/api/types'
import { formatCacheFilledLabel, getPeerString } from './format'

const torrent = (overrides: Partial<TorrentStat> = {}): TorrentStat => ({ hash: 'abc', ...overrides })

describe('getPeerString', () => {
  it('returns null when torrent is missing', () => {
    expect(getPeerString(null)).toBeNull()
    expect(getPeerString(undefined)).toBeNull()
  })

  it('returns em dash when active peers are unknown', () => {
    expect(getPeerString(torrent())).toBe('—')
  })

  it('formats active/total peers and seeders', () => {
    expect(getPeerString(torrent({ active_peers: 5, total_peers: 12, connected_seeders: 3 }))).toBe('5/12 · 3')
  })

  it('defaults missing totals and seeders to zero', () => {
    expect(getPeerString(torrent({ active_peers: 2 }))).toBe('2/0 · 0')
  })
})

describe('formatCacheFilledLabel', () => {
  it('returns null for incomplete input', () => {
    expect(formatCacheFilledLabel(null, 100)).toBeNull()
    expect(formatCacheFilledLabel(10, 0)).toBeNull()
  })

  it('omits percent until over capacity by default', () => {
    const label = formatCacheFilledLabel(50, 100)
    expect(label).toContain('/')
    expect(label).not.toMatch(/%/)
  })

  it('appends percent when over capacity', () => {
    const label = formatCacheFilledLabel(274, 256)
    expect(label).toMatch(/107%/)
  })

  it('always appends percent when requested', () => {
    const label = formatCacheFilledLabel(50, 100, { percent: 'always' })
    expect(label).toMatch(/50%/)
  })
})
