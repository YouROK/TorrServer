import axios from 'axios'
import { useQuery } from 'react-query'
import { torrentsHost } from 'utils/Hosts'

export const OFFLINE_QUOTA_KEY = 'offlineQuota'

// State of the offline storage: { used, available, max_size, free, limited_by, rotation, ... }.
// available = -1 means that no limit is known. The values change slowly (a download grows the
// used part only by whole torrents, the free space by what is written), so a slow poll is enough.
export const useOfflineQuota = enabled =>
  useQuery(OFFLINE_QUOTA_KEY, async () => (await axios.post(torrentsHost(), { action: 'offline_quota' })).data, {
    enabled,
    refetchInterval: 10000,
    refetchOnWindowFocus: false,
    retry: false,
  }).data
