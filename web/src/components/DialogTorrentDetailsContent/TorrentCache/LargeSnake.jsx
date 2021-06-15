import styled, { css } from 'styled-components'
import Measure from 'react-measure'
import { useState } from 'react'
import { v4 as uuidv4 } from 'uuid'
import { useTranslation } from 'react-i18next'

import {
  defaultBackgroundColor,
  defaultBorderColor,
  progressColor,
  completeColor,
  activeColor,
  rangeColor,
} from './colors'
import getShortCacheMap from './getShortCacheMap'

const borderWidth = 1
const defaultPieceSize = 14
const pieceSizeForMiniMap = 23
const gapBetweenPieces = 3
const miniCacheMaxHeight = 340

const ScrollNotification = styled.div`
  margin-top: 10px;
  text-transform: uppercase;
  color: rgba(0, 0, 0, 0.5);
  align-self: center;
`

const SnakeWrapper = styled.div`
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

const PercentagePiece = styled.div`
  background: ${completeColor};
  height: ${({ percentage }) => (percentage / 100) * 12}px;
`

export default function LargeSnake({ cacheMap, isMini, preloadPiecesAmount }) {
  const { t } = useTranslation()
  const [dimensions, setDimensions] = useState({ width: 0, height: 0 })
  const pieceSize = isMini ? pieceSizeForMiniMap : defaultPieceSize

  let piecesInOneRow
  let shotCacheMap
  if (isMini) {
    const pieceSizeWithGap = pieceSize + gapBetweenPieces
    piecesInOneRow = Math.floor((dimensions.width * 0.95) / pieceSizeWithGap)
    shotCacheMap = isMini && getShortCacheMap({ cacheMap, preloadPiecesAmount, piecesInOneRow })
  }

  return isMini ? (
    <Measure bounds onResize={({ bounds }) => setDimensions(bounds)}>
      {({ measureRef }) => (
        <div style={{ display: 'flex', flexDirection: 'column' }}>
          <SnakeWrapper ref={measureRef} pieceSize={pieceSize} piecesInOneRow={piecesInOneRow}>
            {shotCacheMap.map(({ className, id, percentage }) => (
              <span key={id || uuidv4()} className={className}>
                {percentage > 0 && percentage <= 100 && <PercentagePiece percentage={percentage} />}
              </span>
            ))}
          </SnakeWrapper>

          {dimensions.height >= miniCacheMaxHeight && <ScrollNotification>{t('ScrollDown')}</ScrollNotification>}
        </div>
      )}
    </Measure>
  ) : (
    <SnakeWrapper pieceSize={pieceSize}>
      {cacheMap.map(({ className, id, percentage }) => (
        <span key={id || uuidv4()} className={className}>
          {percentage > 0 && percentage <= 100 && <PercentagePiece percentage={percentage} />}
        </span>
      ))}
    </SnakeWrapper>
  )
}
