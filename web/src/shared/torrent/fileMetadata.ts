import type { PlayableFile, TorrentFileStat } from 'shared/api/types'

interface PersistedFileMetadata {
  TorrServer?: { Files?: TorrentFileStat[] }
}

/**
 * Saved (`IN_DB`) torrents have no live `Torrent` object, so the backend's `file_stats` list is
 * always empty for them — the file list from the last time the torrent was loaded is instead
 * persisted verbatim as JSON in `TorrentStat.data` (`server/torr/dbwrapper.go` `AddTorrentDB`).
 * This mirrors that shape so the UI can still resolve/play a saved torrent's files up front.
 */
export function filesFromMetadata(data?: string): PlayableFile[] {
  if (!data) return []
  try {
    const parsed = JSON.parse(data) as PersistedFileMetadata
    const files = parsed.TorrServer?.Files || []
    return files.map(file => ({
      id: file.id ?? file.Id ?? 0,
      path: file.path ?? file.Path ?? '',
      length: file.length ?? file.Length ?? 0,
    }))
  } catch {
    return []
  }
}
