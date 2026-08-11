import type { CacheMapItem, CachePiece, CacheReader, TorrentCache } from 'shared/api/types'

/** Soft caps for mini LOD view. */
export const SNAKE_MAX_CELLS_DETAILED = 6000

/** Detailed 1:1 window target rows; mini uses fewer. */
export const SNAKE_FOCUS_TARGET_ROWS = 16
export const SNAKE_FOCUS_TARGET_ROWS_MINI = 10

export interface CacheDrawModel {
  cells: CacheMapItem[]
  piecesCount: number
  /** Pieces merged into one cell when > 1 (full-torrent LOD). */
  bucketSize: number
  /** Inclusive piece range covered by cells[0]..cells[n-1]. */
  windowStart?: number
  windowEnd?: number
}

/**
 * Reader.End is inclusive (matches Go `inRanges`: ind >= Start && ind <= End).
 * Returns clamped [start, end] or null if the range is empty / invalid.
 */
export const clampReaderRangeInclusive = (
  start: number | null | undefined,
  end: number | null | undefined,
  piecesCount: number,
): { start: number; end: number } | null => {
  if (start == null || end == null || piecesCount <= 0) return null
  const s = Math.max(0, Math.min(piecesCount - 1, Math.floor(start)))
  const e = Math.max(0, Math.min(piecesCount - 1, Math.floor(end)))
  if (e < s) return null
  return { start: s, end: e }
}

/** Iterate every piece id in an inclusive reader range. */
export const forEachPieceInReaderRange = (
  start: number | null | undefined,
  end: number | null | undefined,
  piecesCount: number,
  fn: (pieceId: number) => void,
) => {
  const range = clampReaderRangeInclusive(start, end, piecesCount)
  if (!range) return
  for (let id = range.start; id <= range.end; id++) fn(id)
}

export const forEachPiece = (pieces: TorrentCache['Pieces'], fn: (id: number, piece: CachePiece) => void) => {
  if (!pieces) return
  if (Array.isArray(pieces)) {
    for (let i = 0; i < pieces.length; i++) {
      const piece = pieces[i]
      if (piece) fn(i, piece)
    }
    return
  }
  for (const [key, piece] of Object.entries(pieces)) {
    if (!piece) continue
    const id = Number(key)
    if (Number.isFinite(id)) fn(id, piece)
  }
}

/** Byte-accurate fill 0–100 from Size/Length (matches Filled Σ Size). */
export const pieceFillPercentage = (piece: CachePiece | undefined, pieceLength: number): number => {
  if (!piece) return 0
  const length = piece.Length || pieceLength || 0
  const rawSize = piece.Size || 0
  // Do not trust Completed alone — API can mark complete while Size is 0/partial,
  // which painted full-green cells that were not counted in Filled.
  if (length <= 0) return 0
  const size = Math.min(rawSize, length)
  return Math.min(100, (size / length) * 100)
}

/**
 * Map anacrolix priorities to snake labels: 2=H, 3=R, 4=N, 5=A.
 * Playhead always displays as A in debug. Incomplete pieces in the reader
 * window get at least H so labels are visible (API often leaves them at 0/1).
 */
export const resolveDisplayPriority = (
  id: number,
  apiPriority: number,
  completed: boolean,
  readers: CacheReader[] | undefined,
): number => {
  if (!readers?.length) return apiPriority >= 2 ? apiPriority : 0

  const onPlayhead = readers.some(r => r.Reader === id)
  if (onPlayhead) return 5

  if (apiPriority >= 2) return apiPriority
  if (completed) return 0

  let best = 0
  for (const r of readers) {
    if (r.Reader == null || r.Start == null || r.End == null) continue
    const readerPos = r.Reader
    const end = r.End
    if (id < readerPos || id > end) continue

    let inferred = 2 // incomplete in-window → at least High (H)
    if (id === readerPos + 1) inferred = 4
    else {
      const span = Math.max(1, end - readerPos)
      const rah = readerPos + Math.max(2, Math.floor(span * 0.45))
      if (id <= rah) inferred = 3
      else if (id <= rah + 5) inferred = 2
    }
    if (inferred > best) best = inferred
  }
  return best > 0 ? best : apiPriority
}

/** Priority → debug letter (master convention). */
export const priorityDebugLabel = (priority: number): string => {
  if (priority === 2) return 'H'
  if (priority === 3) return 'R'
  if (priority === 4) return 'N'
  if (priority === 5) return 'A'
  return ''
}

/**
 * Servers before the Active flag omit it entirely; absent means active so the
 * playhead keeps working against older builds.
 */
export const isReaderActive = (reader: CacheReader): boolean => reader.Active !== false

