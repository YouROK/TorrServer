import { describe, expect, it } from 'vitest'
import type { TorrentCache } from 'shared/api/types'

import {
  buildCacheDrawModel,
  buildFocusModel,
  clampReaderRangeInclusive,
  forEachPieceInReaderRange,
  pieceFillPercentage,
  priorityDebugLabel,
  resolveFocusWindow,
  resolveFocusWindowSize,
} from './buildCacheMap'

describe('clampReaderRangeInclusive', () => {
  it('treats End as inclusive and clamps to piecesCount', () => {
    expect(clampReaderRangeInclusive(2, 5, 10)).toEqual({ start: 2, end: 5 })
    expect(clampReaderRangeInclusive(0, 0, 10)).toEqual({ start: 0, end: 0 })
    expect(clampReaderRangeInclusive(-3, 99, 10)).toEqual({ start: 0, end: 9 })
  })

  it('returns null for empty or inverted ranges', () => {
    expect(clampReaderRangeInclusive(5, 2, 10)).toBeNull()
    expect(clampReaderRangeInclusive(null, 2, 10)).toBeNull()
    expect(clampReaderRangeInclusive(1, 2, 0)).toBeNull()
  })
})

describe('forEachPieceInReaderRange', () => {
  it('visits every id including End', () => {
    const ids: number[] = []
    forEachPieceInReaderRange(3, 5, 20, id => ids.push(id))
    expect(ids).toEqual([3, 4, 5])
  })
})

describe('buildCacheDrawModel', () => {
  it('returns empty model for zero pieces', () => {
    const model = buildCacheDrawModel({ PiecesCount: 0 }, 100)
    expect(model.cells).toEqual([])
    expect(model.bucketSize).toBe(1)
  })

  it('keeps 1:1 when pieces fit budget', () => {
    const cache: TorrentCache = {
      PiecesCount: 4,
      PiecesLength: 100,
      Pieces: {
        0: { Size: 100, Length: 100, Priority: 2 },
        2: { Size: 50, Length: 100, Priority: 5 },
      },
      Readers: [{ Reader: 2, Start: 1, End: 3 }],
    }
    const model = buildCacheDrawModel(cache, 100)
    expect(model.bucketSize).toBe(1)
    expect(model.cells).toHaveLength(4)
    expect(model.cells[0].percentage).toBe(100)
    expect(model.cells[0].completed).toBe(true)
    expect(model.cells[0].priority).toBe(2)
    expect(model.cells[1].percentage).toBe(0)
    expect(model.cells[2].percentage).toBe(50)
    expect(model.cells[2].isReader).toBe(true)
    expect(model.cells[1].isReaderRange).toBe(true)
    expect(model.cells[3].isReaderRange).toBe(true)
    expect(model.cells[0].pieceStart).toBe(0)
    expect(model.cells[0].pieceEnd).toBe(0)
  })

  it('merges adjacent pieces with byte-accurate fill', () => {
    const cache: TorrentCache = {
      PiecesCount: 6,
      PiecesLength: 100,
      Pieces: {
        0: { Size: 100, Length: 100 },
        1: { Size: 0, Length: 100 },
        2: { Size: 50, Length: 100 },
      },
    }
    // budget 3 → bucketSize 2 → 3 cells covering [0-1],[2-3],[4-5]
    const model = buildCacheDrawModel(cache, 3)
    expect(model.bucketSize).toBe(2)
    expect(model.cells).toHaveLength(3)
    // bucket0: 100/200 = 50%
    expect(model.cells[0].percentage).toBe(50)
    expect(model.cells[0].pieceStart).toBe(0)
    expect(model.cells[0].pieceEnd).toBe(1)
    // bucket1: 50/200 = 25%
    expect(model.cells[1].percentage).toBe(25)
    expect(model.cells[1].pieceStart).toBe(2)
    expect(model.cells[1].pieceEnd).toBe(3)
  })

  it('marks inclusive End piece as range', () => {
    const cache: TorrentCache = {
      PiecesCount: 5,
      PiecesLength: 10,
      Readers: [{ Reader: 0, Start: 2, End: 2 }],
    }
    const model = buildCacheDrawModel(cache, 100)
    expect(model.cells[2].isReaderRange).toBe(true)
    expect(model.cells[1].isReaderRange).toBe(false)
    expect(model.cells[3].isReaderRange).toBe(false)
  })
})

