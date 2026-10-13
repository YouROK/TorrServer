import { useQuery, type UseQueryResult } from '@tanstack/react-query'

import { getSettings, SETTINGS_QUERY_KEY } from 'shared/api/settings'
import type { BTSets } from 'shared/api/types'

export function useSettingsQuery(options?: { enabled?: boolean }): UseQueryResult<BTSets, Error> {
  return useQuery({
    queryKey: SETTINGS_QUERY_KEY,
    queryFn: ({ signal }) => getSettings(signal),
    enabled: options?.enabled ?? true,
    staleTime: 30_000,
  })
}