const cellFromPiece = (
  id: number,
  piece: CachePiece | undefined,
  pieceLength: number,
  isReader: boolean,
  isReaderIdle: boolean,
  isReaderRange: boolean,
  readers: CacheReader[] | undefined,
): CacheMapItem => {
  const percentage = pieceFillPercentage(piece, pieceLength)
  // Visual "complete" follows bytes, not the API Complete flag alone.
  const completed = percentage >= 100
  const apiPriority = piece?.Priority || 0
  return {
    percentage,
    priority: resolveDisplayPriority(id, apiPriority, completed, readers),
    completed,
    isReader,
    isReaderIdle,
    isReaderRange,
    pieceStart: id,
    pieceEnd: id,
  }
}

/**
 * Full-torrent LOD snake model: always covers 0..PiecesCount-1.
 * If there are more pieces than maxCells, adjacent pieces are merged with
 * byte-accurate fill (Size/Length). Kept for tests / possible future overview —
 * production UI uses {@link buildFocusModel} only.
 */
export const buildCacheDrawModel = (cache: TorrentCache, maxCells: number): CacheDrawModel => {
  const piecesCount = cache.PiecesCount ?? 0
  if (piecesCount <= 0 || maxCells < 1) {
    return { cells: [], piecesCount, bucketSize: 1, windowStart: 0, windowEnd: -1 }
  }

  const budget = Math.max(1, Math.min(maxCells, SNAKE_MAX_CELLS_DETAILED))
  const bucketSize = Math.max(1, Math.ceil(piecesCount / budget))
  const numBuckets = Math.ceil(piecesCount / bucketSize)
  const pieceLength = cache.PiecesLength || 0

  const filled = new Float64Array(numBuckets)
  const capacity = new Float64Array(numBuckets)
  const priority = new Int16Array(numBuckets)
  const readerBits = new Uint8Array(numBuckets)
  const rangeBits = new Uint8Array(numBuckets)

  if (pieceLength > 0) {
    for (let b = 0; b < numBuckets; b++) {
      const pieceStart = b * bucketSize
      const pieceEnd = Math.min(pieceStart + bucketSize, piecesCount)
      capacity[b] = (pieceEnd - pieceStart) * pieceLength
    }
  }

  forEachPiece(cache.Pieces, (id, piece) => {
    if (id < 0 || id >= piecesCount) return
    const bucket = Math.floor(id / bucketSize)
    const length = piece.Length || pieceLength || 0
    const rawSize = piece.Size || 0
    const size = length > 0 ? Math.min(rawSize, length) : rawSize
    filled[bucket] += size

    if (pieceLength <= 0 && length > 0) capacity[bucket] += length

    const prio = piece.Priority || 0
    if (prio > priority[bucket]) priority[bucket] = prio
  })

  for (const reader of cache.Readers || []) {
    if (reader.Reader != null && reader.Reader >= 0 && reader.Reader < piecesCount) {
      readerBits[Math.floor(reader.Reader / bucketSize)] = 1
    }
    forEachPieceInReaderRange(reader.Start, reader.End, piecesCount, id => {
      rangeBits[Math.floor(id / bucketSize)] = 1
    })
  }

  const cells: CacheMapItem[] = new Array(numBuckets)
  for (let b = 0; b < numBuckets; b++) {
    const cap = capacity[b]
    let percentage = 0
    if (cap > 0) percentage = Math.min(100, (filled[b] / cap) * 100)
    const completed = cap > 0 && filled[b] >= cap - 0.5
    const pieceStart = b * bucketSize
    const pieceEnd = Math.min(pieceStart + bucketSize, piecesCount) - 1

    cells[b] = {
      percentage: completed ? 100 : percentage,
      priority: priority[b],
      completed,
      isReader: Boolean(readerBits[b]),
      isReaderRange: Boolean(rangeBits[b]),
      pieceStart,
      pieceEnd,
    }
  }

  return {
    cells,
    piecesCount,
    bucketSize,
    windowStart: 0,
    windowEnd: piecesCount - 1,
  }
}

export interface FocusWindow {
  start: number
  end: number
  /** Furthest reader piece, or null when there is no reader at all. */
  readerPiece: number | null
  /** False when every reader is idle — the playhead is frozen. */
  readerActive: boolean
}

export interface FocusWindowOptions {
  /** Previous window start — enables dead-zone camera while a reader is active. */
  lastWindowStart?: number
  /** Fraction of window kept as edge margin before scrolling (default 0.18). */
  edgeMarginRatio?: number
}

/**
 * Slack cells on each side of the reader range. This doubles as the distance the
 * head may walk before the camera has to follow, so it keeps the playhead
 * visibly moving instead of pinning it to a fixed spot.
 */
const READER_RANGE_MARGIN = 8

/**
 * Widest inclusive span across the given readers, or null when none report a
 * usable range. Used to keep the buffered tail on screen while the head walks.
 */
