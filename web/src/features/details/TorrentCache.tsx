import { memo, useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { CacheMapItem, TorrentCache as TorrentCacheData } from 'shared/api/types'
import {
  isReaderActive,
  priorityDebugLabel,
  resolveFocusVisibleCells,
  SNAKE_FOCUS_TARGET_ROWS,
  SNAKE_FOCUS_TARGET_ROWS_MINI,
} from 'shared/cache/buildCacheMap'
import { cheapPiecesFingerprint, cheapReadersFingerprint } from 'shared/cache/cacheFingerprint'
import { drawSnake, hitTestSnakeCell, setupHiDpiCanvas } from 'shared/cache/drawSnake'
import { resolvePieceMetrics, resolveSnakeSettings, type SnakeThemeMode } from 'shared/cache/snakeSettings'
import { snakeCameraKey } from 'shared/cache/snakeSession'
import { useCreateFocusMap } from 'shared/cache/useUpdateCache'
import { humanizeSize } from 'shared/lib/format'
import { useThemePreference } from 'shared/theme/useThemePreference'

export type SnakeViewMode = 'detailed' | 'mini'

export interface TorrentCacheProps {
  cache: TorrentCacheData
  /** detailed — 1:1 reader window sized to the drawable grid. */
  mode?: SnakeViewMode
  isSnakeDebugMode?: boolean
  /** Torrent hash — restores the focus window after the dialog or tab is remounted. */
  hash?: string
}

/**
 * ResizeObserver reports the content box, so the first synchronous read must
 * subtract padding too — otherwise the grid is laid out for a pane ~18px wider
 * than reality and reflows right after mount. `clientWidth` is used instead of
 * `getBoundingClientRect` because the modal enter animation scales the subtree.
 */
const measureContentBox = (el: HTMLElement) => {
  const style = getComputedStyle(el)
  const width = el.clientWidth - (parseFloat(style.paddingLeft) || 0) - (parseFloat(style.paddingRight) || 0)
  const height = el.clientHeight - (parseFloat(style.paddingTop) || 0) - (parseFloat(style.paddingBottom) || 0)
  return { width: Math.max(0, width), height: Math.max(0, height) }
}

const emptyCell = (): CacheMapItem => ({
  percentage: 0,
  priority: 0,
  isReader: false,
  isReaderRange: false,
})

/** Canvas-based piece map ("snake") showing cache fill, playhead and priorities. */
function TorrentCache({ cache, mode = 'detailed', isSnakeDebugMode, hash }: TorrentCacheProps) {
  const { t } = useTranslation()
  const { isDark, palette } = useThemePreference()
  const theme: SnakeThemeMode = isDark ? 'dark' : 'light'
  const isMiniView = mode === 'mini'

  const [containerWidth, setContainerWidth] = useState(0)
  const [containerHeight, setContainerHeight] = useState(0)
  const canvasRef = useRef<HTMLCanvasElement | null>(null)
  const scrollWrapperRef = useRef<HTMLDivElement | null>(null)
  const rootRef = useRef<HTMLDivElement | null>(null)
  const drawFrame = useRef(0)
  const lastDrawKey = useRef('')
  const [tooltip, setTooltip] = useState<{ index: number; x: number; y: number; text: string } | null>(null)

  // Layout effect: measuring after paint showed an empty pane for one frame on
  // every mount (dialog open, tab switch) before the snake appeared.
  // ResizeObserver updates are applied synchronously — rAF delayed the second
  // measure by a paint and flashed a wrong-sized grid.
  useLayoutEffect(() => {
    const el = isMiniView ? rootRef.current : scrollWrapperRef.current
    if (!el || typeof ResizeObserver === 'undefined') return
    const apply = (width: number, height: number) => {
      setContainerWidth(width)
      if (!isMiniView) setContainerHeight(height)
    }
    const observer = new ResizeObserver(entries => {
      const box = entries[0]?.contentRect
      apply(box?.width ?? 0, box?.height ?? 0)
    })
    observer.observe(el)
    const box = measureContentBox(el)
    apply(box.width, box.height)
    return () => observer.disconnect()
  }, [isMiniView])

  const variant = isMiniView ? 'mini' : 'default'
  const baseSettings = useMemo(() => {
    void palette
    return resolveSnakeSettings(theme, variant)
  }, [theme, variant, palette])

  // Detailed: fixed ~20px piece cells. The pane is filled by showing more pieces
  // (full cols×rows budget), not by growing individual squares.
  const { pieceSize, gap } = useMemo(
    () => resolvePieceMetrics(baseSettings, containerWidth, isMiniView, 0),
    [baseSettings, containerWidth, isMiniView],
  )

  const canvasWidth =
    containerWidth > 0 ? (isMiniView ? Math.max(containerWidth - 8, containerWidth * 0.96) : containerWidth) : 0

  const emptyRowCount = isMiniView ? 4 : 6
  const targetRows = isMiniView ? SNAKE_FOCUS_TARGET_ROWS_MINI : SNAKE_FOCUS_TARGET_ROWS

  const cellStride = pieceSize + gap
  const piecesPerRow = canvasWidth > 0 ? Math.max(1, Math.floor(canvasWidth / cellStride)) : 0

  // Detailed: fit rows to pane once RO reports real height (≥2 footprints).
  // Until then use targetRows so a Getting-Info placeholder still paints a full
  // grid instead of a short strip with empty space below.
  const heightReady = !isMiniView && cellStride > 0 && containerHeight >= cellStride * 2
  const maxFitRows = heightReady
    ? Math.max(2, Math.floor(containerHeight / cellStride))
    : isMiniView
      ? emptyRowCount
      : targetRows

  // Drawable budget first — focus window must match cols × rows (no silent slice).
  const visibleCellBudget =
    piecesPerRow > 0
      ? piecesPerRow * maxFitRows
      : resolveFocusVisibleCells(containerWidth, isMiniView, isMiniView ? 0 : containerHeight)

  // Persist camera only after the pane height is real — placeholder row counts
  // must not overwrite the saved window. Still paint as soon as width is known.
  const cameraReady = isMiniView ? containerWidth > 0 : heightReady
  const focusModel = useCreateFocusMap(cache, visibleCellBudget, cameraReady ? snakeCameraKey(hash, mode) : undefined)
  const cells = focusModel.cells
  // Idle heads stay on screen but stop moving — only then show "frozen".
  const hasIdleHead = (cache.Readers ?? []).some(r => r.Reader != null && !isReaderActive(r))

  // No PiecesCount yet (Getting Info): fill the drawable pane with empty cells
  // so the Cache tab never shows a half-empty gray box.
  const placeholderRows = isMiniView ? emptyRowCount : maxFitRows
  const placeholderCount = Math.max(piecesPerRow, 1) * placeholderRows

  const rowCount =
    piecesPerRow > 0
      ? Math.max(
          cells.length > 0 ? Math.ceil(cells.length / piecesPerRow) : placeholderRows,
          isMiniView ? emptyRowCount : 1,
        )
      : 0
  const fittedRows = Math.min(rowCount, maxFitRows)
  const canvasHeight = fittedRows > 0 ? fittedRows * cellStride : 0

  const startingX = piecesPerRow > 0 ? Math.ceil((canvasWidth - cellStride * piecesPerRow) / 2) : 0

  const drawCells = useMemo(() => {
    if (cells.length > 0) return cells
    return Array.from({ length: placeholderCount }, emptyCell)
  }, [cells, placeholderCount])

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

  // Footer range matches drawn cells (window is sized to the grid).
  const footerStart = focusModel.windowStart
  const footerEnd =
    focusModel.windowStart != null && drawCells.length > 0 && drawCells[0]?.pieceStart != null
      ? (drawCells[drawCells.length - 1]?.pieceEnd ??
        drawCells[drawCells.length - 1]?.pieceStart ??
        focusModel.windowEnd)
      : focusModel.windowEnd

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas || !canvasWidth || !canvasHeight || !piecesPerRow) return

    const drawKey = [
      canvasWidth,
      canvasHeight,
      piecesPerRow,
      pieceSize,
      gap,
      isSnakeDebugMode ? 1 : 0,
      theme,
      palette,
      cheapPiecesFingerprint(cache.Pieces),
      cheapReadersFingerprint(cache.Readers),
      drawCells.length,
      footerStart ?? -1,
    ].join('|')
    if (drawKey === lastDrawKey.current) return
    lastDrawKey.current = drawKey

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
    cache.Pieces,
    cache.Readers,
    footerStart,
  ])

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

  // Paint as soon as width is known — Getting Info still gets a full empty grid.
  // Footer (piece range) only when the focus model has real piece ids.
  const showCanvas = containerWidth > 0 && piecesPerRow > 0 && canvasHeight > 0
  const showFooter =
    showCanvas && cells.length > 0 && footerStart != null && footerEnd != null && footerEnd >= footerStart

  return (
    <div ref={rootRef} className={`relative flex w-full min-w-0 flex-col ${isMiniView ? '' : 'min-h-0 flex-1'}`}>
      <div
        ref={scrollWrapperRef}
        className={`ts-details-cache-snake relative w-full min-w-0 rounded-lg border border-border bg-surface-secondary p-2 ${
          isMiniView ? 'grid max-h-[420px] justify-center overflow-hidden' : 'min-h-0 min-w-0 flex-1 overflow-hidden'
        }`}
      >
        {showCanvas ? (
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

      {tooltip ? (
        <div
          className='pointer-events-none absolute z-20 whitespace-nowrap rounded-md border border-border bg-surface-tertiary px-2 py-1 text-xs leading-snug text-foreground shadow-lg'
          style={{ left: tooltip.x, top: tooltip.y }}
        >
          {tooltip.text}
        </div>
      ) : null}

      {showFooter ? (
        <div className='mt-2 flex shrink-0 flex-col items-center gap-1'>
          <p className='text-xs uppercase tracking-wide text-muted'>
            {t('SnakeFocusRange', { start: footerStart, end: footerEnd })}
            {hasIdleHead ? ` · ${t('SnakeIdleFrozen')}` : null}
          </p>
          <p className='flex flex-wrap items-center justify-center gap-x-3 gap-y-0.5 text-[10px] text-muted'>
            <span className='inline-flex items-center gap-1'>
              <span className='size-2 rounded-[2px] bg-accent' aria-hidden />
              {t('SnakeLegendCached')}
            </span>
            <span className='inline-flex items-center gap-1'>
              <span className='size-2 rounded-[2px] border-2 border-foreground bg-transparent' aria-hidden />
              {t('SnakeLegendHead')}
            </span>
            <span className='inline-flex items-center gap-1'>
              <span
                className='size-2 rounded-[2px] border border-[#c4a882] bg-[#c4a882]/30 dark:border-[#c4a882]'
                aria-hidden
              />
              {t('SnakeLegendRange')}
            </span>
          </p>
        </div>
      ) : null}
    </div>
  )
}

export default memo(TorrentCache, (prev, next) => {
  if (prev.mode !== next.mode) return false
  if (prev.isSnakeDebugMode !== next.isSnakeDebugMode) return false
  if (prev.hash !== next.hash) return false
  const a = prev.cache
  const b = next.cache
  return (
    a.PiecesCount === b.PiecesCount &&
    a.PiecesLength === b.PiecesLength &&
    a.Capacity === b.Capacity &&
    a.Filled === b.Filled &&
    cheapPiecesFingerprint(a.Pieces) === cheapPiecesFingerprint(b.Pieces) &&
    cheapReadersFingerprint(a.Readers) === cheapReadersFingerprint(b.Readers)
  )
})
