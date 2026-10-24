import axios from 'axios'

import type { TorrentCache } from 'shared/api/types'
import { cacheHost } from 'shared/api/hosts'

/** Piece-cache snapshot for the snake. `signal` aborts on unmount or hash switch. */
export const getCache = async (hash: string, signal?: AbortSignal): Promise<TorrentCache> => {
  const { data } = await axios.post<TorrentCache | null>(cacheHost(), { action: 'get', hash }, { signal })
  return data || {}
}
