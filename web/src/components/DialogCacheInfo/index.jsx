import { useEffect, useState } from 'react'
import Typography from '@material-ui/core/Typography'
import DialogTitle from '@material-ui/core/DialogTitle'
import DialogContent from '@material-ui/core/DialogContent'
import { getPeerString, humanizeSize } from 'utils/Utils'
import { Stage, Layer } from 'react-konva'
import Measure from 'react-measure'
import { useUpdateCache, useCreateCacheMap } from 'components/DialogTorrentDetailsContent/customHooks'

import SingleBlock from './SingleBlock'

export default function DialogCacheInfo({ hash }) {
  const [dimensions, setDimensions] = useState({ width: -1, height: -1 })
  const [isShortView, setIsShortView] = useState(true)
  const [isLoading, setIsLoading] = useState(true)
  const [stageSettings, setStageSettings] = useState({
    boxHeight: null,
    strokeWidth: null,
    marginBetweenBlocks: null,
    stageOffset: null,
  })

  const cache = useUpdateCache(hash)
  const cacheMap = useCreateCacheMap(cache, () => setIsLoading(false))

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
    updateStageSettings(24, 4)
  }, [])

  const { boxHeight, strokeWidth, marginBetweenBlocks, stageOffset } = stageSettings

  const preloadPiecesAmount = Math.round(cache.Capacity / cache.PiecesLength - 1)
  const blockSizeWithMargin = boxHeight + strokeWidth + marginBetweenBlocks
  const piecesInOneRow = Math.floor((dimensions.width * 0.9) / blockSizeWithMargin)
  const amountOfBlocksToRenderInShortView =
    preloadPiecesAmount === piecesInOneRow
      ? preloadPiecesAmount - 1
      : preloadPiecesAmount + piecesInOneRow - (preloadPiecesAmount % piecesInOneRow) - 1
  const amountOfRows = Math.ceil((isShortView ? amountOfBlocksToRenderInShortView : cacheMap.length) / piecesInOneRow)
  let activeId = null

  return (
    <Measure bounds onResize={contentRect => setDimensions(contentRect.bounds)}>
      {({ measureRef }) => (
        <div ref={measureRef}>
          <DialogTitle id='form-dialog-title'>
            <Typography>
              <b>Hash </b> <span style={{ wordBreak: 'break-word' }}>{cache.Hash}</span>
              <br />
              <b>Capacity </b> {humanizeSize(cache.Capacity)}
              <br />
              <b>Filled </b> {humanizeSize(cache.Filled)}
              <br />
              <b>Torrent size </b>{' '}
              {cache.Torrent && cache.Torrent.torrent_size && humanizeSize(cache.Torrent.torrent_size)}
              <br />
              <b>Pieces length </b> {humanizeSize(cache.PiecesLength)}
              <br />
              <b>Pieces count </b> {cache.PiecesCount}
              <br />
              <b>Peers: </b> {getPeerString(cache.Torrent)}
              <br />
              <b>Download speed </b>{' '}
              {cache.Torrent && cache.Torrent.download_speed ? `${humanizeSize(cache.Torrent.download_speed)}/sec` : ''}
              <br />
              <b>Upload speed </b>{' '}
              {cache.Torrent && cache.Torrent.upload_speed ? `${humanizeSize(cache.Torrent.upload_speed)}/sec` : ''}
              <br />
              <b>Status </b> {cache.Torrent && cache.Torrent.stat_string && cache.Torrent.stat_string}
            </Typography>
          </DialogTitle>

          <DialogContent>
            <button
              type='button'
              onClick={() => {
                if (isShortView) {
                  updateStageSettings(12, 2)
                  setIsShortView(false)
                } else {
                  updateStageSettings(24, 4)
                  setIsShortView(true)
                }
                setIsLoading(true)
              }}
            >
              updateStageSettings
            </button>
            {isLoading ? (
              'loading'
            ) : (
              <Stage
                style={{ display: 'flex', justifyContent: 'center' }}
                offset={{ x: -stageOffset, y: -stageOffset }}
                width={stageOffset + blockSizeWithMargin * piecesInOneRow}
                height={stageOffset + blockSizeWithMargin * amountOfRows}
              >
                <Layer>
                  {cacheMap.map(({ id, percentage, isComplete, inProgress, isActive, isReaderRange }) => {
                    const currentRow = Math.floor((isShortView ? id - activeId : id) / piecesInOneRow)

                    // -------- related only for short view -------
                    if (isActive) activeId = id
                    const shouldBeRendered =
                      isActive || (id - activeId <= amountOfBlocksToRenderInShortView && id - activeId >= 0)
                    // --------------------------------------------

                    return isShortView ? (
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
            )}
          </DialogContent>
        </div>
      )}
    </Measure>
  )
}

/*
{
    "Hash": "41e36c8de915d80db83fc134bee4e7e2d292657e",
    "Capacity": 209715200,
    "Filled": 2914808,
    "PiecesLength": 4194304,
    "PiecesCount": 2065,
    "DownloadSpeed": 32770.860273455524,
    "Pieces": {
        "2064": {
            "Id": 2064,
            "Length": 2914808,
            "Size": 162296,
            "Completed": false
        }
    }
}
 */
