import { memo } from 'react'
import isEqual from 'lodash/isEqual'

import { useCreateCacheMap } from '../customHooks'
import LargeSnake from './LargeSnake'

const TorrentCache = memo(
  ({ cache, isMini }) => {
    const cacheMap = useCreateCacheMap(cache)

    const preloadPiecesAmount = Math.round(cache.Capacity / cache.PiecesLength - 1)

    return <LargeSnake isMini={isMini} cacheMap={cacheMap} preloadPiecesAmount={preloadPiecesAmount} />
  },
  (prev, next) => isEqual(prev.cache.Pieces, next.cache.Pieces) && isEqual(prev.cache.Readers, next.cache.Readers),
)

export default TorrentCache
