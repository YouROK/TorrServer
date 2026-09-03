import axios from 'axios'
import type { QueryClient } from '@tanstack/react-query'

import type { TorrentStat } from 'shared/api/types'
import { torrentUploadHost, torrentsHost } from 'shared/api/hosts'
import { IN_DB } from 'shared/torrent/states'

export const TORRENTS_QUERY_KEY = ['torrents'] as const

const torrentApiError = (err: unknown, fallback: string): Error & { response?: { status?: number } } => {
  let message = fallback
  let status: number | undefined
  if (axios.isAxiosError(err)) {
    status = err.response?.status
    const data = err.response?.data as { error?: string } | undefined
    if (data?.error) message = String(data.error)
    else if (err.message) message = err.message
  } else if (err instanceof Error && err.message) {
    message = err.message
  }
  const error = new Error(message) as Error & { response?: { status?: number } }
  if (status != null) error.response = { status }
  return error
}

/** Drop keeps the DB row — never remove the card; flip to idle so the grid doesn't vanish→GSAP-reappear. */
export const markTorrentsDroppedInList = (queryClient: QueryClient, hashes: string | string[]): void => {
  const list = Array.isArray(hashes) ? hashes : [hashes]
  const drop = new Set(list.map(hash => hash.toLowerCase()))
  queryClient.setQueryData<TorrentStat[]>(TORRENTS_QUERY_KEY, prev =>
    prev?.map(item => {
      if (!item.hash || !drop.has(item.hash.toLowerCase())) return item
      return {
        ...item,
        stat: IN_DB,
        download_speed: 0,
        upload_speed: 0,
        active_peers: 0,
        total_peers: 0,
        connected_seeders: 0,
        pending_peers: 0,
        half_open_peers: 0,
      }
    }),
  )
  for (const hash of list) {
    queryClient.setQueryData<TorrentStat>(['torrent', hash], prev =>
      prev
        ? {
            ...prev,
            stat: IN_DB,
            download_speed: 0,
            upload_speed: 0,
            active_peers: 0,
          }
        : prev,
    )
  }
}

/** Merge torrent rows into the list cache (new hashes prepended) for instant UI paint. */
export const upsertTorrentsInList = (queryClient: QueryClient, torrents: TorrentStat | TorrentStat[]): void => {
  const items = Array.isArray(torrents) ? torrents : [torrents]
  if (!items.length) return
  queryClient.setQueryData<TorrentStat[]>(TORRENTS_QUERY_KEY, prev => {
    const list = prev ? [...prev] : []
    for (const item of items) {
      if (!item?.hash) continue
      const idx = list.findIndex(row => row.hash?.toLowerCase() === item.hash.toLowerCase())
      if (idx >= 0) list[idx] = { ...list[idx], ...item }
      else list.unshift(item)
    }
    return list
  })
}

/**
 * Paint from known rows (add/upload response), then reconcile in the background.
 * Does not await the list round-trip — that was the main post-add lag.
 */
export const refreshTorrentsList = async (
  queryClient: QueryClient,
  options?: { torrents?: TorrentStat | TorrentStat[] },
): Promise<void> => {
  if (options?.torrents) upsertTorrentsInList(queryClient, options.torrents)
  // Fire-and-forget: UI already has upserted rows; don't block dialog close on list GET.
  void queryClient.invalidateQueries({ queryKey: TORRENTS_QUERY_KEY })
}

/** List torrents; attaches HTTP `status` on the thrown Error for auth/offline UI. */
export const getTorrents = async (): Promise<TorrentStat[]> => {
  try {
    const { data } = await axios.post<TorrentStat[]>(torrentsHost(), { action: 'list' })
    return data
  } catch (err) {
    throw torrentApiError(err, 'Failed to load torrents')
  }
}

export const getTorrent = async (hash: string): Promise<TorrentStat> => {
  try {
    const { data } = await axios.post<TorrentStat>(torrentsHost(), { action: 'get', hash })
    return data
  } catch (err) {
    throw torrentApiError(err, 'Failed to get torrent')
  }
}

export interface AddTorrentInput {
  link: string
  title?: string
  category?: string
  poster?: string
  save_to_db?: boolean
}

export const addTorrent = async (input: AddTorrentInput): Promise<TorrentStat> => {
  try {
    const { data } = await axios.post<TorrentStat>(torrentsHost(), {
      action: 'add',
      link: input.link,
      title: input.title || undefined,
      category: input.category || undefined,
      poster: input.poster ?? '',
      save_to_db: input.save_to_db ?? true,
    })
    return data
  } catch (err) {
    throw torrentApiError(err, 'Failed to add torrent')
  }
}

export interface UpdateTorrentInput {
  hash: string
  title?: string
  category?: string
  poster?: string
}

export const updateTorrent = async (input: UpdateTorrentInput): Promise<void> => {
  try {
    await axios.post(torrentsHost(), {
      action: 'set',
      hash: input.hash,
      title: input.title || undefined,
      category: input.category || undefined,
      poster: input.poster ?? '',
    })
  } catch (err) {
    throw torrentApiError(err, 'Failed to update torrent')
  }
}

/** Stop swarm activity but keep the torrent in the DB (`action: drop`). */
export const dropTorrent = async (hash: string): Promise<void> => {
  try {
    await axios.post(torrentsHost(), { action: 'drop', hash })
  } catch (err) {
    throw torrentApiError(err, 'Failed to drop torrent')
  }
}

/** Permanently delete the torrent from the DB (`action: rem`). */
export const removeTorrent = async (hash: string): Promise<void> => {
  try {
    await axios.post(torrentsHost(), { action: 'rem', hash })
  } catch (err) {
    throw torrentApiError(err, 'Failed to remove torrent')
  }
}

export const wipeTorrents = async (): Promise<void> => {
  try {
    await axios.post(torrentsHost(), { action: 'wipe' })
  } catch (err) {
    throw torrentApiError(err, 'Failed to wipe torrents')
  }
}

export interface UploadTorrentMeta {
  title?: string
  category?: string
  poster?: string
  save?: boolean
}

/** Multipart `.torrent` upload; form field `save` mirrors server "save to DB" flag. */
export const uploadTorrent = async (file: File, meta: UploadTorrentMeta = {}): Promise<TorrentStat | TorrentStat[]> => {
  const data = new FormData()
  data.append('save', meta.save === false ? 'false' : 'true')
  data.append('file', file)
  if (meta.title) data.append('title', meta.title)
  if (meta.category) data.append('category', meta.category)
  if (meta.poster) data.append('poster', meta.poster)
  try {
    const { data: status } = await axios.post<TorrentStat | TorrentStat[]>(torrentUploadHost(), data)
    return status
  } catch (err) {
    throw torrentApiError(err, 'Failed to upload torrent')
  }
}
