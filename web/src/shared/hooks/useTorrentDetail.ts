import { useQuery, useQueryClient, type UseQueryResult } from '@tanstack/react-query'

import { getTorrent, TORRENTS_QUERY_KEY, upsertTorrentsInList } from 'shared/api/torrents'
import type { TorrentStat } from 'shared/api/types'

import { torrentDetailRefetchIntervalMs } from './pollInterval'

/**
 * Live torrent detail poll for the details sheet.
 * Uses list-row `initial` for instant paint; GETTING_INFO polls at 500ms and mirrors
 * into the library cache. PRELOAD / idle use a steadier cadence.
 */
export function useTorrentDetail(hash: string | undefined, initial?: TorrentStat): UseQueryResult<TorrentStat, Error> {
  const queryClient = useQueryClient()

  return useQuery({
    queryKey: ['torrent', hash],
    queryFn: async ({ signal }) => {
      const data = await getTorrent(hash!, signal)
      // Keep library cards in lockstep while metadata resolves.
      if (queryClient.getQueryData(TORRENTS_QUERY_KEY)) {
        upsertTorrentsInList(queryClient, data)
      }
      return data
    },
    enabled: Boolean(hash),
    initialData: initial,
    refetchInterval: query => torrentDetailRefetchIntervalMs(document.hidden, query.state.data?.stat),
    retry: 1,
  })
}
