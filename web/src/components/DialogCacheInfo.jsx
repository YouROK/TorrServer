import { useEffect, useRef, useState } from 'react'
import Typography from '@material-ui/core/Typography'
import DialogTitle from '@material-ui/core/DialogTitle'
import DialogContent from '@material-ui/core/DialogContent'
import { getPeerString, humanizeSize } from 'utils/Utils'
import { cacheHost } from 'utils/Hosts'
import styled, { css } from 'styled-components'

const boxHeight = 12

const CacheWrapper = styled.div`
  padding-left: 6px;
  padding-right: 2px;
  line-height: 11px;

  .piece {
    width: ${boxHeight}px;
    height: ${boxHeight}px;
    background-color: #eef2f4;
    border: 1px solid #eef2f4;
    display: inline-block;
    margin-right: 1px;
  }
  .piece-complete {
    background-color: #3fb57a;
    border-color: #3fb57a;
  }
  .piece-loading {
    background-color: #00d0d0;
    border-color: #00d0d0;
  }
  .reader-range {
    border-color: #9a9aff;
  }
  .piece-reader {
    border-color: #000000;
  }
`

const PieceInProgress = styled.div`
  ${({ prc }) => css`
    position: relative;
    z-index: 1;
    background-color: #3fb57a;

    top: -1px;
    left: -1px;
    width: 12px;
    height: ${prc * boxHeight}px;
  `}
`

export default function DialogCacheInfo({ hash }) {
  const [cache, setCache] = useState({})
  const [pMap, setPMap] = useState([])
  const timerID = useRef(null)
  const componentIsMounted = useRef(true)

  useEffect(
    // this function is required to notify "getCache" when NOT to make state update
    () => () => {
      componentIsMounted.current = false
    },
    [],
  )

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
    if (!cache?.PiecesCount || !cache?.Pieces) return

    const { Pieces, PiecesCount, Readers } = cache

    const map = []

    for (let i = 0; i < PiecesCount; i++) {
      const cls = ['piece']
      let prc = 0

      const currentPiece = Pieces[i]
      if (currentPiece) {
        if (currentPiece.Completed && currentPiece.Size === currentPiece.Length) cls.push('piece-complete')
        else cls.push('piece-loading')

        prc = (currentPiece.Size / currentPiece.Length).toFixed(2)
      }

      Readers.forEach(r => {
        if (i === r.Reader) return cls.push('piece-reader')
        if (i >= r.Start && i <= r.End) cls.push('reader-range')
      })

      map.push({ prc, className: cls.join(' '), id: i })
    }

    setPMap(map)
  }, [cache])

  return (
    <div>
      <DialogTitle id='form-dialog-title'>
        <Typography>
          <b>Hash </b> {cache.Hash}
          <br />
          <b>Capacity </b> {humanizeSize(cache.Capacity)}
          <br />
          <b>Filled </b> {humanizeSize(cache.Filled)}
          <br />
          <b>Torrent size </b> {cache.Torrent && cache.Torrent.torrent_size && humanizeSize(cache.Torrent.torrent_size)}
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
        <CacheWrapper>
          {pMap.map(({ prc, className: currentPieceCalss, id }) => (
            <span key={id} className={currentPieceCalss}>
              {prc > 0 && prc < 1 && <PieceInProgress prc={prc} />}
            </span>
          ))}
        </CacheWrapper>
      </DialogContent>
    </div>
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
