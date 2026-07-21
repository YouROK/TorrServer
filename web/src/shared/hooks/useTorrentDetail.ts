import { useQuery, type UseQueryResult } from '@tanstack/react-query'

import { getTorrent } from 'shared/api/torrents'
import type { TorrentStat } from 'shared/api/types'

/**
 * Live torrent detail poll for the details sheet.
 * Uses list-row `initial` for instant paint; refetches every 2s (5s while the tab is hidden).
 */
export function useTorrentDetail(hash: string | undefined, initial?: TorrentStat): UseQueryResult<TorrentStat, Error> {
  return useQuery({
    queryKey: ['torrent', hash],
    queryFn: () => getTorrent(hash!),
    enabled: Boolean(hash),
    initialData: initial,
    refetchInterval: () => (document.hidden ? 5000 : 2000),
    retry: 1,
  })
}
