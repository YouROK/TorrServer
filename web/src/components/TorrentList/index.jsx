import { useMemo } from 'react'
import TorrentCard from 'components/TorrentCard'
import CircularProgress from '@material-ui/core/CircularProgress'
import { TorrentListWrapper, CenteredGrid } from 'components/App/style'

import NoServerConnection from './NoServerConnection'
import AddFirstTorrent from './AddFirstTorrent'

import OfflineToolbar from 'components/OfflineToolbar'
import { useOfflineFilter } from 'components/OfflineToolbar/useOfflineFilter'
import { filterTorrents } from 'utils/offline'

export default function TorrentList({ isOffline, isLoading, sortABC, torrents, sortCategory }) {
  // The summary and the status filter are built over ALL torrents (not only the selected category)
  const { summary, hasOffline, filter, setFilter } = useOfflineFilter(torrents)

  const sortedTorrents = useMemo(() => {
    if (!torrents) return []
    const filtered = filterTorrents(
      torrents.filter(torrent => sortCategory === 'all' || torrent.category === sortCategory),
      filter,
    )

    if (sortABC) {
      return [...filtered].sort((a, b) => (a.title || '').localeCompare(b.title || '') || a.hash.localeCompare(b.hash))
    }

    // Default: keep API order but stabilize by hash to prevent jumping
    return [...filtered].sort((a, b) => {
      const tsA = a.timestamp || 0
      const tsB = b.timestamp || 0
      if (tsA !== tsB) return tsB - tsA
      return a.hash.localeCompare(b.hash)
    })
  }, [torrents, sortCategory, sortABC, filter])

  if (isLoading || isOffline || !torrents?.length) {
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

  return (
    <TorrentListWrapper>
      {hasOffline && (
        <div style={{ gridColumn: '1 / -1' }}>
          <OfflineToolbar torrents={torrents} summary={summary} filter={filter} onFilterChange={setFilter} />
        </div>
      )}
      {sortedTorrents.map(torrent => (
        <TorrentCard key={torrent.hash} torrent={torrent} />
      ))}
    </TorrentListWrapper>
  )
}
