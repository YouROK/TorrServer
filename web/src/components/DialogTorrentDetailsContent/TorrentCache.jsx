import { useEffect, useState } from 'react'
import DialogContent from '@material-ui/core/DialogContent'
import { Stage, Layer } from 'react-konva'
import Measure from 'react-measure'

import SingleBlock from './SingleBlock'

export default function TorrentCache({ cache, cacheMap, isMini }) {
  const [dimensions, setDimensions] = useState({ width: -1, height: -1 })
  const [stageSettings, setStageSettings] = useState({
    boxHeight: null,
    strokeWidth: null,
    marginBetweenBlocks: null,
    stageOffset: null,
  })

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
    isMini ? updateStageSettings(24, 4) : updateStageSettings(12, 2)
  }, [isMini])

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
    <Measure bounds onResize={contentRect => setDimensions(contentRect.bounds)}>
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
}