const resolveReaderRangeSpan = (readers: CacheReader[], piecesCount: number): { start: number; end: number } | null => {
  let start = Number.POSITIVE_INFINITY
  let end = Number.NEGATIVE_INFINITY
  for (const reader of readers) {
    const range = clampReaderRangeInclusive(reader.Start, reader.End, piecesCount)
    if (!range) continue
    if (range.start < start) start = range.start
    if (range.end > end) end = range.end
  }
  if (!Number.isFinite(start) || !Number.isFinite(end)) return null
  return { start, end }
}

/** Readers that drive the camera: active ones, or every reader when all are idle. */
const resolveDrivingReaders = (cache: TorrentCache): { readers: CacheReader[]; readerActive: boolean } => {
  const all = cache.Readers || []
  const active = all.filter(isReaderActive)
  return { readers: active.length > 0 ? active : all, readerActive: active.length > 0 }
}

/**
 * Inclusive span of pieces that still hold bytes. Server memory can keep a mid-file
 * buffer after every reader is gone — the camera uses this for a cold open.
 */
const resolveFilledSpan = (cache: TorrentCache, piecesCount: number): { start: number; end: number } | null => {
  let start = Number.POSITIVE_INFINITY
  let end = Number.NEGATIVE_INFINITY
  forEachPiece(cache.Pieces, (id, piece) => {
    if (id < 0 || id >= piecesCount) return
    if ((piece.Size || 0) <= 0) return
    if (id < start) start = id
    if (id > end) end = id
  })
  if (!Number.isFinite(start) || !Number.isFinite(end)) return null
  return { start, end }
}

/** Window start that keeps as much of `span` visible as possible (centered when it fits). */
const anchorStartForSpan = (span: { start: number; end: number }, windowSize: number): number => {
  const spanWidth = span.end - span.start + 1
  if (spanWidth <= windowSize) {
    return span.start - Math.floor((windowSize - spanWidth) / 2)
  }
  return span.start + Math.floor((spanWidth - windowSize) / 2)
}

/** Full drawable budget — small piece cells fill the pane; do not shrink to the reader range. */
const resolveWindowSize = (visibleCells: number, piecesCount: number): number =>
  Math.max(1, Math.min(piecesCount, Math.max(1, visibleCells)))

const clampWindow = (start: number, windowSize: number, piecesCount: number): { start: number; end: number } => {
  let s = start
  if (s < 0) s = 0
  let end = s + windowSize - 1
  if (end >= piecesCount) {
    end = piecesCount - 1
    s = Math.max(0, end - windowSize + 1)
  }
  return { start: s, end }
}

/**
 * How many cells the window will hold (full drawable budget), without resolving
 * its position — useful for layout that needs the count before the model exists.
 */
export const resolveFocusWindowSize = (cache: TorrentCache, visibleCells: number): number => {
  const piecesCount = cache.PiecesCount ?? 0
  if (piecesCount <= 0) return 0
  return resolveWindowSize(visibleCells, piecesCount)
}

/**
 * Sliding 1:1 window filling the drawable grid, positioned by a dead-zone
 * camera: the head walks across cells and the window scrolls only once it
 * nears an edge. No playhead: keep sticky lastStart, else center on remaining
 * filled pieces (server cache can outlive readers), else piece 0.
 */
