import { memo, useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { CacheMapItem, TorrentCache as TorrentCacheData } from 'shared/api/types'
import { priorityDebugLabel, resolveFocusVisibleCells, resolveFocusWindow } from 'shared/cache/buildCacheMap'
import { drawSnake, hitTestSnakeCell, setupHiDpiCanvas } from 'shared/cache/drawSnake'
import { resolvePieceMetrics, resolveSnakeSettings, type SnakeThemeMode } from 'shared/cache/snakeSettings'
import { useCreateFocusMap } from 'shared/cache/useUpdateCache'
import { humanizeSize } from 'shared/lib/format'
import { useThemePreference } from 'shared/theme/useThemePreference'

export type SnakeViewMode = 'detailed' | 'mini'

export interface TorrentCacheProps {
  cache: TorrentCacheData
  /** detailed/mini — both use a 1:1 reader window (no LOD merge). */
  mode?: SnakeViewMode
  isSnakeDebugMode?: boolean
}

/** Resume auto-follow this long after the user last scrolled/touched the map. */
const FOLLOW_RESUME_DELAY_MS = 4000

const emptyCell = (): CacheMapItem => ({
  percentage: 0,
  priority: 0,
  isReader: false,
  isReaderRange: false,
})

const piecesFingerprint = (pieces: TorrentCacheData['Pieces']) => {
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

const readersFingerprint = (readers: TorrentCacheData['Readers']) => {
  if (!readers?.length) return ''
  return readers.map(r => `${r.Reader ?? ''}:${r.Start ?? ''}-${r.End ?? ''}`).join('|')
}

/** Canvas-based piece map ("snake") showing cache fill, playhead and priorities. */
function TorrentCache({ cache, mode = 'detailed', isSnakeDebugMode }: TorrentCacheProps) {
  const { t } = useTranslation()
  const { isDark, palette } = useThemePreference()
  const theme: SnakeThemeMode = isDark ? 'dark' : 'light'
  const isMiniView = mode === 'mini'

  const [containerWidth, setContainerWidth] = useState(0)
  const [containerHeight, setContainerHeight] = useState(0)
  const canvasRef = useRef<HTMLCanvasElement | null>(null)
  const scrollWrapperRef = useRef<HTMLDivElement | null>(null)
  const rootRef = useRef<HTMLDivElement | null>(null)
  const lastWindowStartRef = useRef<number | undefined>(undefined)
  const drawFrame = useRef(0)
  const isFollowingPlayhead = useRef(true)
  const resumeFollowTimer = useRef(0)
  const [tooltip, setTooltip] = useState<{ index: number; x: number; y: number; text: string } | null>(null)

  useEffect(() => {
    const el = isMiniView ? rootRef.current : scrollWrapperRef.current
    if (!el || typeof ResizeObserver === 'undefined') return
    const observer = new ResizeObserver(entries => {
      const box = entries[0]?.contentRect
      setContainerWidth(box?.width ?? 0)
      if (!isMiniView) setContainerHeight(box?.height ?? 0)
    })
    observer.observe(el)
    const rect = el.getBoundingClientRect()
    setContainerWidth(rect.width)
    if (!isMiniView) setContainerHeight(rect.height)
    return () => observer.disconnect()
  }, [isMiniView])

  const visibleCellBudget = useMemo(
    () => resolveFocusVisibleCells(containerWidth, isMiniView, isMiniView ? 0 : containerHeight),
    [containerWidth, containerHeight, isMiniView],
  )
  const focusModel = useCreateFocusMap(cache, visibleCellBudget)
  const cells = focusModel.cells
  const hasActiveReaders = (cache.Readers?.length ?? 0) > 0

  useEffect(() => {
    if (focusModel.windowStart != null) lastWindowStartRef.current = focusModel.windowStart
  }, [focusModel.windowStart])

  const variant = isMiniView ? 'mini' : 'default'
  // Re-resolve when palette changes so canvas accents track CSS `--accent`.
  const baseSettings = useMemo(() => {
    void palette
    return resolveSnakeSettings(theme, variant)
  }, [theme, variant, palette])

  const { pieceSize, gap } = useMemo(
    () => resolvePieceMetrics(baseSettings, containerWidth, isMiniView, cells.length),
    [baseSettings, containerWidth, isMiniView, cells.length],
  )

  const canvasWidth =
    containerWidth > 0 ? (isMiniView ? Math.max(containerWidth - 8, containerWidth * 0.96) : containerWidth) : 0
  const cellStride = pieceSize + gap
  const piecesPerRow = canvasWidth > 0 ? Math.max(1, Math.floor(canvasWidth / cellStride)) : 0

  const emptyRowCount = isMiniView ? 4 : 6
  const canvasHeight =
    piecesPerRow > 0
      ? Math.max(cells.length > 0 ? Math.ceil(cells.length / piecesPerRow) : emptyRowCount, emptyRowCount) * cellStride
      : 0
  /** Reserve mini shell height before ResizeObserver so Overview actions don't jump when the snake mounts. */
  const miniShellMinHeight = isMiniView ? emptyRowCount * cellStride + 16 : undefined

  const startingX = piecesPerRow > 0 ? Math.ceil((canvasWidth - cellStride * piecesPerRow) / 2) : 0

  const drawCells = useMemo(
    () => (cells.length > 0 ? cells : Array.from({ length: Math.max(piecesPerRow, 1) * emptyRowCount }, emptyCell)),
    [cells, piecesPerRow, emptyRowCount],
  )

  const cacheAriaLabel = useMemo(() => {
    const { Filled, Capacity } = cache
    if (Filled != null && Capacity != null) {
      return t('SnakeCacheSummary', {
        filled: humanizeSize(Filled),
        capacity: humanizeSize(Capacity),
      })
    }
    return t('Cache')
  }, [cache, t])

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas || !canvasWidth || !canvasHeight || !piecesPerRow) return

    cancelAnimationFrame(drawFrame.current)
    drawFrame.current = requestAnimationFrame(() => {
      const ctx = setupHiDpiCanvas(canvas, canvasWidth, canvasHeight)
      if (!ctx) return
      drawSnake({
        ctx,
        cells: drawCells,
        canvasWidth,
        canvasHeight,
        piecesInOneRow: piecesPerRow,
        pieceSize,
        gap,
        startingX,
        theme,
        variant,
        isSnakeDebugMode,
        isMini: isMiniView,
      })
    })

    return () => cancelAnimationFrame(drawFrame.current)
  }, [
    canvasHeight,
    canvasWidth,
    piecesPerRow,
    startingX,
    pieceSize,
    gap,
    drawCells,
    variant,
    isMiniView,
    theme,
    palette,
    isSnakeDebugMode,
  ])

  useEffect(() => {
    /** The mini preview never scrolls internally (see render below), so there's nothing to follow. */
    if (isMiniView) return
    if (!scrollWrapperRef.current || !piecesPerRow || cellStride <= 0) return
    if (!isFollowingPlayhead.current) return
    if (!hasActiveReaders) return
    const focusWindow = resolveFocusWindow(cache, visibleCellBudget, {
      lastWindowStart: lastWindowStartRef.current,
    })
    if (!focusWindow || focusWindow.readerPiece == null || focusModel.windowStart == null) return
    const localIndex = focusWindow.readerPiece - focusModel.windowStart
    if (localIndex < 0) return
    const row = Math.floor(localIndex / piecesPerRow)
    const rowTop = row * cellStride
    const el = scrollWrapperRef.current
    const viewTop = el.scrollTop
    const viewBottom = viewTop + el.clientHeight
    if (rowTop < viewTop || rowTop + cellStride > viewBottom) {
      el.scrollTop = Math.max(0, rowTop - el.clientHeight / 3)
    }
  }, [
    cache,
    visibleCellBudget,
    focusModel.windowStart,
    piecesPerRow,
    cellStride,
    drawCells,
    isMiniView,
    hasActiveReaders,
  ])

  useEffect(() => {
    if (isMiniView) return
    const el = scrollWrapperRef.current
    if (!el) return
    const pauseFollowing = () => {
      isFollowingPlayhead.current = false
      window.clearTimeout(resumeFollowTimer.current)
      resumeFollowTimer.current = window.setTimeout(() => {
        isFollowingPlayhead.current = true
      }, FOLLOW_RESUME_DELAY_MS)
    }
    el.addEventListener('wheel', pauseFollowing, { passive: true })
    el.addEventListener('touchstart', pauseFollowing, { passive: true })
    el.addEventListener('pointerdown', pauseFollowing)
    return () => {
      window.clearTimeout(resumeFollowTimer.current)
      el.removeEventListener('wheel', pauseFollowing)
      el.removeEventListener('touchstart', pauseFollowing)
      el.removeEventListener('pointerdown', pauseFollowing)
    }
  }, [containerWidth, isMiniView])

  const formatTooltipText = useCallback(
    (cell: CacheMapItem) => {
      const start = cell.pieceStart
      const end = cell.pieceEnd
      if (start == null) return ''
      const fillPercent = cell.completed || (cell.percentage || 0) >= 99.5 ? 100 : Math.round(cell.percentage || 0)
      const priorityLabel = priorityDebugLabel(cell.priority || 0)
      const priorityPart = priorityLabel ? ` · ${priorityLabel}` : ''
      if (end != null && end !== start) {
        return t('SnakeTooltipBucket', { start, end, fill: fillPercent }) + priorityPart
      }
      return t('SnakeTooltipPiece', { id: start, fill: fillPercent }) + priorityPart
    },
    [t],
  )

  const cellAtPoint = useCallback(
    (clientX: number, clientY: number) => {
      if (!piecesPerRow) return null
      const canvas = canvasRef.current
      const root = rootRef.current
      if (!canvas || !root) return null
      const canvasRect = canvas.getBoundingClientRect()
      const rootRect = root.getBoundingClientRect()
      const localX = clientX - canvasRect.left
      const localY = clientY - canvasRect.top
      const index = hitTestSnakeCell(localX, localY, {
        piecesInOneRow: piecesPerRow,
        pieceSize,
        gap,
        startingX,
        cellCount: drawCells.length,
      })
      if (index < 0) return null
      const text = formatTooltipText(drawCells[index])
      if (!text) return null
      return { index, x: clientX - rootRect.left + 12, y: clientY - rootRect.top + 12, text }
    },
    [piecesPerRow, pieceSize, gap, startingX, drawCells, formatTooltipText],
  )

  const handleCanvasMove = useCallback(
    (event: React.MouseEvent<HTMLCanvasElement>) => {
      setTooltip(cellAtPoint(event.clientX, event.clientY))
    },
    [cellAtPoint],
  )

  /** Touch has no hover — tap a piece to pin its tooltip, tap it again (or elsewhere) to dismiss. */
  const handleCanvasTap = useCallback(
    (event: React.MouseEvent<HTMLCanvasElement>) => {
      const next = cellAtPoint(event.clientX, event.clientY)
      setTooltip(current => (next && current?.index === next.index ? null : next))
    },
    [cellAtPoint],
  )

  useEffect(() => {
    if (!tooltip) return
    const dismissIfOutside = (event: PointerEvent) => {
      if (!(event.target instanceof Node) || !rootRef.current?.contains(event.target)) {
        setTooltip(null)
      }
    }
    document.addEventListener('pointerdown', dismissIfOutside)
    return () => document.removeEventListener('pointerdown', dismissIfOutside)
  }, [tooltip])

  return (
    <div ref={rootRef} className={`relative flex w-full min-w-0 flex-col ${isMiniView ? '' : 'min-h-0 flex-1'}`}>
      <div
        ref={scrollWrapperRef}
        className={`relative w-full min-w-0 rounded-lg border border-border bg-surface-secondary p-2 ${
          isMiniView
            ? 'grid max-h-[420px] justify-center overflow-hidden'
            : 'min-h-0 min-w-0 flex-1 overflow-auto overscroll-contain'
        }`}
        style={
          isMiniView
            ? miniShellMinHeight != null
              ? { minHeight: miniShellMinHeight }
              : undefined
            : { WebkitOverflowScrolling: 'touch' }
        }
      >
        {piecesPerRow > 0 && canvasHeight > 0 ? (
          <canvas
            ref={canvasRef}
            role='img'
            aria-label={cacheAriaLabel}
            className='block max-w-full'
            onMouseMove={handleCanvasMove}
            onMouseLeave={() => setTooltip(null)}
            onClick={handleCanvasTap}
          />
        ) : null}
      </div>

      {isMiniView ? (
        <p className='mt-1.5 text-center text-[10px] uppercase tracking-wider text-muted/70'>{t('ScrollDown')}</p>
      ) : null}

      {tooltip ? (
        <div
          className='pointer-events-none absolute z-20 whitespace-nowrap rounded-md border border-border bg-surface-tertiary px-2 py-1 text-xs leading-snug text-foreground shadow-lg'
          style={{ left: tooltip.x, top: tooltip.y }}
        >
          {tooltip.text}
        </div>
      ) : null}

      {focusModel.windowStart != null &&
      focusModel.windowEnd != null &&
      focusModel.windowEnd >= focusModel.windowStart ? (
        <p className='mt-2 shrink-0 self-center text-xs uppercase tracking-wide text-muted'>
          {t('SnakeFocusRange', { start: focusModel.windowStart, end: focusModel.windowEnd })}
          {!hasActiveReaders ? ` · ${t('SnakeIdleFrozen')}` : null}
        </p>
      ) : null}
    </div>
  )
}

export default memo(TorrentCache, (prev, next) => {
  if (prev.mode !== next.mode) return false
  if (prev.isSnakeDebugMode !== next.isSnakeDebugMode) return false
  const a = prev.cache
  const b = next.cache
  return (
    a.PiecesCount === b.PiecesCount &&
    a.PiecesLength === b.PiecesLength &&
    a.Capacity === b.Capacity &&
    a.Filled === b.Filled &&
    piecesFingerprint(a.Pieces) === piecesFingerprint(b.Pieces) &&
    readersFingerprint(a.Readers) === readersFingerprint(b.Readers)
  )
})
