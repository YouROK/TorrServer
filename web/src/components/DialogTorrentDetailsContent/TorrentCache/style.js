import styled, { css } from 'styled-components'

import {
  defaultBackgroundColor,
  defaultBorderColor,
  progressColor,
  completeColor,
  activeColor,
  rangeColor,
  gapBetweenPieces,
  miniCacheMaxHeight,
  borderWidth,
} from './snakeSettings'

export const ScrollNotification = styled.div`
  margin-top: 10px;
  text-transform: uppercase;
  color: rgba(0, 0, 0, 0.5);
  align-self: center;
`

export const SnakeWrapper = styled.div`
  ${({ pieceSize, piecesInOneRow }) => css`
    display: grid;
    gap: ${gapBetweenPieces}px;
    grid-template-columns: repeat(${piecesInOneRow || 'auto-fit'}, ${pieceSize}px);
    grid-auto-rows: max-content;
    justify-content: center;

    ${piecesInOneRow &&
    css`
      max-height: ${miniCacheMaxHeight}px;
      overflow: auto;
    `}

    .piece {
      width: ${pieceSize}px;
      height: ${pieceSize}px;
      background: ${defaultBackgroundColor};
      border: ${borderWidth}px solid ${defaultBorderColor};
      display: grid;
      align-items: end;

      &-loading {
        background: ${progressColor};
        border-color: ${progressColor};
      }
      &-complete {
        background: ${completeColor};
        border-color: ${completeColor};
      }
      &-reader {
        border-color: ${activeColor};
      }
    }

    .reader-range {
      border-color: ${rangeColor};
    }
  `}
`

export const PercentagePiece = styled.div`
  background: ${completeColor};
  height: ${({ percentage }) => percentage}%;
`
