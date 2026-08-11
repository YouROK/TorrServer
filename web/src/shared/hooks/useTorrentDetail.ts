import { useQuery, useQueryClient, type UseQueryResult } from '@tanstack/react-query'

import { getTorrent, TORRENTS_QUERY_KEY, upsertTorrentsInList } from 'shared/api/torrents'
import type { TorrentStat } from 'shared/api/types'
import { GETTING_INFO, PRELOAD } from 'shared/torrent/states'

/**
 * Live torrent detail poll for the details sheet.
 * Uses list-row `initial` for instant paint; GETTING_INFO polls at 500ms and mirrors
 * into the library cache. PRELOAD / idle use a steadier cadence.
 */
export function useTorrentDetail(hash: string | undefined, initial?: TorrentStat): UseQueryResult<TorrentStat, Error> {
  const queryClient = useQueryClient()

  return useQuery({
    queryKey: ['torrent', hash],
    queryFn: async () => {
      const data = await getTorrent(hash!)
      // Keep library cards in lockstep while metadata resolves.
      if (queryClient.getQueryData(TORRENTS_QUERY_KEY)) {
        upsertTorrentsInList(queryClient, data)
      }
      return data
    },
    enabled: Boolean(hash),
    initialData: initial,
    refetchInterval: query => {
      if (document.hidden) return 5000
      const stat = query.state.data?.stat
      if (stat === GETTING_INFO) return 500
      if (stat === PRELOAD) return 1000
      return 2000
    },
    retry: 1,
  })
}
