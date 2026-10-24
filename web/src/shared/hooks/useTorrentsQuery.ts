import { useQuery, type UseQueryResult } from '@tanstack/react-query'

import type { TorrentStat } from 'shared/api/types'
import { getTorrents, TORRENTS_QUERY_KEY } from 'shared/api/torrents'

import { torrentsRefetchIntervalMs } from './pollInterval'

/** Single owner for the torrents list poll — reuse everywhere (Shell, TorrentsPage, MultiAdd). */
export function useTorrentsQuery(options?: { enabled?: boolean }): UseQueryResult<TorrentStat[], Error> {
  return useQuery({
    queryKey: TORRENTS_QUERY_KEY,
    queryFn: ({ signal }) => getTorrents(signal),
    retry: 1,
    enabled: options?.enabled ?? true,
    refetchInterval: query => torrentsRefetchIntervalMs(document.hidden, query.state.data),
  })
}
