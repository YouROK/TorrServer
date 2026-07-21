import axios from 'axios'

import type { ViewedFileEntry } from 'shared/api/types'
import { viewedHost } from 'shared/api/hosts'

export const VIEWED_QUERY_KEY = (hash: string) => ['viewed', hash] as const

/** Full viewed rows including resume timecodes (server field `timecode`). */
export const listViewedEntries = async (hash: string): Promise<ViewedFileEntry[]> => {
  const { data } = await axios.post<ViewedFileEntry[] | null>(viewedHost(), { action: 'list', hash })
  return Array.isArray(data) ? data : []
}

/**
 * Cross-torrent viewed history. Empty hash asks the server for every Viewed key
 * (`ListViewed("")` on the backend).
 */
export const listAllViewedEntries = async (): Promise<ViewedFileEntry[]> => listViewedEntries('')

export const ALL_VIEWED_QUERY_KEY = ['viewed', 'all'] as const

/** Sorted file indexes marked viewed — used by details badges / playlist-from-latest. */
export const listViewedFiles = async (hash: string): Promise<number[] | undefined> => {
  const entries = await listViewedEntries(hash)
  if (entries.length === 0) return undefined
  return entries.map(itm => itm.file_index).sort((a, b) => a - b)
}

/** Clear every viewed mark for a torrent (`file_index: -1` is the server sentinel). */
export const clearViewedFiles = async (hash: string): Promise<void> => {
  await axios.post(viewedHost(), { action: 'rem', hash, file_index: -1 })
}

/** Clear viewed mark for a single file (keeps other files' progress). */
export const remViewedFile = async (hash: string, fileIndex: number): Promise<void> => {
  await axios.post(viewedHost(), { action: 'rem', hash, file_index: fileIndex })
}

/** Mark viewed and optionally persist resume `timecode` (seconds). */
export const setViewedFile = async (hash: string, fileIndex: number, timecode = 0): Promise<void> => {
  await axios.post(viewedHost(), { action: 'set', hash, file_index: fileIndex, timecode })
}
