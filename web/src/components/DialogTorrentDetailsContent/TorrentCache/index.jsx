import { memo } from 'react'
import isEqual from 'lodash/isEqual'

import { useCreateCacheMap } from '../customHooks'
import LargeSnake from './LargeSnake'
import DefaultSnake from './DefaultSnake'

const TorrentCache = memo(
  ({ cache, isMini }) => {
    const cacheMap = useCreateCacheMap(cache)

    const preloadPiecesAmount = Math.round(cache.Capacity / cache.PiecesLength - 1)
    const isSnakeLarge = cacheMap.length > 7000

    return isMini ? (
      <DefaultSnake isMini cacheMap={cacheMap} preloadPiecesAmount={preloadPiecesAmount} />
    ) : isSnakeLarge ? (
      <LargeSnake cacheMap={cacheMap} />
    ) : (
      <DefaultSnake cacheMap={cacheMap} preloadPiecesAmount={preloadPiecesAmount} />
    )
  },
  (prev, next) => isEqual(prev.cache.Pieces, next.cache.Pieces) && isEqual(prev.cache.Readers, next.cache.Readers),
)

export default TorrentCache
