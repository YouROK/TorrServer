import { useEffect, useState } from 'react'
import DialogContent from '@material-ui/core/DialogContent'
import { Stage, Layer } from 'react-konva'
import Measure from 'react-measure'
import { v4 as uuidv4 } from 'uuid'
import styled from 'styled-components'
import { useTranslation } from 'react-i18next'

import SingleBlock from './SingleBlock'
import getShortCacheMap from './getShortCacheMap'

const ScrollNotification = styled.div`
  margin-top: 10px;
  text-transform: uppercase;
  color: rgba(0, 0, 0, 0.5);
  align-self: center;
`

export default function DefaultSnake({ isMini, cacheMap, preloadPiecesAmount }) {
  const { t } = useTranslation()
  const [dimensions, setDimensions] = useState({ width: 0, height: 0 })
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
    if (isMini) return dimensions.width < 500 ? updateStageSettings(20, 3) : updateStageSettings(24, 4)
    updateStageSettings(12, 2)
  }, [isMini, dimensions.width])

  const miniCacheMaxHeight = 340

  const { boxHeight, strokeWidth, marginBetweenBlocks, stageOffset } = stageSettings

  const blockSizeWithMargin = boxHeight + strokeWidth + marginBetweenBlocks
  const piecesInOneRow = Math.floor((dimensions.width * 0.9) / blockSizeWithMargin)

  const shortCacheMap = isMini ? getShortCacheMap({ cacheMap, preloadPiecesAmount, piecesInOneRow }) : []

  const amountOfRows = Math.ceil((isMini ? shortCacheMap.length : cacheMap.length) / piecesInOneRow)

  const getItemCoordinates = blockOrder => {
    const currentRow = Math.floor(blockOrder / piecesInOneRow)
    const x = (blockOrder % piecesInOneRow) * blockSizeWithMargin || 0
    const y = currentRow * blockSizeWithMargin || 0

    return { x, y }
  }

  return (
    <Measure bounds onResize={({ bounds }) => setDimensions(bounds)}>
      {({ measureRef }) => (
        <div style={{ display: 'flex', flexDirection: 'column' }}>
          <DialogContent
            ref={measureRef}
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
                      const { x, y } = getItemCoordinates(i)

                      return (
                        <SingleBlock
                          key={uuidv4()}
                          x={x}
                          y={y}
                          percentage={percentage}
                          inProgress={inProgress}
                          isComplete={isComplete}
                          isReaderRange={isReaderRange}
                          isActive={isActive}
                          boxHeight={boxHeight}
                          strokeWidth={strokeWidth}
                        />
                      )
                    })
                  : cacheMap.map(({ id, percentage, isComplete, inProgress, isActive, isReaderRange }) => {
                      const { x, y } = getItemCoordinates(id)

                      return (
                        <SingleBlock
                          key={uuidv4()}
                          x={x}
                          y={y}
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

          {isMini &&
            (stageOffset + blockSizeWithMargin * amountOfRows || 0) >= miniCacheMaxHeight &&
            dimensions.height >= miniCacheMaxHeight && <ScrollNotification>{t('ScrollDown')}</ScrollNotification>}
        </div>
      )}
    </Measure>
  )
}
