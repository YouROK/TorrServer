import { useState } from 'react'
import { Typography } from '@material-ui/core'
import TorrentCard from 'components/TorrentCard'
import CircularProgress from '@material-ui/core/CircularProgress'
import { TorrentListWrapper, CenteredGrid } from 'components/App/style'
import { useTranslation } from 'react-i18next'
import { useQuery } from 'react-query'
import { getTorrents } from 'utils/Utils'

export default function TorrentList() {
  const { t } = useTranslation()
  const [isOffline, setIsOffline] = useState(false)
  const { data: torrents, isLoading } = useQuery('torrents', getTorrents, {
    retry: 1,
    refetchInterval: 1000,
    onError: () => setIsOffline(true),
    onSuccess: () => setIsOffline(false),
  })

  if (isLoading || isOffline || !torrents.length) {
    return (
      <CenteredGrid>
        {isOffline ? (
          <Typography>{t('Offline')}</Typography>
        ) : isLoading ? (
          <CircularProgress />
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
