import { memo, useEffect, useState } from 'react'
import DialogContent from '@material-ui/core/DialogContent'
import { Stage, Layer } from 'react-konva'
import Measure from 'react-measure'
import isEqual from 'lodash/isEqual'
import styled from 'styled-components'
import { v4 as uuidv4 } from 'uuid'

import SingleBlock from './SingleBlock'
import { useCreateCacheMap } from './customHooks'

const ScrollNotification = styled.div`
  margin-top: 10px;
  text-transform: uppercase;
  color: rgba(0, 0, 0, 0.5);
`

const TorrentCache = memo(
  ({ cache, isMini }) => {
    const [dimensions, setDimensions] = useState({ width: 0, height: 0 })
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

    const miniCacheMaxHeight = 340

    const { boxHeight, strokeWidth, marginBetweenBlocks, stageOffset } = stageSettings
    const preloadPiecesAmount = Math.round(cache.Capacity / cache.PiecesLength - 1)
    const blockSizeWithMargin = boxHeight + strokeWidth + marginBetweenBlocks
    const piecesInOneRow = Math.floor((dimensions.width * 0.9) / blockSizeWithMargin)
    const amountOfBlocksToRenderInShortView =
      preloadPiecesAmount === piecesInOneRow
        ? preloadPiecesAmount - 1
        : preloadPiecesAmount + piecesInOneRow - (preloadPiecesAmount % piecesInOneRow) - 1 || 0
    const amountOfRows = Math.ceil((isMini ? amountOfBlocksToRenderInShortView : cacheMap.length) / piecesInOneRow)
    const activeId = null

    const cacheMapWithoutEmptyBlocks = cacheMap.filter(({ isComplete, inProgress }) => inProgress || isComplete)
    const extraEmptyBlocksForFillingLine =
      cacheMapWithoutEmptyBlocks.length < amountOfBlocksToRenderInShortView
        ? new Array(amountOfBlocksToRenderInShortView - cacheMapWithoutEmptyBlocks.length + 1).fill({})
        : []
    const shortCacheMap = [...cacheMapWithoutEmptyBlocks, ...extraEmptyBlocksForFillingLine]

    return (
      <Measure bounds onResize={({ bounds }) => setDimensions(bounds)}>
        {({ measureRef }) => (
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }} ref={measureRef}>
            <DialogContent
              {...(isMini
                ? { style: { padding: 0, maxHeight: `${miniCacheMaxHeight}px`, overflow: 'auto' } }
                : { style: { padding: 0 } })}
            >
              <Stage
                style={{ display: 'flex', justifyContent: 'center' }}
                offset={{ x: -stageOffset, y: -stageOffset }}
                width={stageOffset + blockSizeWithMargin * piecesInOneRow || 0}
                height={stageOffset + blockSizeWithMargin * amountOfRows || 0}
              >
                <Layer>
                  {isMini
                    ? shortCacheMap.map(({ percentage, isComplete, inProgress, isActive, isReaderRange }, i) => {
                        const currentRow = Math.floor(i / piecesInOneRow)
                        const shouldBeRendered = inProgress || isComplete || i <= amountOfBlocksToRenderInShortView

                        return (
                          shouldBeRendered && (
                            <SingleBlock
                              key={uuidv4()}
                              x={(i % piecesInOneRow) * blockSizeWithMargin}
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
                        )
                      })
                    : cacheMap.map(({ id, percentage, isComplete, inProgress, isActive, isReaderRange }) => {
                        const currentRow = Math.floor((isMini ? id - activeId : id) / piecesInOneRow)

                        return (
                          <SingleBlock
                            key={uuidv4()}
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

            {dimensions.height >= miniCacheMaxHeight && <ScrollNotification>scroll down</ScrollNotification>}
          </div>
        )}
      </Measure>
    )
  },
  (prev, next) => isEqual(prev.cache.Pieces, next.cache.Pieces) && isEqual(prev.cache.Readers, next.cache.Readers),
)

export default TorrentCache
