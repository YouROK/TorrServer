import { rgba } from 'polished'
import { mainColors } from 'style/colors'

export const snakeSettings = {
  dark: {
    default: {
      borderWidth: 2,
      pieceSize: 14,
      gapBetweenPieces: 3,
      borderColor: mainColors.dark.secondary,
      completeColor: rgba(mainColors.dark.primary, 0.65),
      backgroundColor: '#f1eff3',
      progressColor: mainColors.dark.secondary,
      readerColor: '#000',
      rangeColor: '#cda184',
    },
    mini: {
      cacheMaxHeight: 340,
      borderWidth: 2,
      pieceSize: 23,
      gapBetweenPieces: 6,
      borderColor: '#545a5e',
      completeColor: '#545a5e',
      backgroundColor: '#dee3e5',
      progressColor: '#dee3e5',
      readerColor: '#000',
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
      readerColor: '#2d714f',
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
