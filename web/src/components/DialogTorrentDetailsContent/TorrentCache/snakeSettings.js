import { cacheBackground } from '../style'

export const defaultBorderWidth = 1
export const miniBorderWidth = 2
export const defaultPieceSize = 14
export const pieceSizeForMiniMap = 23
export const defaultGapBetweenPieces = 3
export const miniGapBetweenPieces = 6
export const miniCacheMaxHeight = 340

export const defaultBorderColor = '#dbf2e8'
export const defaultBackgroundColor = '#fff'
export const miniBackgroundColor = cacheBackground
export const completeColor = '#00a572'
export const progressColor = '#86beee'
export const activeColor = '#000'
export const rangeColor = '#afa6e3'

export const createGradient = (ctx, percentage) => {
  const gradient = ctx.createLinearGradient(0, 12, 0, 0)
  gradient.addColorStop(0, completeColor)
  gradient.addColorStop(percentage / 100, completeColor)
  gradient.addColorStop(percentage / 100, progressColor)
  gradient.addColorStop(1, progressColor)

  return gradient
}
