import styled from 'styled-components'
import { useEffect, useRef, useState } from 'react'
import { Typography } from '@material-ui/core'
import { torrentsHost } from 'utils/Hosts'
import TorrentCard from 'components/TorrentCard'
import axios from 'axios'

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

export default function TorrentList() {
  const [torrents, setTorrents] = useState([])
  const [offline, setOffline] = useState(true)
  const timerID = useRef(-1)

  useEffect(() => {
    timerID.current = setInterval(() => {
      // getting torrent list
      axios
        .post(torrentsHost(), { action: 'list' })
        .then(({ data }) => {
          // updating torrent list
          setTorrents(data)
          setOffline(false)
        })
        .catch(() => {
          // resetting torrent list
          setTorrents([])
          setOffline(true)
        })
    }, 1000)

    return () => clearInterval(timerID.current)
  }, [])

  return (
    <TorrentListWrapper>
      {offline ? (
        <Typography>Offline</Typography>
      ) : !torrents.length ? (
        <Typography>No torrents added</Typography>
      ) : (
        torrents.map(torrent => <TorrentCard key={torrent.hash} torrent={torrent} />)
      )}
    </TorrentListWrapper>
  )
}
