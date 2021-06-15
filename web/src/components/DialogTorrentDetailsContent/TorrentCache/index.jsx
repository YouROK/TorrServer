import Measure from 'react-measure'
import { useState, memo } from 'react'
import { v4 as uuidv4 } from 'uuid'
import { useTranslation } from 'react-i18next'
import isEqual from 'lodash/isEqual'

import { useCreateCacheMap } from '../customHooks'
import { gapBetweenPieces, miniCacheMaxHeight, pieceSizeForMiniMap, defaultPieceSize } from './snakeSettings'
import getShortCacheMap from './getShortCacheMap'
import { SnakeWrapper, PercentagePiece, ScrollNotification } from './style'

const TorrentCache = ({ cache, isMini }) => {
  const { t } = useTranslation()
  const [dimensions, setDimensions] = useState({ width: 0, height: 0 })
  const cacheMap = useCreateCacheMap(cache)

  const preloadPiecesAmount = Math.round(cache.Capacity / cache.PiecesLength - 1)

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

export default memo(
  TorrentCache,
  (prev, next) => isEqual(prev.cache.Pieces, next.cache.Pieces) && isEqual(prev.cache.Readers, next.cache.Readers),
)
