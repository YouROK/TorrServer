import axios from 'axios'

import type { SearchResultItem, TorznabCapsResponse } from 'shared/api/types'
import { searchHost, torznabCapsHost, torznabSearchHost } from 'shared/api/hosts'

export type { SearchResultItem }

export type TorznabSearchOptions = {
  index?: number
  cat?: string
  offset?: number
  limit?: number
  signal?: AbortSignal
}

/** Torznab indexer search (`/torznab/search/`). */
export const searchTorznab = async (query: string, options?: TorznabSearchOptions): Promise<SearchResultItem[]> => {
  const params: Record<string, string | number> = { query }

  if (options?.index != null) params.index = options.index
  if (options?.cat) params.cat = options.cat
  if (options?.offset != null && options.offset > 0) params.offset = options.offset
  if (options?.limit != null && options.limit > 0) params.limit = options.limit

  const { data } = await axios.get<SearchResultItem[]>(torznabSearchHost(), {
    params,
    signal: options?.signal,
  })
  return Array.isArray(data) ? data : []
}

export const fetchTorznabCaps = async (index: number, signal?: AbortSignal): Promise<TorznabCapsResponse> => {
  const { data } = await axios.get<TorznabCapsResponse>(torznabCapsHost(), { params: { index }, signal })
  return data
}

/** Built-in RuTor mirror search (`/search/`) — separate from Torznab indexers. */
export const searchRutor = async (query: string, signal?: AbortSignal): Promise<SearchResultItem[]> => {
  const { data } = await axios.get<SearchResultItem[]>(searchHost(), { params: { query }, signal })
  return Array.isArray(data) ? data : []
}
