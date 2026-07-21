import { useQuery, type UseQueryResult } from '@tanstack/react-query'

import type { TorrentStat } from 'shared/api/types'
import { getTorrents, TORRENTS_QUERY_KEY } from 'shared/api/torrents'

/** Single owner for the torrents list poll — reuse everywhere (Shell, TorrentsPage, MultiAdd). */
export function useTorrentsQuery(options?: { enabled?: boolean }): UseQueryResult<TorrentStat[], Error> {
  return useQuery({
    queryKey: TORRENTS_QUERY_KEY,
    queryFn: getTorrents,
    retry: 1,
    enabled: options?.enabled ?? true,
    refetchInterval: () => {
      if (document.hidden) return 10_000
      // Details/settings/etc. open — ease off the list poll to cut server + battery load.
      if (document.body.dataset.modalOpen === 'true') return 5_000
      return 1000
    },
  })
}
