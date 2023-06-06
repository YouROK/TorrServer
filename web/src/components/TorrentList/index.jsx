import TorrentCard from 'components/TorrentCard'
import CircularProgress from '@material-ui/core/CircularProgress'
import { TorrentListWrapper, CenteredGrid } from 'components/App/style'

import NoServerConnection from './NoServerConnection'
import AddFirstTorrent from './AddFirstTorrent'

export default function TorrentList({ isOffline, isLoading, sortABC, torrents }) {
  if (isLoading || isOffline || !torrents.length) {
    return (
      <CenteredGrid>
        {isOffline ? (
          <NoServerConnection />
        ) : isLoading ? (
          <CircularProgress color='secondary' />
        ) : (
          !torrents.length && <AddFirstTorrent />
        )}
      </CenteredGrid>
    )
  }

  return sortABC ? (
    <TorrentListWrapper>
      {torrents
        .sort((a, b) => a.title > b.title)
        .map(torrent => (
          <TorrentCard key={torrent.hash} torrent={torrent} />
        ))}
    </TorrentListWrapper>
  ) : (
    <TorrentListWrapper>
      {torrents.map(torrent => (
        <TorrentCard key={torrent.hash} torrent={torrent} />
      ))}
    </TorrentListWrapper>
  )
}