export const resolveFocusWindow = (
  cache: TorrentCache,
  visibleCells: number,
  options?: FocusWindowOptions,
): FocusWindow | null => {
  const piecesCount = cache.PiecesCount ?? 0
  if (piecesCount <= 0) return null

  // Idle readers still position the camera so the frozen head stays on screen.
  const { readers: drivingReaders, readerActive } = resolveDrivingReaders(cache)

  let readerPiece: number | null = null
  if (drivingReaders.length > 0) {
    // Prefer the furthest-ahead reader so preload/stream progress drives the window
    // (min id stuck the view at piece 0 when dual preload readers exist).
    let bestReader = -1
    for (const r of drivingReaders) {
      if (r.Reader != null && r.Reader >= 0 && r.Reader < piecesCount && r.Reader > bestReader) {
        bestReader = r.Reader
      }
    }
    if (bestReader >= 0) readerPiece = bestReader
  }

  const range = resolveReaderRangeSpan(drivingReaders, piecesCount)
  const windowSize = resolveWindowSize(visibleCells, piecesCount)
  const lastStart = options?.lastWindowStart
  const marginRatio = options?.edgeMarginRatio ?? 0.18
  const margin = Math.max(1, Math.floor(windowSize * marginRatio))

  // No playhead (player closed): sticky camera first — usually still overlaps the
  // server-held buffer. Cold open: frame remaining filled pieces, else piece 0.
  if (readerPiece == null) {
    let preferred: number
    if (lastStart != null) {
      preferred = lastStart
    } else {
      const filled = resolveFilledSpan(cache, piecesCount)
      if (filled) preferred = anchorStartForSpan(filled, windowSize)
      else if (range) preferred = range.start - READER_RANGE_MARGIN
      else preferred = 0
    }
    const { start, end } = clampWindow(preferred, windowSize, piecesCount)
    return { start, end, readerPiece: null, readerActive }
  }

  let start: number
  if (lastStart == null) {
    // First frame: anchor to the reader range. The head sits near its left edge
    // because everything behind it is evicted — the buffer extends to the right.
    start = range ? range.start - READER_RANGE_MARGIN : readerPiece - Math.floor(windowSize / 2)
  } else {
    start = lastStart
    const tentativeEnd = start + windowSize - 1
    const leftBound = start + margin
    const rightBound = tentativeEnd - margin
    if (readerPiece < leftBound) {
      start = readerPiece - margin
    } else if (readerPiece > rightBound) {
      start = readerPiece - (windowSize - 1 - margin)
    }
  }

  // The dead zone alone would drift the window until the buffered tail falls off
  // the grid, so pin it to a band where the whole reader range stays visible.
  // The band is 2 × READER_RANGE_MARGIN wide, which is the head's walking room.
  const rangeLo = range ? range.end - windowSize + 1 : null
  const rangeHi = range ? range.start : null
  if (rangeLo != null && rangeHi != null && rangeLo <= rangeHi) {
    if (start < rangeLo) start = rangeLo
    if (start > rangeHi) start = rangeHi
  } else {
    // Range wider than the grid (or absent): just keep the head on screen,
    // with an edge margin when there is room for one.
    const keepIn = windowSize > margin * 2 ? margin : 0
    const maxStart = readerPiece - keepIn
    const minStart = readerPiece + keepIn - windowSize + 1
    if (start > maxStart) start = maxStart
    if (start < minStart) start = minStart
  }

  const clamped = clampWindow(start, windowSize, piecesCount)
  return { start: clamped.start, end: clamped.end, readerPiece, readerActive }
}

/**
 * 1:1 focus model: one cell per piece in [window.start, window.end].
 */
export const buildFocusModel = (
  cache: TorrentCache,
  visibleCells: number,
  options?: FocusWindowOptions,
): CacheDrawModel => {
  const piecesCount = cache.PiecesCount ?? 0
  const window = resolveFocusWindow(cache, visibleCells, options)
  if (!window || piecesCount <= 0) {
    return { cells: [], piecesCount, bucketSize: 1, windowStart: 0, windowEnd: -1 }
  }

  const pieceLength = cache.PiecesLength || 0
  const readers = cache.Readers || []
  const pieceById = new Map<number, CachePiece>()
  forEachPiece(cache.Pieces, (id, piece) => {
    if (id >= window.start && id <= window.end) pieceById.set(id, piece)
  })

  const readerSet = new Set<number>()
  const activeReaderSet = new Set<number>()
  const rangeSet = new Set<number>()
  for (const reader of readers) {
    if (reader.Reader != null && reader.Reader >= window.start && reader.Reader <= window.end) {
      readerSet.add(reader.Reader)
      if (isReaderActive(reader)) activeReaderSet.add(reader.Reader)
    }
    forEachPieceInReaderRange(reader.Start, reader.End, piecesCount, id => {
      if (id >= window.start && id <= window.end) rangeSet.add(id)
    })
  }
  // A piece shared by an active and an idle reader counts as active.
  const isIdleHead = (id: number) => readerSet.has(id) && !activeReaderSet.has(id)

  const cells: CacheMapItem[] = []
  for (let id = window.start; id <= window.end; id++) {
    cells.push(
      cellFromPiece(id, pieceById.get(id), pieceLength, readerSet.has(id), isIdleHead(id), rangeSet.has(id), readers),
    )
  }

  return {
    cells,
    piecesCount,
    bucketSize: 1,
    windowStart: window.start,
    windowEnd: window.end,
  }
}

/** Visible cell budget for the 1:1 window (cols × rows). */
export const resolveFocusVisibleCells = (containerWidth: number, isMini = false, containerHeight = 0): number => {
  const defaultRows = isMini ? SNAKE_FOCUS_TARGET_ROWS_MINI : SNAKE_FOCUS_TARGET_ROWS
  const cellFootprint = isMini ? 26 + 5 : 20 + 4
  if (!containerWidth || containerWidth <= 0) return 10 * defaultRows
  const cols = Math.max(1, Math.floor(containerWidth / cellFootprint))
  // Detailed cache tab: fit rows to the available height so the map never needs a scrollbar.
  // Ignore sub-threshold heights (collapsed flex) — fall back to target rows until RO settles.
  const rows =
    !isMini && containerHeight >= cellFootprint * 2
      ? Math.max(2, Math.floor(containerHeight / cellFootprint))
      : defaultRows
  return cols * rows
}