describe('pieceFillPercentage', () => {
  it('ignores Completed when Size is empty (avoids false-green vs Filled)', () => {
    expect(pieceFillPercentage({ Completed: true, Size: 0, Length: 100 }, 100)).toBe(0)
  })

  it('returns 0 when Length is missing/zero (cannot compute partial)', () => {
    expect(pieceFillPercentage({ Size: 50 }, 0)).toBe(0)
    expect(pieceFillPercentage({ Size: 50, Length: 0 }, 0)).toBe(0)
  })

  it('computes partial fill from Size/Length', () => {
    expect(pieceFillPercentage({ Size: 25, Length: 100 }, 100)).toBe(25)
    expect(pieceFillPercentage({ Size: 200, Length: 100 }, 100)).toBe(100)
    expect(pieceFillPercentage({ Completed: true, Size: 100, Length: 100 }, 100)).toBe(100)
  })
})

describe('resolveFocusWindow / buildFocusModel', () => {
  it('never exceeds visibleCells and keeps the head in view', () => {
    const cache: TorrentCache = {
      PiecesCount: 500,
      PiecesLength: 1024,
      // Huge capacity would previously inflate the window past the drawable grid.
      Capacity: 256 * 1024 * 1024,
      Readers: [{ Reader: 100, Start: 90, End: 120 }],
    }
    const visible = 24
    const window = resolveFocusWindow(cache, visible)
    expect(window).not.toBeNull()
    expect(window!.end - window!.start + 1).toBeLessThanOrEqual(visible)
    expect(window!.start).toBeLessThanOrEqual(100)
    expect(window!.end).toBeGreaterThanOrEqual(100)

    const model = buildFocusModel(cache, visible)
    expect(model.cells).toHaveLength(window!.end - window!.start + 1)
    expect(model.windowStart).toBe(window!.start)
    expect(model.windowEnd).toBe(window!.end)
  })

  it('fills the full visibleCells budget with small piece cells', () => {
    const cache: TorrentCache = {
      PiecesCount: 5000,
      PiecesLength: 2 * 1024 * 1024,
      Capacity: 256 * 1024 * 1024,
      // 128-piece cache window against a ~400-cell grid — still draw the whole grid.
      Readers: [{ Reader: 1005, Start: 1000, End: 1127 }],
    }
    const visible = 396
    const window = resolveFocusWindow(cache, visible)!
    expect(window.end - window.start + 1).toBe(visible)
    // The whole reader range stays covered inside the larger window.
    expect(window.start).toBeLessThanOrEqual(1000)
    expect(window.end).toBeGreaterThanOrEqual(1127)
    expect(buildFocusModel(cache, visible).cells).toHaveLength(visible)
  })

  it('keeps the whole reader range visible while the head advances', () => {
    const visible = 396
    const cache = (reader: number, start: number, end: number): TorrentCache => ({
      PiecesCount: 5000,
      PiecesLength: 2 * 1024 * 1024,
      Capacity: 256 * 1024 * 1024,
      Readers: [{ Reader: reader, Start: start, End: end }],
    })

    let lastStart = resolveFocusWindow(cache(1005, 1000, 1127), visible)!.start
    const seen = new Set<number>([lastStart])
    // Walk the head forward the way a stream does; the range moves with it.
    for (let step = 1; step <= 60; step++) {
      const window = resolveFocusWindow(cache(1005 + step, 1000 + step, 1127 + step), visible, {
        lastWindowStart: lastStart,
      })!
      // The buffered tail must never fall off the grid.
      expect(window.start).toBeLessThanOrEqual(1000 + step)
      expect(window.end).toBeGreaterThanOrEqual(1127 + step)
      lastStart = window.start
      seen.add(lastStart)
    }
    // The camera follows rather than staying pinned.
    expect(seen.size).toBeGreaterThan(1)
  })

  it('keeps sticky lastStart when the drawable budget changes', () => {
    const cache: TorrentCache = {
      PiecesCount: 5000,
      PiecesLength: 1024,
      Readers: [{ Reader: 1200, Start: 1100, End: 1300 }],
    }
    const first = resolveFocusWindow(cache, 200)!
    const next = resolveFocusWindow(cache, 360, { lastWindowStart: first.start })!
    // Budget grew; sticky start is kept (clamped) so the head is not re-anchored.
    expect(next.start).toBe(first.start)
    expect(next.end - next.start + 1).toBe(360)
    expect(next.start).toBeLessThanOrEqual(1200)
    expect(next.end).toBeGreaterThanOrEqual(1200)
  })

  it('reports the window size without resolving its position', () => {
    const cache: TorrentCache = {
      PiecesCount: 5000,
      PiecesLength: 2 * 1024 * 1024,
      Capacity: 256 * 1024 * 1024,
      Readers: [{ Reader: 1005, Start: 1000, End: 1127 }],
    }
    const size = resolveFocusWindowSize(cache, 396)
    const window = resolveFocusWindow(cache, 396)!
    expect(size).toBe(396)
    expect(size).toBe(window.end - window.start + 1)
    expect(resolveFocusWindowSize({ PiecesCount: 0 }, 396)).toBe(0)
  })

  it('reports readerActive false and marks the head idle when the reader is stopped', () => {
    const cache: TorrentCache = {
      PiecesCount: 200,
      PiecesLength: 1024,
      Capacity: 40 * 1024,
      Readers: [{ Reader: 80, Start: 70, End: 100, Active: false }],
    }
    const window = resolveFocusWindow(cache, 40)!
    expect(window.readerActive).toBe(false)
    // The frozen head still positions the camera so it stays on screen.
    expect(window.readerPiece).toBe(80)

    const head = buildFocusModel(cache, 40).cells.find(c => c.isReader)
    expect(head?.isReaderIdle).toBe(true)
  })

  it('ignores idle readers when an active one exists', () => {
    const cache: TorrentCache = {
      PiecesCount: 400,
      PiecesLength: 1024,
      Capacity: 40 * 1024,
      Readers: [
        { Reader: 300, Start: 290, End: 320, Active: false },
        { Reader: 100, Start: 90, End: 120, Active: true },
      ],
    }
    const window = resolveFocusWindow(cache, 40)!
    expect(window.readerActive).toBe(true)
    // Furthest-ahead is picked among active readers only, so the idle 300 loses.
    expect(window.readerPiece).toBe(100)

    const head = buildFocusModel(cache, 40).cells.find(c => c.isReader)
    expect(head?.pieceStart).toBe(100)
    expect(head?.isReaderIdle).toBe(false)
  })

  it('treats a missing Active flag as active (older servers)', () => {
    const cache: TorrentCache = {
      PiecesCount: 200,
      PiecesLength: 1024,
      Capacity: 40 * 1024,
      Readers: [{ Reader: 80, Start: 70, End: 100 }],
    }
    expect(resolveFocusWindow(cache, 40)!.readerActive).toBe(true)
  })

  it('keeps windowStart when head advances inside dead zone', () => {
    const base: TorrentCache = {
      PiecesCount: 200,
      PiecesLength: 1024,
      Capacity: 40 * 1024,
      Readers: [{ Reader: 80, Start: 70, End: 100 }],
    }
    const first = resolveFocusWindow(base, 40)
    expect(first).not.toBeNull()
    const next = resolveFocusWindow({ ...base, Readers: [{ Reader: 82, Start: 70, End: 100 }] }, 40, {
      lastWindowStart: first!.start,
    })
    expect(next!.start).toBe(first!.start)
    expect(next!.readerPiece).toBe(82)
  })

  it('scrolls window when head passes right margin', () => {
    const cache: TorrentCache = {
      PiecesCount: 200,
      PiecesLength: 1024,
      Capacity: 40 * 1024,
      Readers: [{ Reader: 50, Start: 40, End: 80 }],
    }
    const first = resolveFocusWindow(cache, 40)!
    // Jump head near the right edge of the frozen window
    const rightEdge = first.end - 2
    const scrolled = resolveFocusWindow({ ...cache, Readers: [{ Reader: rightEdge, Start: 40, End: 120 }] }, 40, {
      lastWindowStart: first.start,
    })!
    expect(scrolled.start).toBeGreaterThan(first.start)
    expect(scrolled.readerPiece).toBe(rightEdge)
  })

  it('keeps sticky lastStart when every reader is gone', () => {
    const active: TorrentCache = {
      PiecesCount: 100,
      PiecesLength: 1024,
      Capacity: 20 * 1024,
      Readers: [{ Reader: 40, Start: 30, End: 50 }],
    }
    const first = resolveFocusWindow(active, 20)!
    expect(first.start).toBeGreaterThan(0)
    const closed = resolveFocusWindow({ ...active, Readers: [] }, 20, { lastWindowStart: first.start })!
    expect(closed.readerPiece).toBeNull()
    expect(closed.start).toBe(first.start)
    expect(closed.end).toBe(first.start + 19)
  })

  it('cold-opens on remaining filled pieces when there is no sticky camera', () => {
    const cache: TorrentCache = {
      PiecesCount: 200,
      PiecesLength: 1024,
      Capacity: 40 * 1024,
      Readers: [],
      Pieces: {
        80: { Size: 1024, Length: 1024 },
        90: { Size: 512, Length: 1024 },
        95: { Size: 1024, Length: 1024 },
      },
    }
    const window = resolveFocusWindow(cache, 40)!
    expect(window.readerPiece).toBeNull()
    expect(window.start).toBeLessThanOrEqual(80)
    expect(window.end).toBeGreaterThanOrEqual(95)
  })

  it('cold-opens at piece 0 when there are no readers and no filled pieces', () => {
    const cache: TorrentCache = {
      PiecesCount: 100,
      PiecesLength: 1024,
      Capacity: 20 * 1024,
      Readers: [],
    }
    const window = resolveFocusWindow(cache, 20)!
    expect(window.start).toBe(0)
    expect(window.end).toBe(19)
  })

  it('keeps an idle frozen head on screen when Active is false', () => {
    const cache: TorrentCache = {
      PiecesCount: 200,
      PiecesLength: 1024,
      Capacity: 40 * 1024,
      Readers: [{ Reader: 80, Start: 70, End: 100, Active: false }],
    }
    const window = resolveFocusWindow(cache, 40, { lastWindowStart: 0 })!
    expect(window.readerPiece).toBe(80)
    expect(window.readerActive).toBe(false)
    expect(window.start).toBeLessThanOrEqual(80)
    expect(window.end).toBeGreaterThanOrEqual(80)
  })

  it('builds 1:1 cells with priorities including empty Size=0 entries', () => {
    const cache: TorrentCache = {
      PiecesCount: 20,
      PiecesLength: 100,
      Capacity: 500,
      Pieces: {
        8: { Size: 100, Length: 100, Priority: 2 },
        9: { Size: 0, Length: 100, Priority: 5 },
      },
      Readers: [{ Reader: 9, Start: 7, End: 11 }],
    }
    const model = buildFocusModel(cache, 10)
    expect(model.bucketSize).toBe(1)
    expect(model.cells.length).toBeGreaterThan(0)
    const readerCell = model.cells.find(c => c.isReader)
    expect(readerCell).toBeDefined()
    expect(readerCell!.pieceStart).toBe(9)
    expect(readerCell!.priority).toBe(5)
    expect(readerCell!.percentage).toBe(0)
  })

  it('uses Size/Length for fill even when Completed is set with partial Size', () => {
    const cache: TorrentCache = {
      PiecesCount: 20,
      PiecesLength: 100,
      Capacity: 500,
      Pieces: {
        5: { Size: 10, Length: 100, Completed: true, Priority: 0 },
      },
      Readers: [{ Reader: 5, Start: 5, End: 12 }],
    }
    const model = buildFocusModel(cache, 12)
    const cell = model.cells.find(c => c.pieceStart === 5)
    expect(cell?.percentage).toBe(10)
    expect(cell?.completed).toBe(false)
  })

  it('infers H/R/N/A when API priority is missing on incomplete range pieces', () => {
    const cache: TorrentCache = {
      PiecesCount: 40,
      PiecesLength: 100,
      Capacity: 800,
      Pieces: {},
      Readers: [{ Reader: 10, Start: 10, End: 30 }],
    }
    const model = buildFocusModel(cache, 20)
    const byId = (id: number) => model.cells.find(c => c.pieceStart === id)
    expect(byId(10)?.priority).toBe(5) // A on playhead
    expect(byId(11)?.priority).toBe(4) // N
    expect(byId(12)?.priority).toBeGreaterThanOrEqual(2) // R or H
    // Last incomplete cell in the built window should still be labeled (≥ H)
    const lastIncomplete = [...model.cells].reverse().find(c => !c.completed && !c.isReader)
    expect(lastIncomplete?.priority).toBeGreaterThanOrEqual(2)
    expect(priorityDebugLabel(5)).toBe('A')
    expect(priorityDebugLabel(4)).toBe('N')
    expect(priorityDebugLabel(3)).toBe('R')
    expect(priorityDebugLabel(2)).toBe('H')
  })

  it('keeps A on completed playhead for debug', () => {
    const cache: TorrentCache = {
      PiecesCount: 20,
      PiecesLength: 100,
      Capacity: 500,
      Pieces: {
        5: { Size: 100, Length: 100, Completed: true, Priority: 0 },
      },
      Readers: [{ Reader: 5, Start: 5, End: 12 }],
    }
    const model = buildFocusModel(cache, 12)
    const playhead = model.cells.find(c => c.isReader)
    expect(playhead?.priority).toBe(5)
    expect(priorityDebugLabel(playhead!.priority)).toBe('A')
  })
})
