import { useEffect, useMemo, useState } from 'react'
import { FILTER_ALL, filterTorrents, getFilterCount, getTorrentsSummary } from 'utils/offline'

// Keeps the status filter of the torrent list and derives everything the toolbar needs.
// Usage in the list component:
//   const { summary, hasOffline, filter, setFilter, filteredTorrents } = useOfflineFilter(torrents)
export const useOfflineFilter = torrents => {
  const [filter, setFilter] = useState(FILTER_ALL)
  const summary = useMemo(() => getTorrentsSummary(torrents), [torrents])
  const hasOffline = summary.offline > 0

  // The selected group may become empty (last paused torrent resumed, offline copy removed, ...):
  // fall back to "all" instead of showing an empty list with no way to tell why.
  const filterIsEmpty = filter !== FILTER_ALL && (!hasOffline || getFilterCount(summary, filter) === 0)
  useEffect(() => {
    if (filterIsEmpty) setFilter(FILTER_ALL)
  }, [filterIsEmpty])

  const activeFilter = filterIsEmpty ? FILTER_ALL : filter
  const filteredTorrents = useMemo(() => filterTorrents(torrents, activeFilter), [torrents, activeFilter])

  return { summary, hasOffline, filter: activeFilter, setFilter, filteredTorrents }
}
