import styled from 'styled-components'
import { useEffect, useRef, useState } from 'react'
import { Typography } from '@material-ui/core'
import { torrentsHost } from 'utils/Hosts'
import Torrent from 'components/Torrent'

const TorrentListWrapper = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, 350px);
  gap: 30px;

  @media (max-width: 600px), (max-height: 500px) {
    gap: 10px;
    grid-template-columns: repeat(auto-fit, 310px);
  }

  @media (max-width: 410px) {
    grid-template-columns: minmax(min-content, 290px);
  }
`

const getTorrentList = (callback, errorCallback) => {
  fetch(torrentsHost(), {
    method: 'post',
    body: JSON.stringify({ action: 'list' }),
    headers: {
      Accept: 'application/json, text/plain, */*',
      'Content-Type': 'application/json',
    },
  })
    .then(res => res.json())
    .then(callback)
    .catch(() => errorCallback())
}

export default function TorrentList() {
  const [torrents, setTorrents] = useState([])
  const [offline, setOffline] = useState(true)
  const timerID = useRef(-1)

  const updateTorrentList = torrs => {
    setTorrents(torrs)
    setOffline(false)
  }

  const resetTorrentList = () => {
    setTorrents([])
    setOffline(true)
  }

  useEffect(() => {
    timerID.current = setInterval(() => {
      getTorrentList(updateTorrentList, resetTorrentList)
    }, 1000)

    return () => {
      clearInterval(timerID.current)
    }
  }, [])

  return (
    <TorrentListWrapper>
      {offline ? (
        <Typography>Offline</Typography>
      ) : !torrents.length ? (
        <Typography>No torrents added</Typography>
      ) : (
        torrents && torrents.map(torrent => <Torrent key={torrent.hash} torrent={torrent} />)
      )}
    </TorrentListWrapper>
  )
}
