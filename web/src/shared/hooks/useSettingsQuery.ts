import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
  type UseQueryResult,
} from '@tanstack/react-query'

import { getSettings, setSettings, SETTINGS_QUERY_KEY } from 'shared/api/settings'
import type { BTSets } from 'shared/api/types'
import { notifySettingsChanged } from 'shared/lib/settingsEvents'

export function useSettingsQuery(options?: { enabled?: boolean }): UseQueryResult<BTSets, Error> {
  return useQuery({
    queryKey: SETTINGS_QUERY_KEY,
    queryFn: ({ signal }) => getSettings(signal),
    enabled: options?.enabled ?? true,
    staleTime: 30_000,
  })
}

/** Persists BTSets and broadcasts {@link notifySettingsChanged} so non-query listeners refresh. */
export function useSaveSettingsMutation(): UseMutationResult<void, Error, BTSets> {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (sets: BTSets) => setSettings(sets),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: SETTINGS_QUERY_KEY })
      notifySettingsChanged()
    },
  })
}
