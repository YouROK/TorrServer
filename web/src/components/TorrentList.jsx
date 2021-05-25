import styled from 'styled-components'
import { useEffect, useRef, useState } from 'react'
import { Typography } from '@material-ui/core'
import { torrentsHost } from 'utils/Hosts'

import Torrent from './Torrent'

const TorrentListWrapper = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 350px));
  gap: 30px;
`

export default function TorrentList() {
  const [torrents, setTorrents] = useState([])
  const [offline, setOffline] = useState(true)
  const timerID = useRef(-1)

  useEffect(() => {
    timerID.current = setInterval(() => {
      getTorrentList(torrs => {
        if (torrs) setOffline(false)
        else setOffline(true)
        setTorrents(torrs)
      })
    }, 1000)

    return () => {
      clearInterval(timerID.current)
    }
  }, [])

  return (
    <TorrentListWrapper>
      {offline ? (
        <Typography>Offline</Typography>
      ) : (
        torrents && torrents.map(torrent => <Torrent key={torrent.hash} torrent={torrent} />)
      )}
    </TorrentListWrapper>
  )
}

function getTorrentList(callback) {
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
}
