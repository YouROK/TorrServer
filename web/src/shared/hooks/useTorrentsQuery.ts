import { useQuery, type UseQueryResult } from '@tanstack/react-query'

import type { TorrentStat } from 'shared/api/types'
import { getTorrents, TORRENTS_QUERY_KEY } from 'shared/api/torrents'
import { GETTING_INFO } from 'shared/torrent/states'

const isGettingInfo = (list: TorrentStat[] | undefined) => Boolean(list?.some(torrent => torrent.stat === GETTING_INFO))

/** Single owner for the torrents list poll — reuse everywhere (Shell, TorrentsPage, MultiAdd). */
export function useTorrentsQuery(options?: { enabled?: boolean }): UseQueryResult<TorrentStat[], Error> {
  return useQuery({
    queryKey: TORRENTS_QUERY_KEY,
    queryFn: getTorrents,
    retry: 1,
    enabled: options?.enabled ?? true,
    refetchInterval: query => {
      if (document.hidden) return 10_000
      // Only metadata resolve needs a fast list poll; PRELOAD stays on the steady cadence.
      if (isGettingInfo(query.state.data)) return 500
      return 1000
    },
  })
}
