import { memo, useEffect, useState } from 'react'
import DialogContent from '@material-ui/core/DialogContent'
import { Stage, Layer } from 'react-konva'
import Measure from 'react-measure'
import { isEqual } from 'lodash'

import SingleBlock from './SingleBlock'
import { useCreateCacheMap } from './customHooks'

const TorrentCache = memo(
  ({ cache, isMini }) => {
    const [dimensions, setDimensions] = useState({ width: -1, height: -1 })
    const [stageSettings, setStageSettings] = useState({
      boxHeight: null,
      strokeWidth: null,
      marginBetweenBlocks: null,
      stageOffset: null,
    })

    const cacheMap = useCreateCacheMap(cache)

    const updateStageSettings = (boxHeight, strokeWidth) => {
      setStageSettings({
        boxHeight,
        strokeWidth,
        marginBetweenBlocks: strokeWidth,
        stageOffset: strokeWidth * 2,
      })
    }

    useEffect(() => {
      // initializing stageSettings
      if (isMini) return dimensions.width < 500 ? updateStageSettings(20, 3) : updateStageSettings(24, 4)
      updateStageSettings(12, 2)
    }, [isMini, dimensions.width])

    const { boxHeight, strokeWidth, marginBetweenBlocks, stageOffset } = stageSettings
    const preloadPiecesAmount = Math.round(cache.Capacity / cache.PiecesLength - 1)
    const blockSizeWithMargin = boxHeight + strokeWidth + marginBetweenBlocks
    const piecesInOneRow = Math.floor((dimensions.width * 0.9) / blockSizeWithMargin)
    const amountOfBlocksToRenderInShortView =
      preloadPiecesAmount === piecesInOneRow
        ? preloadPiecesAmount - 1
        : preloadPiecesAmount + piecesInOneRow - (preloadPiecesAmount % piecesInOneRow) - 1
    const amountOfRows = Math.ceil((isMini ? amountOfBlocksToRenderInShortView : cacheMap.length) / piecesInOneRow)
    let activeId = null

    return (
      <Measure bounds onResize={({ bounds }) => setDimensions(bounds)}>
        {({ measureRef }) => (
          <div ref={measureRef}>
            <DialogContent style={{ padding: 0 }}>
              <Stage
                style={{ display: 'flex', justifyContent: 'center' }}
                offset={{ x: -stageOffset, y: -stageOffset }}
                width={stageOffset + blockSizeWithMargin * piecesInOneRow || 0}
                height={stageOffset + blockSizeWithMargin * amountOfRows || 0}
              >
                <Layer>
                  {cacheMap.map(({ id, percentage, isComplete, inProgress, isActive, isReaderRange }) => {
                    const currentRow = Math.floor((isMini ? id - activeId : id) / piecesInOneRow)

                    // -------- related only for short view -------
                    if (isActive) activeId = id
                    const shouldBeRendered =
                      isActive || (id - activeId <= amountOfBlocksToRenderInShortView && id - activeId >= 0)
                    // --------------------------------------------

                    return isMini ? (
                      shouldBeRendered && (
                        <SingleBlock
                          key={id}
                          x={((id - activeId) % piecesInOneRow) * blockSizeWithMargin}
                          y={currentRow * blockSizeWithMargin}
                          percentage={percentage}
                          inProgress={inProgress}
                          isComplete={isComplete}
                          isReaderRange={isReaderRange}
                          isActive={isActive}
                          boxHeight={boxHeight}
                          strokeWidth={strokeWidth}
                        />
                      )
                    ) : (
                      <SingleBlock
                        key={id}
                        x={(id % piecesInOneRow) * blockSizeWithMargin}
                        y={currentRow * blockSizeWithMargin}
                        percentage={percentage}
                        inProgress={inProgress}
                        isComplete={isComplete}
                        isReaderRange={isReaderRange}
                        isActive={isActive}
                        boxHeight={boxHeight}
                        strokeWidth={strokeWidth}
                      />
                    )
                  })}
                </Layer>
              </Stage>
            </DialogContent>
          </div>
        )}
      </Measure>
    )
  },
  (prev, next) => isEqual(prev.cache.Pieces, next.cache.Pieces) && isEqual(prev.cache.Readers, next.cache.Readers),
)

export default TorrentCache
