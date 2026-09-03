import { describe, expect, it } from 'vitest'

import type { TorrentCache, TorrentStat } from 'shared/api/types'
import {
  bufferAheadBytes,
  bufferFillPercent,
  formatBufferAheadLabel,
  formatBufferFilledLabel,
  formatCacheFilledLabel,
  formatSizeToClassicUnits,
  getPeerString,
  humanizeSize,
  resolveBufferTargetBytes,
} from './format'

const torrent = (overrides: Partial<TorrentStat> = {}): TorrentStat => ({ hash: 'abc', ...overrides })

describe('humanizeSize', () => {
  it('rounds to whole units', () => {
    expect(humanizeSize(64 * 1024 * 1024)).toMatch(/^64 /)
    expect(humanizeSize(1.5 * 1024 * 1024)).toMatch(/^2 /)
    expect(humanizeSize(512)).toMatch(/^512 /)
    expect(humanizeSize(64 * 1024 * 1024)).not.toMatch(/\./)
  })

  it('returns an em dash for missing values', () => {
    expect(humanizeSize(null)).toBe('—')
    expect(humanizeSize(undefined)).toBe('—')
  })
})

describe('formatSizeToClassicUnits', () => {
  it('rounds to whole units', () => {
    expect(formatSizeToClassicUnits(64 * 1024 * 1024)).toBe('64 MB')
    expect(formatSizeToClassicUnits(1.5 * 1024 * 1024)).toBe('2 MB')
  })
})

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

  it('shows raw filled when over capacity but caps percent at 100', () => {
    const label = formatCacheFilledLabel(274, 256)
    expect(label).toMatch(/100%/)
    expect(label).not.toMatch(/107%/)
    expect(label).toContain('/')
  })

  it('always appends percent when requested', () => {
    const label = formatCacheFilledLabel(50, 100, { percent: 'always' })
    expect(label).toMatch(/50%/)
  })
})

describe('resolveBufferTargetBytes', () => {
  it('matches server Preload size: Capacity × PreloadCache% with no 32 MiB floor', () => {
    const capacity = 64 * 1024 * 1024
    // 64 MB × 25% = 16 MB — what the server actually preloads.
    expect(resolveBufferTargetBytes(capacity, 25)).toBe(capacity * 0.25)
  })

  it('uses Capacity × PreloadCache%', () => {
    const capacity = 256 * 1024 * 1024
    expect(resolveBufferTargetBytes(capacity, 50)).toBe(capacity * 0.5)
  })

  it('never exceeds Cache Size', () => {
    const capacity = 64 * 1024 * 1024
    expect(resolveBufferTargetBytes(capacity, 100)).toBe(capacity)
  })

  it('returns null when preload percent is zero', () => {
    expect(resolveBufferTargetBytes(64 * 1024 * 1024, 0)).toBeNull()
  })
})

describe('formatBufferFilledLabel', () => {
  it('orders filled before target', () => {
    const label = formatBufferFilledLabel(10 * 1024 * 1024, 32 * 1024 * 1024, { percent: 'always' })
    expect(label).toMatch(/31%/)
    expect(label).toContain('/')
  })

  it('caps percent at 100 when filled exceeds target', () => {
    const label = formatBufferFilledLabel(267 * 1024 * 1024, 64 * 1024 * 1024, { percent: 'always' })
    expect(label).toMatch(/100%/)
    expect(label).not.toMatch(/418%/)
  })

  it('computes capped percent', () => {
    expect(bufferFillPercent(40, 32)).toBe(100)
    expect(bufferFillPercent(16, 32)).toBe(50)
  })
})

describe('formatBufferAheadLabel', () => {
  it('returns a size-ahead label without a / target fraction', () => {
    const label = formatBufferAheadLabel(242 * 1024 * 1024)
    expect(label).toBeTruthy()
    expect(label).not.toContain('/')
    expect(label).toMatch(/242/)
  })

  it('uses the empty-ahead label when there is no contiguous reserve', () => {
    const label = formatBufferAheadLabel(0)
    expect(label).toBeTruthy()
    expect(label).not.toMatch(/0 /)
  })

  it('returns null for missing input', () => {
    expect(formatBufferAheadLabel(null)).toBeNull()
    expect(formatBufferAheadLabel(undefined)).toBeNull()
  })
})

describe('bufferAheadBytes', () => {
  const cache = (pieces: TorrentCache['Pieces'], readers: TorrentCache['Readers']): TorrentCache => ({
    PiecesCount: 10,
    PiecesLength: 100,
    Pieces: pieces,
    Readers: readers,
  })

  it('sums contiguous ready pieces from the playhead', () => {
    const model = cache(
      {
        3: { Size: 100, Length: 100 },
        4: { Size: 100, Length: 100 },
        5: { Size: 100, Length: 100 },
      },
      [{ Reader: 3, Start: 3, End: 9, Active: true }],
    )
    expect(bufferAheadBytes(model)).toBe(300)
  })

  it('stops at the first hole rather than counting the whole cache', () => {
    const model = cache(
      {
        3: { Size: 100, Length: 100 },
        4: { Size: 40, Length: 100 },
        // Piece 5 is ready but unreachable — a stall happens at 4.
        5: { Size: 100, Length: 100 },
      },
      [{ Reader: 3, Start: 3, End: 9, Active: true }],
    )
    // Full head + partial next piece (counted), then stop.
    expect(bufferAheadBytes(model)).toBe(140)
  })

  it('counts a partial playhead piece instead of reporting zero', () => {
    const model = cache({ 3: { Size: 40, Length: 100 } }, [{ Reader: 3, Start: 3, End: 9, Active: true }])
    expect(bufferAheadBytes(model)).toBe(40)
  })

  it('returns zero when the playhead piece is missing', () => {
    const model = cache({ 5: { Size: 100, Length: 100 } }, [{ Reader: 3, Start: 3, End: 9, Active: true }])
    expect(bufferAheadBytes(model)).toBe(0)
  })

  it('stops at a missing piece', () => {
    const model = cache({ 3: { Size: 100, Length: 100 }, 5: { Size: 100, Length: 100 } }, [
      { Reader: 3, Start: 3, End: 9, Active: true },
    ])
    expect(bufferAheadBytes(model)).toBe(100)
  })

  it('ignores pieces behind the playhead', () => {
    const model = cache({ 0: { Size: 100, Length: 100 }, 3: { Size: 100, Length: 100 } }, [
      { Reader: 3, Start: 0, End: 9, Active: true },
    ])
    expect(bufferAheadBytes(model)).toBe(100)
  })

  it('returns null without an active reader so callers fall back to Filled', () => {
    expect(bufferAheadBytes(cache({ 3: { Size: 100, Length: 100 } }, []))).toBeNull()
    expect(
      bufferAheadBytes(cache({ 3: { Size: 100, Length: 100 } }, [{ Reader: 3, Start: 3, End: 9, Active: false }])),
    ).toBeNull()
    expect(bufferAheadBytes(undefined)).toBeNull()
  })
})
