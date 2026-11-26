import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import TorrentCard from 'components/TorrentCard'
import CircularProgress from '@material-ui/core/CircularProgress'
import TextField from '@material-ui/core/TextField'
import SearchIcon from '@material-ui/icons/Search'
import { TorrentListWrapper, CenteredGrid } from 'components/App/style'

import NoServerConnection from './NoServerConnection'
import AddFirstTorrent from './AddFirstTorrent'
import { SearchFieldWrapper } from './style'

export default function TorrentList({ isOffline, isLoading, sortABC, torrents, sortCategory }) {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')
  const sortedTorrents = useMemo(() => {
    if (!torrents) return []
    const query = search.trim().toLowerCase()
    const filtered = torrents.filter(
      torrent =>
        (sortCategory === 'all' || torrent.category === sortCategory) &&
        (!query || [torrent.title, torrent.name].some(name => (name || '').toLowerCase().includes(query))),
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
  }, [torrents, sortCategory, sortABC, search])

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
      <SearchFieldWrapper>
        <SearchIcon color='action' />
        <TextField
          id='torrent-search'
          type='search'
          label={t('Search')}
          value={search}
          onChange={event => setSearch(event.target.value)}
          variant='outlined'
          size='small'
          fullWidth
        />
      </SearchFieldWrapper>
      {sortedTorrents.map(torrent => (
        <TorrentCard key={torrent.hash} torrent={torrent} />
      ))}
    </TorrentListWrapper>
  )
}
