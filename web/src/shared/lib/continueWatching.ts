/**
 * Local "Continue Watching" shelf (browser localStorage only — not server viewed).
 *
 * Rules:
 * - Skip remembering until playback passes 15s (avoids noise from accidental opens).
 * - Cap at CONTINUE_LIMIT (12) newest entries.
 * - {@link listContinueWatching} can prune orphans when given the live torrent hash set.
 */

const CONTINUE_KEY = 'torrserver.continueWatching'
/** Max remembered items; older rows fall off the end of the list. */
const CONTINUE_LIMIT = 12

export interface ContinueWatchEntry {
  hash: string
  fileIndex: number
  title: string
  fileName: string
  timecode: number
  updatedAt: number
}

function readAll(): ContinueWatchEntry[] {
  try {
    const raw = localStorage.getItem(CONTINUE_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw) as ContinueWatchEntry[]
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

function writeAll(entries: ContinueWatchEntry[]) {
  try {
    localStorage.setItem(CONTINUE_KEY, JSON.stringify(entries.slice(0, CONTINUE_LIMIT)))
  } catch {
    // quota / private mode — ignore
  }
}

/** Upsert one resume row; no-op when `timecode < 15`. */
export function rememberContinueWatching(entry: Omit<ContinueWatchEntry, 'updatedAt'>): void {
  if (!entry.hash || entry.fileIndex == null || entry.timecode < 15) return
  const next: ContinueWatchEntry = { ...entry, updatedAt: Date.now() }
  const rest = readAll().filter(item => !(item.hash === next.hash && item.fileIndex === next.fileIndex))
  writeAll([next, ...rest])
}

/**
 * Newest-first list. When `knownHashes` is provided, drops entries for torrents
 * no longer in the library and persists the prune.
 */
export function listContinueWatching(knownHashes?: Set<string>): ContinueWatchEntry[] {
  const entries = readAll()
  if (!knownHashes) return entries
  const filtered = entries.filter(item => knownHashes.has(item.hash))
  if (filtered.length !== entries.length) writeAll(filtered)
  return filtered
}

/** Remove one file or every entry for a torrent hash. */
export function removeContinueWatching(hash: string, fileIndex?: number): void {
  writeAll(
    readAll().filter(item =>
      fileIndex == null ? item.hash !== hash : !(item.hash === hash && item.fileIndex === fileIndex),
    ),
  )
}
