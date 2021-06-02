import styled from 'styled-components'
import { useEffect, useRef, useState } from 'react'
import { Typography } from '@material-ui/core'
import { torrentsHost } from 'utils/Hosts'
import TorrentCard from 'components/TorrentCard'
import axios from 'axios'
import CircularProgress from '@material-ui/core/CircularProgress'
import { TorrentListWrapper } from 'App/style'

const CenteredGrid = styled.div`
  height: 100%;
  display: grid;
  place-items: center;
`

export default function TorrentList() {
  const [torrents, setTorrents] = useState([])
  const [isLoading, setIsLoading] = useState(true)
  const [isOffline, setIsOffline] = useState(true)
  const timerID = useRef(-1)

  useEffect(() => {
    timerID.current = setInterval(() => {
      // getting torrent list
      axios
        .post(torrentsHost(), { action: 'list' })
        .then(({ data }) => {
          // updating torrent list
          setTorrents(data)
          setIsOffline(false)
        })
        .catch(() => {
          // resetting torrent list
          setTorrents([])
          setIsOffline(true)
        })
        .finally(() => setIsLoading(false))
    }, 1000)

    return () => clearInterval(timerID.current)
  }, [])

  return isLoading ? (
    <CenteredGrid>
      <CircularProgress />
    </CenteredGrid>
  ) : isOffline ? (
    <CenteredGrid>
      <Typography>Offline</Typography>
    </CenteredGrid>
  ) : !torrents.length ? (
    <CenteredGrid>
      <Typography>No torrents added</Typography>
    </CenteredGrid>
  ) : (
    <TorrentListWrapper>
      {torrents.map(torrent => (
        <TorrentCard key={torrent.hash} torrent={torrent} />
      ))}
    </TorrentListWrapper>
  )
}
