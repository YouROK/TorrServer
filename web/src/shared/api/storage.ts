import axios from 'axios'

import { storageSettingsHost } from 'shared/api/hosts'

export type StorageBackend = 'json' | 'bbolt'

export interface StorageSettings {
  settings: StorageBackend
  viewed: StorageBackend
  viewedCount?: number
}

/**
 * Default persistence backends when `/storage/settings` has never been saved:
 * editable JSON for BTSets, robust BBolt for viewed/resume history.
 */
export const defaultStorageSettings = (): StorageSettings => ({
  settings: 'json',
  viewed: 'bbolt',
})

/** Coerces unknown server values to the two known backends (safe defaults on garbage). */
export const getStorageSettings = async (signal?: AbortSignal): Promise<StorageSettings> => {
  const { data } = await axios.get<Partial<StorageSettings> & { viewedCount?: number }>(storageSettingsHost(), {
    signal,
  })
  return {
    settings: data.settings === 'bbolt' ? 'bbolt' : 'json',
    viewed: data.viewed === 'json' ? 'json' : 'bbolt',
    viewedCount: typeof data.viewedCount === 'number' ? data.viewedCount : undefined,
  }
}

export const setStorageSettings = async (prefs: StorageSettings): Promise<void> => {
  const { data } = await axios.post<{ status?: string; error?: string }>(storageSettingsHost(), {
    settings: prefs.settings,
    viewed: prefs.viewed,
  })
  if (data?.error) throw new Error(data.error)
  if (data?.status && data.status !== 'ok') throw new Error(data.status)
}
