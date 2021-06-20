import { themeColors } from 'style/colors'

export const snakeSettings = {
  default: {
    borderWidth: 1,
    pieceSize: 14,
    gapBetweenPieces: 3,
    borderColor: '#dbf2e8',
    completeColor: themeColors.primary,
    backgroundColor: '#fff',
    progressColor: '#b3dfc9',
    readerColor: '#000',
    rangeColor: '#afa6e3',
  },
  mini: {
    cacheMaxHeight: 340,
    borderWidth: 2,
    pieceSize: 23,
    gapBetweenPieces: 6,
    borderColor: '#4db380',
    completeColor: '#4db380',
    backgroundColor: '#dbf2e8',
    progressColor: '#dbf2e8',
    readerColor: '#2d714f',
    rangeColor: '#afa6e3',
  },
}

export const createGradient = (ctx, percentage, snakeType) => {
  const { pieceSize, completeColor, progressColor } = snakeSettings[snakeType]

  const gradient = ctx.createLinearGradient(0, pieceSize, 0, 0)
  gradient.addColorStop(0, completeColor)
  gradient.addColorStop(percentage / 100, completeColor)
  gradient.addColorStop(percentage / 100, progressColor)
  gradient.addColorStop(1, progressColor)

  return gradient
}
