import { useEffect, useRef, useState } from 'react'
import Typography from '@material-ui/core/Typography'
import DialogTitle from '@material-ui/core/DialogTitle'
import DialogContent from '@material-ui/core/DialogContent'
import { getPeerString, humanizeSize } from 'utils/Utils'
import { cacheHost } from 'utils/Hosts'
import { Stage, Layer } from 'react-konva'
import Measure from 'react-measure'

import SingleBlock from './SingleBlock'

export default function DialogCacheInfo({ hash }) {
  const [cache, setCache] = useState({})
  const [pMap, setPMap] = useState([])
  const timerID = useRef(null)
  const componentIsMounted = useRef(true)
  const [dimensions, setDimensions] = useState({ width: -1, height: -1 })
  const [isShortView, setIsShortView] = useState(true)
  const [isLoading, setIsLoading] = useState(true)
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
    updateStageSettings(24, 4)

    return () => {
      // this function is required to notify "getCache" when NOT to make state update
      componentIsMounted.current = false
    }
  }, [])

  useEffect(() => {
    if (hash) {
      timerID.current = setInterval(() => {
        getCache(hash, value => {
          // this is required to avoid memory leak
          if (componentIsMounted.current) setCache(value)
        })
      }, 100)
    } else clearInterval(timerID.current)

    return () => {
      clearInterval(timerID.current)
    }
  }, [hash])

  useEffect(() => {
    if (!cache.PiecesCount || !cache.Pieces) return

    const { Pieces, PiecesCount, Readers } = cache

    const map = []

    for (let i = 0; i < PiecesCount; i++) {
      const newPiece = { id: i }

      const currentPiece = Pieces[i]
      if (currentPiece) {
        if (currentPiece.Completed && currentPiece.Size === currentPiece.Length) newPiece.isComplete = true
        else {
          newPiece.inProgress = true
          newPiece.percentage = (currentPiece.Size / currentPiece.Length).toFixed(2)
        }
      }

      Readers.forEach(r => {
        if (i === r.Reader) newPiece.isActive = true
        if (i >= r.Start && i <= r.End) newPiece.isReaderRange = true
      })

      map.push(newPiece)
    }

    setPMap(map)
    setIsLoading(false)
  }, [cache])

  const { boxHeight, strokeWidth, marginBetweenBlocks, stageOffset } = stageSettings

  const preloadPiecesAmount = Math.round(cache.Capacity / cache.PiecesLength - 1)
  const blockSizeWithMargin = boxHeight + strokeWidth + marginBetweenBlocks
  const piecesInOneRow = Math.floor((dimensions.width * 0.9) / blockSizeWithMargin)
  const amountOfBlocksToRenderInShortView =
    preloadPiecesAmount === piecesInOneRow
      ? preloadPiecesAmount - 1
      : preloadPiecesAmount + piecesInOneRow - (preloadPiecesAmount % piecesInOneRow) - 1
  const amountOfRows = Math.ceil((isShortView ? amountOfBlocksToRenderInShortView : pMap.length) / piecesInOneRow)
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
                  {pMap.map(({ id, percentage, isComplete, inProgress, isActive, isReaderRange }) => {
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

const getCache = (hash, callback) => {
  try {
    fetch(cacheHost(), {
      method: 'post',
      body: JSON.stringify({ action: 'get', hash }),
      headers: {
        Accept: 'application/json, text/plain, */*',
        'Content-Type': 'application/json',
      },
    })
      .then(res => res.json())
      .then(callback, error => {
        callback({})
        console.error(error)
      })
  } catch (e) {
    console.error(e)
    callback({})
  }
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
