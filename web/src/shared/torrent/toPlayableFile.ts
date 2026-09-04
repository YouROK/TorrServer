import type { PlayableFile, TorrentFileStat } from 'shared/api/types'

/** Normalize API file_stats (mixed PascalCase/camelCase) into a PlayableFile. */
export const toPlayableFile = (file: TorrentFileStat): PlayableFile => ({
  id: file.id ?? file.Id ?? 0,
  path: file.path ?? file.Path ?? '',
  length: file.length ?? file.Length ?? 0,
})
