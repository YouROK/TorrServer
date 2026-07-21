import type { CacheMapItem } from 'shared/api/types'

import { priorityDebugLabel } from './buildCacheMap'
import { resolveSnakeSettings, type SnakeThemeMode, type SnakeVariant } from './snakeSettings'

export interface DrawSnakeArgs {
  ctx: CanvasRenderingContext2D
  cells: CacheMapItem[]
  canvasWidth: number
  canvasHeight: number
  piecesInOneRow: number
  pieceSize: number
  gap: number
  startingX: number
  theme: SnakeThemeMode
  variant: SnakeVariant
  isSnakeDebugMode?: boolean
  isMini?: boolean
}

const clamp01 = (n: number) => Math.max(0, Math.min(1, n))

const fillProgress = (
  ctx: CanvasRenderingContext2D,
  size: number,
  percentage: number,
  emptyColor: string,
  fillColor: string,
) => {
  ctx.fillStyle = emptyColor
  ctx.fillRect(0, 0, size, size)

  const ratio = clamp01(percentage / 100)
  if (ratio <= 0) return

  if (ratio >= 1) {
    ctx.fillStyle = fillColor
    ctx.fillRect(0, 0, size, size)
    return
  }

  const filledH = Math.max(2, Math.round(size * ratio))
  ctx.fillStyle = fillColor
  ctx.fillRect(0, size - filledH, size, Math.min(filledH, size))
}

const strokeCell = (ctx: CanvasRenderingContext2D, size: number, color: string, line: number) => {
  const inset = line / 2
  ctx.lineWidth = line
  ctx.strokeStyle = color
  ctx.strokeRect(inset, inset, size - line, size - line)
}

/**
 * Hierarchy: reader (black outline) > range (violet/sand) > cached (green) > idle.
 */
export const drawSnake = ({
  ctx,
  cells,
  canvasWidth,
  canvasHeight,
  piecesInOneRow,
  pieceSize,
  gap,
  startingX,
  theme,
  variant,
  isSnakeDebugMode,
  isMini,
}: DrawSnakeArgs) => {
  const settings = resolveSnakeSettings(theme, variant)
  const {
    borderWidth,
    backgroundColor,
    borderColor,
    completeColor,
    readerColor,
    readerHaloColor,
    rangeColor,
    rangeEmptyColor,
  } = settings

  ctx.clearRect(0, 0, canvasWidth, canvasHeight)
  ctx.imageSmoothingEnabled = false

  const pixelAlign = borderWidth % 2 === 1 ? 0.5 : 0
  const isDark = theme === 'dark'
  const rangeIdle = rangeEmptyColor || (isDark ? 'rgba(205, 161, 132, 0.28)' : 'rgba(175, 166, 227, 0.32)')

  for (let i = 0; i < cells.length; i++) {
    const cell = cells[i] || { percentage: 0, priority: 0 }
    const percentage = cell.percentage || 0
    const isCompleted = Boolean(cell.completed) || percentage >= 100
    const inProgress = percentage > 0 && !isCompleted
    const isReader = Boolean(cell.isReader)
    const isReaderRange = Boolean(cell.isReaderRange)

    const col = i % piecesInOneRow
    const row = Math.floor(i / piecesInOneRow)
    const x = startingX + col * (pieceSize + gap) + pixelAlign
    const y = row * (pieceSize + gap) + pixelAlign

    ctx.save()
    ctx.translate(x, y)

    const emptyColor = isReaderRange && !isCompleted && !inProgress ? rangeIdle : backgroundColor
    fillProgress(
      ctx,
      pieceSize,
      isCompleted ? 100 : percentage,
      emptyColor,
      inProgress || isCompleted ? completeColor : emptyColor,
    )

    if (isReader) {
      // Master-style square outline: optional light halo + black (or theme) stroke
      if (readerHaloColor) strokeCell(ctx, pieceSize, readerHaloColor, isMini ? 4 : 3)
      strokeCell(ctx, pieceSize, readerColor, isMini ? 2.5 : 2)
    } else if (isReaderRange) {
      strokeCell(ctx, pieceSize, rangeColor, 2)
    } else if (inProgress || isCompleted) {
      strokeCell(ctx, pieceSize, completeColor, Math.max(borderWidth, 2))
    } else {
      strokeCell(ctx, pieceSize, borderColor, borderWidth)
    }

    if (isSnakeDebugMode) {
      const info = priorityDebugLabel(cell.priority || 0)
      if (info) {
        const fontSize = Math.max(9, Math.min(isMini ? 13 : 11, Math.floor(pieceSize * 0.55)))
        ctx.font = `bold ${fontSize}px ui-monospace, SFMono-Regular, Menlo, monospace`
        ctx.textAlign = 'center'
        ctx.textBaseline = 'middle'
        const cx = pieceSize / 2
        const cy = pieceSize / 2
        ctx.lineWidth = 3
        ctx.strokeStyle = isDark ? 'rgba(0,0,0,0.85)' : 'rgba(255,255,255,0.95)'
        ctx.strokeText(info, cx, cy)
        ctx.fillStyle = isDark ? '#fff' : '#1a1a1a'
        ctx.fillText(info, cx, cy)
      }
    }

    ctx.restore()
  }
}

export const setupHiDpiCanvas = (
  canvas: HTMLCanvasElement,
  cssWidth: number,
  cssHeight: number,
): CanvasRenderingContext2D | null => {
  const dpr = Math.min(window.devicePixelRatio || 1, 2)
  canvas.width = Math.max(1, Math.round(cssWidth * dpr))
  canvas.height = Math.max(1, Math.round(cssHeight * dpr))
  canvas.style.width = `${cssWidth}px`
  canvas.style.height = `${cssHeight}px`
  const ctx = canvas.getContext('2d')
  if (!ctx) return null
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
  return ctx
}

export interface HitTestArgs {
  piecesInOneRow: number
  pieceSize: number
  gap: number
  startingX: number
  cellCount: number
}

export const hitTestSnakeCell = (x: number, y: number, args: HitTestArgs): number => {
  const { piecesInOneRow, pieceSize, gap, startingX, cellCount } = args
  if (piecesInOneRow < 1 || cellCount < 1 || pieceSize <= 0) return -1
  const stride = pieceSize + gap
  const localX = x - startingX
  if (localX < 0 || y < 0) return -1
  const col = Math.floor(localX / stride)
  const row = Math.floor(y / stride)
  if (col < 0 || col >= piecesInOneRow) return -1
  const inCellX = localX - col * stride
  const inCellY = y - row * stride
  if (inCellX > pieceSize || inCellY > pieceSize) return -1
  const index = row * piecesInOneRow + col
  if (index < 0 || index >= cellCount) return -1
  return index
}
