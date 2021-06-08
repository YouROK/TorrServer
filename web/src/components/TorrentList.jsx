import { useEffect, useRef, useState } from 'react'
import { Typography } from '@material-ui/core'
import { torrentsHost } from 'utils/Hosts'
import TorrentCard from 'components/TorrentCard'
import axios from 'axios'
import CircularProgress from '@material-ui/core/CircularProgress'
import { TorrentListWrapper, CenteredGrid } from 'App/style'
import { useTranslation } from 'react-i18next'

export default function TorrentList() {
  const { t } = useTranslation()
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

  if (isLoading || isOffline || !torrents.length) {
    return (
      <CenteredGrid>
        {isLoading ? (
          <CircularProgress />
        ) : isOffline ? (
          <Typography>{t('Offline')}</Typography>
        ) : (
          !torrents.length && <Typography>{t('NoTorrentsAdded')}</Typography>
        )}
      </CenteredGrid>
    )
  }

  return (
    <TorrentListWrapper>
      {torrents.map(torrent => (
        <TorrentCard key={torrent.hash} torrent={torrent} />
      ))}
    </TorrentListWrapper>
  )
}
