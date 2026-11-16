// Browsers refuse to allocate a canvas that is too big, and the limit differs between browsers, platforms and hardware
// Firefox puts the 2d context into an error state and throws on every draw call (which unmounts the whole app)
// Chromium shows a "sad face" instead of the canvas.
// TorrentCache divides this budget by `shrink` every time the browser rejects the canvas.
const MAX_CANVAS_SIDE = 32767
const MAX_CANVAS_AREA = 16 * 1024 * 1024

// Merges neighbouring pieces into one cell when the map would not fit into the allowed canvas size.
// A merged cell shows the average fill and keeps the reader markers and the highest priority of its pieces.
export default ({ source, piecesInOneRow, pieceSizeWithGap, canvasWidth, shrink = 1 }) => {
  const maxHeight = Math.min(MAX_CANVAS_SIDE, Math.floor(MAX_CANVAS_AREA / Math.max(canvasWidth, 1))) / shrink
  const maxCells = Math.floor(maxHeight / pieceSizeWithGap) * piecesInOneRow
  if (!(maxCells > 0) || source.length <= maxCells) return source

  const piecesPerCell = Math.ceil(source.length / maxCells)
  const merged = []

  for (let start = 0; start < source.length; start += piecesPerCell) {
    const end = Math.min(start + piecesPerCell, source.length)
    const cell = { percentage: 0, priority: 0 }

    for (let i = start; i < end; i++) {
      const { percentage = 0, priority = 0, isReader, isReaderRange } = source[i]
      cell.percentage += percentage
      cell.priority = Math.max(cell.priority, priority)
      if (isReader) cell.isReader = true
      if (isReaderRange) cell.isReaderRange = true
    }

    cell.percentage /= end - start
    merged.push(cell)
  }

  return merged
}
