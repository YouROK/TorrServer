import type { TorrentStat, ViewedFileEntry } from 'shared/api/types'
import type { ContinueWatchEntry } from 'shared/lib/continueWatching'
import { listContinueWatching } from 'shared/lib/continueWatching'

/**
 * Build the Continue Watching shelf.
 *
 * Server viewed rows win per torrent hash (cross-device). LocalStorage fills
 * gaps for hashes with no server progress — never a full mode switch, so clearing
 * the last server chip cannot resurrect unrelated/local duplicates.
 */
export function buildContinueWatchShelf(
  torrents: TorrentStat[] | undefined,
  serverViewed: ViewedFileEntry[] | undefined,
  localTick?: number,
): ContinueWatchEntry[] {
  void localTick
  const byHash = new Map((torrents || []).map(tr => [tr.hash, tr]))
  const hashes = new Set(byHash.keys())

  const bestServer = new Map<string, ViewedFileEntry>()
  for (const row of serverViewed || []) {
    const hash = row.hash
    if (!hash || !hashes.has(hash)) continue
    const prev = bestServer.get(hash)
    if (!prev || (row.timecode ?? 0) >= (prev.timecode ?? 0)) bestServer.set(hash, row)
  }

  const serverRows: ContinueWatchEntry[] = []
  for (const [hash, row] of bestServer) {
    const torrent = byHash.get(hash)
    if (!torrent) continue
    const file = torrent.file_stats?.find(f => (f.id ?? f.Id) === row.file_index)
    const fileName = file?.path || file?.Path || ''
    serverRows.push({
      hash,
      fileIndex: row.file_index,
      title: torrent.title || torrent.name || hash,
      fileName: fileName.split(/[/\\]/).pop() || fileName,
      timecode: row.timecode ?? 0,
      updatedAt: Date.now(),
    })
  }

  const localOnly = listContinueWatching(hashes)
    .filter(entry => !bestServer.has(entry.hash))
    .slice(0, 6)

  return [...serverRows, ...localOnly].slice(0, 6)
}
