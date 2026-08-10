import type { PlayableFile, TorrentFileStat } from 'shared/api/types'
import { streamHost } from 'shared/api/hosts'
import { toPlayableFile } from 'shared/torrent/toPlayableFile'

const fileBaseName = (path: string): string => path.split('\\').pop()?.split('/').pop() || path

const fileStreamUrl = (hash: string, file: Pick<PlayableFile, 'id' | 'path'>): string =>
  `${streamHost()}/${encodeURIComponent(fileBaseName(file.path))}?link=${hash}&index=${file.id}&play`

/**
 * Sidecar caption URL: same basename as the video plus `.srt` / `.vtt` elsewhere in the torrent.
 * Empty string when none found.
 */
export const findCaptionSrc = (file: PlayableFile, allFiles: TorrentFileStat[], hash: string): string => {
  const base = file.path.replace(/\.[^/.]+$/, '')
  const caption = allFiles.find(candidate => {
    const path = candidate.path ?? candidate.Path ?? ''
    const id = candidate.id ?? candidate.Id ?? -1
    return id !== file.id && path.startsWith(base) && /\.(srt|vtt)$/i.test(path)
  })
  if (!caption) return ''
  const captionFile = toPlayableFile(caption)
  return fileStreamUrl(hash, captionFile)
}
