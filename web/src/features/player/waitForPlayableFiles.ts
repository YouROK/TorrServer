import type { PlayableFile, TorrentFileStat } from 'shared/api/types'
import { getTorrent } from 'shared/api/torrents'
import { filesFromMetadata } from 'shared/torrent/fileMetadata'
import { isFilePlayable } from 'shared/torrent/playable'
import { toPlayableFile } from 'shared/torrent/toPlayableFile'
export interface WaitForPlayableFilesOptions {
  /** Max time to wait for live file_stats after waking the torrent. Default 20s. */
  timeoutMs?: number
  /** Poll interval while waiting. Default 500ms. */
  intervalMs?: number
  /** Persisted torrent.data JSON — used immediately when live file_stats are still empty. */
  metadataData?: string
  signal?: AbortSignal
}

export interface WaitForPlayableFilesResult {
  playable: PlayableFile[]
  allFiles: TorrentFileStat[]
}

const sleep = (ms: number, signal?: AbortSignal) =>
  new Promise<void>((resolve, reject) => {
    if (signal?.aborted) {
      reject(new DOMException('Aborted', 'AbortError'))
      return
    }
    const timer = setTimeout(resolve, ms)
    signal?.addEventListener(
      'abort',
      () => {
        clearTimeout(timer)
        reject(new DOMException('Aborted', 'AbortError'))
      },
      { once: true },
    )
  })

function playableFromStats(stats: TorrentFileStat[] | undefined): PlayableFile[] {
  return (stats || []).map(toPlayableFile).filter(file => isFilePlayable(file.path))
}

function playableFromData(data?: string): { playable: PlayableFile[]; allFiles: TorrentFileStat[] } {
  const playable = filesFromMetadata(data).filter(file => isFilePlayable(file.path))
  if (!playable.length) return { playable: [], allFiles: [] }
  // Reconstruct a TorrentFileStat-shaped list for caption lookup / ensureAllTorrentFiles.
  const allFiles: TorrentFileStat[] = playable.map(file => ({
    id: file.id,
    path: file.path,
    length: file.length,
  }))
  return { playable, allFiles }
}

/**
 * Wakes an inactive (IN_DB) torrent via getTorrent, then waits until playable
 * file_stats appear — or falls back to persisted metadata when available.
 */
export async function waitForPlayableFiles(
  hash: string,
  options: WaitForPlayableFilesOptions = {},
): Promise<WaitForPlayableFilesResult> {
  const timeoutMs = options.timeoutMs ?? 20_000
  const intervalMs = options.intervalMs ?? 500
  const signal = options.signal
  const started = Date.now()

  const tryDetail = async (): Promise<WaitForPlayableFilesResult | null> => {
    const detail = await getTorrent(hash)
    const live = playableFromStats(detail.file_stats)
    if (live.length) {
      return { playable: live, allFiles: detail.file_stats || [] }
    }
    const fromDetailData = playableFromData(detail.data)
    if (fromDetailData.playable.length) return fromDetailData
    return null
  }

  // Immediate metadata fallback (card/list often already has torrent.data).
  const fromArg = playableFromData(options.metadataData)
  if (fromArg.playable.length) {
    // Still wake the torrent so playback/stream can proceed.
    try {
      const woken = await tryDetail()
      if (woken?.playable.length) return woken
    } catch {
      // Fall through to metadata-only if get fails briefly.
    }
    return fromArg
  }

  let lastError: unknown
  while (Date.now() - started < timeoutMs) {
    signal?.throwIfAborted()
    try {
      const result = await tryDetail()
      if (result) return result
    } catch (err) {
      lastError = err
      if (signal?.aborted) throw err
    }
    const remaining = timeoutMs - (Date.now() - started)
    if (remaining <= 0) break
    await sleep(Math.min(intervalMs, remaining), signal)
  }

  if (lastError && !(lastError instanceof DOMException && lastError.name === 'AbortError')) {
    // Prefer empty result so caller can toast NoPlayableFiles; network errors bubble only if never got a response.
  }

  return { playable: [], allFiles: [] }
}
