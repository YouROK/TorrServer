import { rgba } from 'polished'
import { mainColors } from 'style/colors'

export const snakeSettings = {
  dark: {
    default: {
      borderWidth: 1,
      pieceSize: 14,
      gapBetweenPieces: 3,
      borderColor: rgba('#fff', 0.2),
      completeColor: rgba(mainColors.dark.primary, 0.5),
      backgroundColor: '#949ca0',
      progressColor: rgba('#fff', 0.2),
      readerColor: '#8f0405',
      rangeColor: '#cda184',
    },
    mini: {
      cacheMaxHeight: 340,
      borderWidth: 2,
      pieceSize: 23,
      gapBetweenPieces: 6,
      borderColor: '#5c6469',
      completeColor: '#5c6469',
      backgroundColor: '#949ca0',
      progressColor: '#949ca0',
      readerColor: '#ccc',
      rangeColor: '#cda184',
    },
  },
  light: {
    default: {
      borderWidth: 1,
      pieceSize: 14,
      gapBetweenPieces: 3,
      borderColor: '#dbf2e8',
      completeColor: mainColors.light.primary,
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
      readerColor: '#0a0a0a',
      rangeColor: '#afa6e3',
    },
  },
}

export const createGradient = (ctx, percentage, theme, snakeType) => {
  const { pieceSize, completeColor, progressColor } = snakeSettings[theme][snakeType]

  const gradient = ctx.createLinearGradient(0, pieceSize, 0, 0)
  gradient.addColorStop(0, completeColor)
  gradient.addColorStop(percentage / 100, completeColor)
  gradient.addColorStop(percentage / 100, progressColor)
  gradient.addColorStop(1, progressColor)

  return gradient
}
